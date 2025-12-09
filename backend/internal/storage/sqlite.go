package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

// Storage defines the interface for database operations
type Storage interface {
	Start(ctx context.Context) error
	Stop() error

	// User operations
	CreateUser(ctx context.Context, username string, addresses []string) (*User, error)
	CreateUserWithPersona(ctx context.Context, username string, addresses []string, personaID int64) (*User, error)
	GetUser(ctx context.Context, username string) (*User, error)
	GetUsers(ctx context.Context) ([]*User, error)
	UpdateUserLastSynced(ctx context.Context, userID int64, lastSynced time.Time) error
	UpdateUserPersona(ctx context.Context, userID int64, personaID int64) error
	UpdateUserProfileImage(ctx context.Context, userID int64, profileImage string) error
	UpdateUserOfficialPnl(ctx context.Context, userID int64, pnl, volume float64) error

	// Address operations
	GetUserAddresses(ctx context.Context, userID int64) ([]*Address, error)

	// Position operations
	UpsertPosition(ctx context.Context, pos *Position) error
	GetUserPositions(ctx context.Context, userID int64) ([]*Position, error)
	DeleteUserPositions(ctx context.Context, userID int64) error

	// Trade operations
	InsertTrade(ctx context.Context, trade *Trade) error
	GetUserTrades(ctx context.Context, userID int64, limit, offset int) ([]*Trade, int, error)
	GetAllTrades(ctx context.Context, filters TradeFilters) ([]*TradeWithUsername, int, error)
	GetUserTradesChronological(ctx context.Context, userID int64) ([]*Trade, error)

	// PNL operations
	InsertPnlSnapshot(ctx context.Context, snapshot *PnlSnapshot) error
	GetUserPnlHistory(ctx context.Context, userID int64, start, end *time.Time) ([]*PnlSnapshot, error)
	DeleteUserPnlSnapshots(ctx context.Context, userID int64) error
	BulkInsertPnlSnapshots(ctx context.Context, snapshots []*PnlSnapshot) error

	// Aggregation operations
	GetUserStats(ctx context.Context, username string) (*UserStats, error)
	GetLeaderboard(ctx context.Context, sortBy, sortDirection string) ([]*UserStats, error)

	// Persona operations
	CreatePersona(ctx context.Context, slug, displayName string) (*Persona, error)
	CreatePersonaWithImage(ctx context.Context, slug, displayName, image string) (*Persona, error)
	GetPersona(ctx context.Context, slug string) (*Persona, error)
	GetPersonas(ctx context.Context) ([]*Persona, error)
	GetPersonaUsers(ctx context.Context, personaID int64) ([]*User, error)
	GetPersonaStats(ctx context.Context, slug string) (*PersonaStats, error)
	GetPersonaLeaderboard(ctx context.Context, sortBy, sortDirection string) ([]*PersonaStats, error)
	GetPersonaPositions(ctx context.Context, slug string) ([]*PositionWithUsername, error)
	GetPersonaTrades(ctx context.Context, slug string, limit, offset int) ([]*TradeWithUsername, int, error)
	GetUserPersonaInfo(ctx context.Context, userID int64) (*PersonaInfo, error)
	UpdatePersonaImage(ctx context.Context, personaID int64, image string) error

	// Results operations
	GetUserResults(ctx context.Context, userID int64, limit, offset int) ([]*Result, int, error)
	GetPersonaResults(ctx context.Context, slug string, limit, offset int) ([]*ResultWithUsername, int, error)
}

// storage is the SQLite implementation of Storage
type storage struct {
	db   *sql.DB
	path string
	log  logrus.FieldLogger
}

var _ Storage = (*storage)(nil)

// NewStorage creates a new Storage instance
func NewStorage(path string, log logrus.FieldLogger) Storage {
	return &storage{
		path: path,
		log:  log.WithField("package", "storage"),
	}
}

// Start initializes the database connection and runs migrations
func (s *storage) Start(ctx context.Context) error {
	s.log.Info("starting storage")

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", s.path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite works best with a single connection
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	s.db = db

	// Run migrations
	if err := runMigrations(ctx, s.db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	s.log.WithField("path", s.path).Info("storage started")
	return nil
}

// Stop closes the database connection
func (s *storage) Stop() error {
	s.log.Info("stopping storage")

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	s.log.Info("storage stopped")
	return nil
}

// CreateUser creates a new user with addresses
func (s *storage) CreateUser(ctx context.Context, username string, addresses []string) (*User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert user
	result, err := tx.ExecContext(ctx,
		"INSERT INTO users (username, created_at) VALUES (?, CURRENT_TIMESTAMP)",
		username,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}

	// Insert addresses
	for _, addr := range addresses {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO addresses (user_id, address) VALUES (?, ?)",
			userID, addr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert address: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetUser(ctx, username)
}

// GetUser retrieves a user by username
func (s *storage) GetUser(ctx context.Context, username string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, username, created_at, last_synced, profile_image, official_pnl, official_volume FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastSynced, &user.ProfileImage, &user.OfficialPnl, &user.OfficialVolume)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// GetUsers retrieves all users
func (s *storage) GetUsers(ctx context.Context) ([]*User, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, username, created_at, last_synced, profile_image, official_pnl, official_volume FROM users ORDER BY username",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastSynced, &user.ProfileImage, &user.OfficialPnl, &user.OfficialVolume); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// UpdateUserLastSynced updates the last synced timestamp for a user
func (s *storage) UpdateUserLastSynced(ctx context.Context, userID int64, lastSynced time.Time) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET last_synced = ? WHERE id = ?",
		lastSynced, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user last synced: %w", err)
	}
	return nil
}

// GetUserAddresses retrieves all addresses for a user
func (s *storage) GetUserAddresses(ctx context.Context, userID int64) ([]*Address, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, user_id, address FROM addresses WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query addresses: %w", err)
	}
	defer rows.Close()

	addresses := make([]*Address, 0)
	for rows.Next() {
		var addr Address
		if err := rows.Scan(&addr.ID, &addr.UserID, &addr.Address); err != nil {
			return nil, fmt.Errorf("failed to scan address: %w", err)
		}
		addresses = append(addresses, &addr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating addresses: %w", err)
	}

	return addresses, nil
}

// UpsertPosition inserts or updates a position
func (s *storage) UpsertPosition(ctx context.Context, pos *Position) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO positions (
			user_id, address, condition_id, asset, market_title, market_slug,
			outcome, size, avg_price, current_price, initial_value, current_value,
			unrealized_pnl, unrealized_pnl_percent, realized_pnl, end_date, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, address, condition_id, asset) DO UPDATE SET
			market_title = excluded.market_title,
			market_slug = excluded.market_slug,
			outcome = excluded.outcome,
			size = excluded.size,
			avg_price = excluded.avg_price,
			current_price = excluded.current_price,
			initial_value = excluded.initial_value,
			current_value = excluded.current_value,
			unrealized_pnl = excluded.unrealized_pnl,
			unrealized_pnl_percent = excluded.unrealized_pnl_percent,
			realized_pnl = excluded.realized_pnl,
			end_date = excluded.end_date,
			updated_at = CURRENT_TIMESTAMP
	`,
		pos.UserID, pos.Address, pos.ConditionID, pos.Asset, pos.MarketTitle, pos.MarketSlug,
		pos.Outcome, pos.Size, pos.AvgPrice, pos.CurrentPrice, pos.InitialValue, pos.CurrentValue,
		pos.UnrealizedPnl, pos.UnrealizedPnlPercent, pos.RealizedPnl, pos.EndDate,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert position: %w", err)
	}
	return nil
}

// GetUserPositions retrieves all positions for a user
func (s *storage) GetUserPositions(ctx context.Context, userID int64) ([]*Position, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, address, condition_id, asset, market_title, market_slug,
			outcome, size, avg_price, current_price, initial_value, current_value,
			unrealized_pnl, unrealized_pnl_percent, realized_pnl, end_date, updated_at
		FROM positions
		WHERE user_id = ?
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	positions := make([]*Position, 0)
	for rows.Next() {
		var pos Position
		if err := rows.Scan(
			&pos.ID, &pos.UserID, &pos.Address, &pos.ConditionID, &pos.Asset,
			&pos.MarketTitle, &pos.MarketSlug, &pos.Outcome, &pos.Size, &pos.AvgPrice,
			&pos.CurrentPrice, &pos.InitialValue, &pos.CurrentValue, &pos.UnrealizedPnl,
			&pos.UnrealizedPnlPercent, &pos.RealizedPnl, &pos.EndDate, &pos.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, &pos)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating positions: %w", err)
	}

	return positions, nil
}

// DeleteUserPositions deletes all positions for a user
func (s *storage) DeleteUserPositions(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM positions WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete positions: %w", err)
	}
	return nil
}

// InsertTrade inserts a new trade
func (s *storage) InsertTrade(ctx context.Context, trade *Trade) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO trades (
			user_id, address, trade_id, condition_id, market_title, market_slug,
			outcome, side, price, size, value, timestamp, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, condition_id, timestamp, side, size, price) DO NOTHING
	`,
		trade.UserID, trade.Address, trade.TradeID, trade.ConditionID, trade.MarketTitle,
		trade.MarketSlug, trade.Outcome, trade.Side, trade.Price, trade.Size, trade.Value,
		trade.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to insert trade: %w", err)
	}
	return nil
}

// GetUserTrades retrieves trades for a user with pagination
func (s *storage) GetUserTrades(ctx context.Context, userID int64, limit, offset int) ([]*Trade, int, error) {
	// Get total count
	var total int
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM trades WHERE user_id = ?",
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count trades: %w", err)
	}

	// Get trades with pagination
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, address, trade_id, condition_id, market_title, market_slug,
			outcome, side, price, size, value, timestamp, created_at
		FROM trades
		WHERE user_id = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query trades: %w", err)
	}
	defer rows.Close()

	trades := make([]*Trade, 0, limit)
	for rows.Next() {
		var trade Trade
		if err := rows.Scan(
			&trade.ID, &trade.UserID, &trade.Address, &trade.TradeID, &trade.ConditionID,
			&trade.MarketTitle, &trade.MarketSlug, &trade.Outcome, &trade.Side, &trade.Price,
			&trade.Size, &trade.Value, &trade.Timestamp, &trade.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, &trade)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating trades: %w", err)
	}

	return trades, total, nil
}

// GetAllTrades retrieves all trades across all users with filtering and pagination
func (s *storage) GetAllTrades(ctx context.Context, filters TradeFilters) ([]*TradeWithUsername, int, error) {
	// Build WHERE clause and args
	whereConditions := make([]string, 0)
	args := make([]any, 0)

	if filters.Username != nil {
		whereConditions = append(whereConditions, "u.username = ?")
		args = append(args, *filters.Username)
	}

	if filters.Side != nil {
		whereConditions = append(whereConditions, "t.side = ?")
		args = append(args, *filters.Side)
	}

	if filters.MinValue != nil {
		whereConditions = append(whereConditions, "t.value >= ?")
		args = append(args, *filters.MinValue)
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + fmt.Sprintf("%s", whereConditions[0])
		for i := 1; i < len(whereConditions); i++ {
			whereClause += " AND " + whereConditions[i]
		}
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM trades t
		JOIN users u ON t.user_id = u.id
		%s
	`, whereClause)

	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count trades: %w", err)
	}

	// Build ORDER BY clause
	sortColumn := "t.timestamp"
	switch filters.SortBy {
	case "value":
		sortColumn = "t.value"
	case "size":
		sortColumn = "t.size"
	case "timestamp":
		sortColumn = "t.timestamp"
	default:
		sortColumn = "t.timestamp"
	}

	sortOrder := "DESC"
	if filters.SortDirection == "asc" {
		sortOrder = "ASC"
	}

	orderByClause := fmt.Sprintf("ORDER BY %s %s", sortColumn, sortOrder)

	// Build full query
	query := fmt.Sprintf(`
		SELECT
			t.id, t.user_id, t.address, t.trade_id, t.condition_id, t.market_title,
			t.market_slug, t.outcome, t.side, t.price, t.size, t.value,
			t.timestamp, t.created_at, u.username
		FROM trades t
		JOIN users u ON t.user_id = u.id
		%s
		%s
		LIMIT ? OFFSET ?
	`, whereClause, orderByClause)

	// Append limit and offset to args
	queryArgs := append(args, filters.Limit, filters.Offset)

	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query trades: %w", err)
	}
	defer rows.Close()

	trades := make([]*TradeWithUsername, 0, filters.Limit)
	for rows.Next() {
		var trade TradeWithUsername
		if err := rows.Scan(
			&trade.ID, &trade.UserID, &trade.Address, &trade.TradeID, &trade.ConditionID,
			&trade.MarketTitle, &trade.MarketSlug, &trade.Outcome, &trade.Side, &trade.Price,
			&trade.Size, &trade.Value, &trade.Timestamp, &trade.CreatedAt, &trade.Username,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, &trade)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating trades: %w", err)
	}

	return trades, total, nil
}

// InsertPnlSnapshot inserts a PNL snapshot
func (s *storage) InsertPnlSnapshot(ctx context.Context, snapshot *PnlSnapshot) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO pnl_snapshots (user_id, timestamp, total_pnl, realized_pnl, unrealized_pnl)
		VALUES (?, ?, ?, ?, ?)
	`,
		snapshot.UserID, snapshot.Timestamp, snapshot.TotalPnl, snapshot.RealizedPnl, snapshot.UnrealizedPnl,
	)
	if err != nil {
		return fmt.Errorf("failed to insert pnl snapshot: %w", err)
	}
	return nil
}

// GetUserPnlHistory retrieves PNL history for a user
func (s *storage) GetUserPnlHistory(ctx context.Context, userID int64, start, end *time.Time) ([]*PnlSnapshot, error) {
	query := `
		SELECT id, user_id, timestamp, total_pnl, realized_pnl, unrealized_pnl
		FROM pnl_snapshots
		WHERE user_id = ?
	`
	args := []any{userID}

	if start != nil {
		query += " AND timestamp >= ?"
		args = append(args, start)
	}
	if end != nil {
		query += " AND timestamp <= ?"
		args = append(args, end)
	}

	query += " ORDER BY timestamp ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pnl history: %w", err)
	}
	defer rows.Close()

	snapshots := make([]*PnlSnapshot, 0)
	for rows.Next() {
		var snapshot PnlSnapshot
		if err := rows.Scan(
			&snapshot.ID, &snapshot.UserID, &snapshot.Timestamp,
			&snapshot.TotalPnl, &snapshot.RealizedPnl, &snapshot.UnrealizedPnl,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pnl snapshot: %w", err)
		}
		snapshots = append(snapshots, &snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pnl snapshots: %w", err)
	}

	return snapshots, nil
}

// GetUserStats retrieves aggregated statistics for a user
func (s *storage) GetUserStats(ctx context.Context, username string) (*UserStats, error) {
	user, err := s.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	addresses, err := s.GetUserAddresses(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user addresses: %w", err)
	}

	addressList := make([]string, len(addresses))
	for i, addr := range addresses {
		addressList[i] = addr.Address
	}

	stats := &UserStats{
		Username:     username,
		Addresses:    addressList,
		ProfileImage: user.ProfileImage,
		LastSynced:   user.LastSynced,
	}

	// Get position stats (only unrealized PnL from current open positions)
	var openPositions int
	var unrealizedPnl sql.NullFloat64
	err = s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) as open_positions,
			COALESCE(SUM(unrealized_pnl), 0) as unrealized_pnl
		FROM positions
		WHERE user_id = ?
	`, user.ID).Scan(&openPositions, &unrealizedPnl)
	if err != nil {
		return nil, fmt.Errorf("failed to get position stats: %w", err)
	}

	stats.OpenPositions = openPositions
	if unrealizedPnl.Valid {
		stats.UnrealizedPnl = unrealizedPnl.Float64
	}

	// Use official PnL from Polymarket if available (all-time accurate data)
	// Otherwise fall back to FIFO calculation from available trade history
	if user.OfficialPnl != nil {
		// Official PnL is the total (realized + unrealized)
		stats.TotalPnl = *user.OfficialPnl
		// Calculate realized as: total - current unrealized
		stats.RealizedPnl = stats.TotalPnl - stats.UnrealizedPnl
	} else {
		// Fall back to FIFO calculation from trade history
		realizedPnl, _, _, err := s.CalculateRealizedPnlFromTrades(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate realized pnl: %w", err)
		}
		stats.RealizedPnl = realizedPnl
		stats.TotalPnl = stats.RealizedPnl + stats.UnrealizedPnl
	}

	// Get trade stats
	var totalTrades int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM trades WHERE user_id = ?", user.ID).Scan(&totalTrades)
	if err != nil {
		return nil, fmt.Errorf("failed to count trades: %w", err)
	}
	stats.TotalTrades = totalTrades

	// Calculate win rate from FIFO (we still need this for win/loss tracking)
	_, wins, totalClosed, err := s.CalculateRealizedPnlFromTrades(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate win rate: %w", err)
	}
	if totalClosed > 0 {
		stats.WinRate = float64(wins) / float64(totalClosed)
	}

	return stats, nil
}

// GetLeaderboard retrieves leaderboard of all users
func (s *storage) GetLeaderboard(ctx context.Context, sortBy, sortDirection string) ([]*UserStats, error) {
	users, err := s.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	leaderboard := make([]*UserStats, 0, len(users))
	for _, user := range users {
		stats, err := s.GetUserStats(ctx, user.Username)
		if err != nil {
			s.log.WithError(err).WithField("username", user.Username).Error("failed to get user stats")
			continue
		}
		leaderboard = append(leaderboard, stats)
	}

	// Sort leaderboard
	// Note: In a production system, this should be done in SQL for better performance
	// For simplicity, we're doing it in Go here

	return leaderboard, nil
}

// GetUserTradesChronological retrieves all trades for a user sorted by timestamp ASC
func (s *storage) GetUserTradesChronological(ctx context.Context, userID int64) ([]*Trade, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, address, trade_id, condition_id, market_title, market_slug,
			outcome, side, price, size, value, timestamp, created_at
		FROM trades
		WHERE user_id = ?
		ORDER BY timestamp ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query trades: %w", err)
	}
	defer rows.Close()

	trades := make([]*Trade, 0)
	for rows.Next() {
		var trade Trade
		if err := rows.Scan(
			&trade.ID, &trade.UserID, &trade.Address, &trade.TradeID, &trade.ConditionID,
			&trade.MarketTitle, &trade.MarketSlug, &trade.Outcome, &trade.Side, &trade.Price,
			&trade.Size, &trade.Value, &trade.Timestamp, &trade.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, &trade)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trades: %w", err)
	}

	return trades, nil
}

// DeleteUserPnlSnapshots deletes all PNL snapshots for a user
func (s *storage) DeleteUserPnlSnapshots(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM pnl_snapshots WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete pnl snapshots: %w", err)
	}
	return nil
}

// BulkInsertPnlSnapshots inserts multiple PNL snapshots in a single transaction
func (s *storage) BulkInsertPnlSnapshots(ctx context.Context, snapshots []*PnlSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO pnl_snapshots (user_id, timestamp, total_pnl, realized_pnl, unrealized_pnl)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, snapshot := range snapshots {
		_, err := stmt.ExecContext(ctx,
			snapshot.UserID, snapshot.Timestamp, snapshot.TotalPnl, snapshot.RealizedPnl, snapshot.UnrealizedPnl,
		)
		if err != nil {
			return fmt.Errorf("failed to insert pnl snapshot: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateUserWithPersona creates a new user with addresses and associates with a persona
func (s *storage) CreateUserWithPersona(ctx context.Context, username string, addresses []string, personaID int64) (*User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert user with persona_id
	result, err := tx.ExecContext(ctx,
		"INSERT INTO users (username, created_at, persona_id) VALUES (?, CURRENT_TIMESTAMP, ?)",
		username, personaID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}

	// Insert addresses
	for _, addr := range addresses {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO addresses (user_id, address) VALUES (?, ?)",
			userID, addr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert address: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetUser(ctx, username)
}

// UpdateUserPersona updates a user's persona association
func (s *storage) UpdateUserPersona(ctx context.Context, userID int64, personaID int64) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET persona_id = ? WHERE id = ?",
		personaID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user persona: %w", err)
	}
	return nil
}

// CreatePersona creates a new persona
func (s *storage) CreatePersona(ctx context.Context, slug, displayName string) (*Persona, error) {
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO personas (slug, display_name, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
		slug, displayName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert persona: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get persona id: %w", err)
	}

	return &Persona{
		ID:          id,
		Slug:        slug,
		DisplayName: displayName,
	}, nil
}

// GetPersona retrieves a persona by slug
func (s *storage) GetPersona(ctx context.Context, slug string) (*Persona, error) {
	var persona Persona
	err := s.db.QueryRowContext(ctx,
		"SELECT id, slug, display_name, image, created_at FROM personas WHERE slug = ?",
		slug,
	).Scan(&persona.ID, &persona.Slug, &persona.DisplayName, &persona.Image, &persona.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("persona not found: %s", slug)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query persona: %w", err)
	}

	return &persona, nil
}

// GetPersonas retrieves all personas
func (s *storage) GetPersonas(ctx context.Context) ([]*Persona, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, slug, display_name, image, created_at FROM personas ORDER BY display_name",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query personas: %w", err)
	}
	defer rows.Close()

	personas := make([]*Persona, 0)
	for rows.Next() {
		var persona Persona
		if err := rows.Scan(&persona.ID, &persona.Slug, &persona.DisplayName, &persona.Image, &persona.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan persona: %w", err)
		}
		personas = append(personas, &persona)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating personas: %w", err)
	}

	return personas, nil
}

// GetPersonaUsers retrieves all users belonging to a persona
func (s *storage) GetPersonaUsers(ctx context.Context, personaID int64) ([]*User, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, username, created_at, last_synced, profile_image, official_pnl, official_volume FROM users WHERE persona_id = ? ORDER BY username",
		personaID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastSynced, &user.ProfileImage, &user.OfficialPnl, &user.OfficialVolume); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// GetPersonaStats retrieves aggregated statistics for a persona across all their users
func (s *storage) GetPersonaStats(ctx context.Context, slug string) (*PersonaStats, error) {
	persona, err := s.GetPersona(ctx, slug)
	if err != nil {
		return nil, err
	}

	users, err := s.GetPersonaUsers(ctx, persona.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get persona users: %w", err)
	}

	stats := &PersonaStats{
		Slug:        persona.Slug,
		DisplayName: persona.DisplayName,
		Image:       persona.Image,
		Usernames:   make([]string, 0, len(users)),
	}

	var totalWins, totalClosed int
	var hasOfficialPnl bool
	var totalOfficialPnl float64

	for _, user := range users {
		stats.Usernames = append(stats.Usernames, user.Username)

		// Get position stats for this user (only unrealized PnL)
		var openPositions int
		var unrealizedPnl sql.NullFloat64
		err = s.db.QueryRowContext(ctx, `
			SELECT
				COUNT(*) as open_positions,
				COALESCE(SUM(unrealized_pnl), 0) as unrealized_pnl
			FROM positions
			WHERE user_id = ?
		`, user.ID).Scan(&openPositions, &unrealizedPnl)
		if err != nil {
			return nil, fmt.Errorf("failed to get position stats for user %s: %w", user.Username, err)
		}

		stats.OpenPositions += openPositions
		if unrealizedPnl.Valid {
			stats.UnrealizedPnl += unrealizedPnl.Float64
		}

		// Calculate win rate data from FIFO for this user
		_, wins, closed, err := s.CalculateRealizedPnlFromTrades(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate win rate for user %s: %w", user.Username, err)
		}
		totalWins += wins
		totalClosed += closed

		// Use official PnL if available, otherwise fall back to FIFO calculation
		if user.OfficialPnl != nil {
			hasOfficialPnl = true
			totalOfficialPnl += *user.OfficialPnl
		}

		// Get trade count for this user
		var tradeCount int
		err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM trades WHERE user_id = ?", user.ID).Scan(&tradeCount)
		if err != nil {
			return nil, fmt.Errorf("failed to count trades for user %s: %w", user.Username, err)
		}
		stats.TotalTrades += tradeCount
	}

	// Use official PnL if any user has it
	if hasOfficialPnl {
		stats.TotalPnl = totalOfficialPnl
		stats.RealizedPnl = stats.TotalPnl - stats.UnrealizedPnl
	} else {
		stats.TotalPnl = stats.RealizedPnl + stats.UnrealizedPnl
	}

	if totalClosed > 0 {
		stats.WinRate = float64(totalWins) / float64(totalClosed)
	}

	return stats, nil
}

// GetPersonaLeaderboard retrieves leaderboard of all personas
func (s *storage) GetPersonaLeaderboard(ctx context.Context, sortBy, sortDirection string) ([]*PersonaStats, error) {
	personas, err := s.GetPersonas(ctx)
	if err != nil {
		return nil, err
	}

	leaderboard := make([]*PersonaStats, 0, len(personas))
	for _, persona := range personas {
		stats, err := s.GetPersonaStats(ctx, persona.Slug)
		if err != nil {
			s.log.WithError(err).WithField("slug", persona.Slug).Error("failed to get persona stats")
			continue
		}
		leaderboard = append(leaderboard, stats)
	}

	return leaderboard, nil
}

// GetPersonaPositions retrieves combined positions across all accounts for a persona
func (s *storage) GetPersonaPositions(ctx context.Context, slug string) ([]*PositionWithUsername, error) {
	persona, err := s.GetPersona(ctx, slug)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			p.id, p.user_id, p.address, p.condition_id, p.asset,
			p.market_title, p.market_slug, p.outcome,
			p.size, p.avg_price, p.current_price,
			p.initial_value, p.current_value,
			p.unrealized_pnl, p.unrealized_pnl_percent, p.realized_pnl,
			p.end_date, p.updated_at,
			u.username
		FROM positions p
		JOIN users u ON p.user_id = u.id
		WHERE u.persona_id = ?
		ORDER BY p.unrealized_pnl DESC
	`, persona.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query persona positions: %w", err)
	}
	defer rows.Close()

	positions := make([]*PositionWithUsername, 0)
	for rows.Next() {
		var pos PositionWithUsername
		err := rows.Scan(
			&pos.ID, &pos.UserID, &pos.Address, &pos.ConditionID, &pos.Asset,
			&pos.MarketTitle, &pos.MarketSlug, &pos.Outcome,
			&pos.Size, &pos.AvgPrice, &pos.CurrentPrice,
			&pos.InitialValue, &pos.CurrentValue,
			&pos.UnrealizedPnl, &pos.UnrealizedPnlPercent, &pos.RealizedPnl,
			&pos.EndDate, &pos.UpdatedAt,
			&pos.Username,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, &pos)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating positions: %w", err)
	}

	return positions, nil
}

// GetPersonaTrades retrieves combined trades across all accounts for a persona
func (s *storage) GetPersonaTrades(ctx context.Context, slug string, limit, offset int) ([]*TradeWithUsername, int, error) {
	persona, err := s.GetPersona(ctx, slug)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var total int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM trades t
		JOIN users u ON t.user_id = u.id
		WHERE u.persona_id = ?
	`, persona.ID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count trades: %w", err)
	}

	// Get trades
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			t.id, t.user_id, t.address, t.trade_id, t.condition_id,
			t.market_title, t.market_slug, t.outcome, t.side,
			t.price, t.size, t.value, t.timestamp, t.created_at,
			u.username
		FROM trades t
		JOIN users u ON t.user_id = u.id
		WHERE u.persona_id = ?
		ORDER BY t.timestamp DESC
		LIMIT ? OFFSET ?
	`, persona.ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query persona trades: %w", err)
	}
	defer rows.Close()

	trades := make([]*TradeWithUsername, 0)
	for rows.Next() {
		var t TradeWithUsername
		err := rows.Scan(
			&t.ID, &t.UserID, &t.Address, &t.TradeID, &t.ConditionID,
			&t.MarketTitle, &t.MarketSlug, &t.Outcome, &t.Side,
			&t.Price, &t.Size, &t.Value, &t.Timestamp, &t.CreatedAt,
			&t.Username,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating trades: %w", err)
	}

	return trades, total, nil
}

// GetUserPersonaInfo retrieves persona info for a user
func (s *storage) GetUserPersonaInfo(ctx context.Context, userID int64) (*PersonaInfo, error) {
	var info PersonaInfo
	err := s.db.QueryRowContext(ctx, `
		SELECT p.slug, p.display_name
		FROM personas p
		JOIN users u ON u.persona_id = p.id
		WHERE u.id = ?
	`, userID).Scan(&info.Slug, &info.DisplayName)

	if err == sql.ErrNoRows {
		return nil, nil // User has no persona
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user persona info: %w", err)
	}

	return &info, nil
}

// UpdateUserProfileImage updates a user's profile image
func (s *storage) UpdateUserProfileImage(ctx context.Context, userID int64, profileImage string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET profile_image = ? WHERE id = ?",
		profileImage, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user profile image: %w", err)
	}
	return nil
}

// UpdateUserOfficialPnl updates a user's official PnL and volume from Polymarket
func (s *storage) UpdateUserOfficialPnl(ctx context.Context, userID int64, pnl, volume float64) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET official_pnl = ?, official_volume = ? WHERE id = ?",
		pnl, volume, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user official pnl: %w", err)
	}
	return nil
}

// CreatePersonaWithImage creates a new persona with an image
func (s *storage) CreatePersonaWithImage(ctx context.Context, slug, displayName, image string) (*Persona, error) {
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO personas (slug, display_name, image, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)",
		slug, displayName, image,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert persona: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get persona id: %w", err)
	}

	return &Persona{
		ID:          id,
		Slug:        slug,
		DisplayName: displayName,
		Image:       &image,
	}, nil
}

// UpdatePersonaImage updates a persona's image
func (s *storage) UpdatePersonaImage(ctx context.Context, personaID int64, image string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE personas SET image = ? WHERE id = ?",
		image, personaID,
	)
	if err != nil {
		return fmt.Errorf("failed to update persona image: %w", err)
	}
	return nil
}

// GetUserResults retrieves resolved positions (results) for a user
// A position is considered "resolved" if:
// 1. The position has realized PnL (position was closed/exited)
// 2. The market has ended (end_date has passed)
func (s *storage) GetUserResults(ctx context.Context, userID int64, limit, offset int) ([]*Result, int, error) {
	// Get total count of resolved positions
	var total int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT condition_id)
		FROM positions
		WHERE user_id = ?
		AND realized_pnl IS NOT NULL
	`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count results: %w", err)
	}

	// Get results with pagination
	// We group by condition_id to avoid duplicates and sum realized_pnl across all positions for that market
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			MIN(id) as id,
			user_id,
			condition_id,
			market_title,
			market_slug,
			outcome,
			COALESCE(SUM(realized_pnl), 0) as realized_pnl,
			SUM(initial_value) as initial_value,
			end_date,
			MAX(updated_at) as resolution_date
		FROM positions
		WHERE user_id = ?
		AND realized_pnl IS NOT NULL
		GROUP BY condition_id, user_id
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query results: %w", err)
	}
	defer rows.Close()

	results := make([]*Result, 0, limit)
	for rows.Next() {
		var result Result
		var endDateStr, resolutionDateStr sql.NullString
		if err := rows.Scan(
			&result.ID,
			&result.UserID,
			&result.ConditionID,
			&result.MarketTitle,
			&result.MarketSlug,
			&result.Outcome,
			&result.RealizedPnl,
			&result.InitialValue,
			&endDateStr,
			&resolutionDateStr,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan result: %w", err)
		}
		// Parse date strings manually since SQLite returns strings
		if endDateStr.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", endDateStr.String); err == nil {
				result.EndDate = &t
			} else if t, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", endDateStr.String); err == nil {
				result.EndDate = &t
			}
		}
		if resolutionDateStr.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", resolutionDateStr.String); err == nil {
				result.ResolutionDate = &t
			} else if t, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", resolutionDateStr.String); err == nil {
				result.ResolutionDate = &t
			}
		}
		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating results: %w", err)
	}

	return results, total, nil
}

// GetPersonaResults retrieves resolved positions (results) across all accounts for a persona
func (s *storage) GetPersonaResults(ctx context.Context, slug string, limit, offset int) ([]*ResultWithUsername, int, error) {
	persona, err := s.GetPersona(ctx, slug)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var total int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT p.condition_id)
		FROM positions p
		JOIN users u ON p.user_id = u.id
		WHERE u.persona_id = ?
		AND p.realized_pnl IS NOT NULL
	`, persona.ID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count persona results: %w", err)
	}

	// Get results
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			MIN(p.id) as id,
			p.user_id,
			p.condition_id,
			p.market_title,
			p.market_slug,
			p.outcome,
			COALESCE(SUM(p.realized_pnl), 0) as realized_pnl,
			SUM(p.initial_value) as initial_value,
			p.end_date,
			MAX(p.updated_at) as resolution_date,
			u.username
		FROM positions p
		JOIN users u ON p.user_id = u.id
		WHERE u.persona_id = ?
		AND p.realized_pnl IS NOT NULL
		GROUP BY p.condition_id, u.username
		ORDER BY p.updated_at DESC
		LIMIT ? OFFSET ?
	`, persona.ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query persona results: %w", err)
	}
	defer rows.Close()

	results := make([]*ResultWithUsername, 0)
	for rows.Next() {
		var result ResultWithUsername
		var endDateStr, resolutionDateStr sql.NullString
		if err := rows.Scan(
			&result.ID,
			&result.UserID,
			&result.ConditionID,
			&result.MarketTitle,
			&result.MarketSlug,
			&result.Outcome,
			&result.RealizedPnl,
			&result.InitialValue,
			&endDateStr,
			&resolutionDateStr,
			&result.Username,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan persona result: %w", err)
		}
		// Parse date strings manually since SQLite returns strings
		if endDateStr.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", endDateStr.String); err == nil {
				result.EndDate = &t
			} else if t, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", endDateStr.String); err == nil {
				result.EndDate = &t
			}
		}
		if resolutionDateStr.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", resolutionDateStr.String); err == nil {
				result.ResolutionDate = &t
			} else if t, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", resolutionDateStr.String); err == nil {
				result.ResolutionDate = &t
			}
		}
		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating persona results: %w", err)
	}

	return results, total, nil
}

// CalculateRealizedPnlFromTrades calculates realized PnL using FIFO cost basis from trade history.
// This is the source of truth for realized PnL since closed positions are deleted during sync.
// Returns: realizedPnl, wins, totalClosed, error
func (s *storage) CalculateRealizedPnlFromTrades(ctx context.Context, userID int64) (float64, int, int, error) {
	trades, err := s.GetUserTradesChronological(ctx, userID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get trades: %w", err)
	}

	// Group trades by condition_id + outcome (each represents a unique position)
	type positionKey struct {
		conditionID string
		outcome     string
	}

	// FIFO lots per position
	inventory := make(map[positionKey][]fifoLot)
	realizedPnl := float64(0)
	wins := 0
	losses := 0

	for _, trade := range trades {
		if trade.ConditionID == nil || trade.Outcome == nil || trade.Side == nil {
			continue
		}
		if trade.Price == nil || trade.Size == nil {
			continue
		}

		key := positionKey{
			conditionID: *trade.ConditionID,
			outcome:     *trade.Outcome,
		}

		price := *trade.Price
		size := *trade.Size

		if *trade.Side == "BUY" {
			// Add to inventory
			inventory[key] = append(inventory[key], fifoLot{
				Shares: size,
				Price:  price,
			})
		} else if *trade.Side == "SELL" {
			// Match against FIFO lots and realize PnL
			lots := inventory[key]
			remainingToSell := size

			for remainingToSell > 0 && len(lots) > 0 {
				lot := &lots[0]

				if lot.Shares <= remainingToSell {
					// Consume entire lot
					costBasis := lot.Shares * lot.Price
					proceeds := lot.Shares * price
					pnl := proceeds - costBasis
					realizedPnl += pnl

					if pnl > 0 {
						wins++
					} else if pnl < 0 {
						losses++
					}

					remainingToSell -= lot.Shares
					lots = lots[1:] // Remove consumed lot
				} else {
					// Partial lot consumption
					costBasis := remainingToSell * lot.Price
					proceeds := remainingToSell * price
					pnl := proceeds - costBasis
					realizedPnl += pnl

					if pnl > 0 {
						wins++
					} else if pnl < 0 {
						losses++
					}

					lot.Shares -= remainingToSell
					remainingToSell = 0
				}
			}
			inventory[key] = lots
		}
	}

	return realizedPnl, wins, wins + losses, nil
}
