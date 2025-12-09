package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// migrations contains the database schema migrations
var migrations = []string{
	// Users table
	`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_synced DATETIME
	)`,

	// Addresses table (many-to-one with users)
	`CREATE TABLE IF NOT EXISTS addresses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		address TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE(user_id, address)
	)`,

	// Positions table (current snapshot)
	`CREATE TABLE IF NOT EXISTS positions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		address TEXT NOT NULL,
		condition_id TEXT NOT NULL,
		asset TEXT NOT NULL,
		market_title TEXT,
		market_slug TEXT,
		outcome TEXT,
		size REAL,
		avg_price REAL,
		current_price REAL,
		initial_value REAL,
		current_value REAL,
		unrealized_pnl REAL,
		unrealized_pnl_percent REAL,
		realized_pnl REAL,
		end_date DATETIME,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE(user_id, address, condition_id, asset)
	)`,

	// Trades table (historical)
	`CREATE TABLE IF NOT EXISTS trades (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		address TEXT NOT NULL,
		trade_id TEXT,
		condition_id TEXT NOT NULL,
		market_title TEXT,
		market_slug TEXT,
		outcome TEXT,
		side TEXT NOT NULL,
		price REAL NOT NULL,
		size REAL NOT NULL,
		value REAL,
		timestamp DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE(user_id, condition_id, timestamp, side, size, price)
	)`,

	// PNL snapshots table (for historical charts)
	`CREATE TABLE IF NOT EXISTS pnl_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		total_pnl REAL,
		realized_pnl REAL,
		unrealized_pnl REAL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`,

	// Indexes
	`CREATE INDEX IF NOT EXISTS idx_positions_user ON positions(user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_trades_user ON trades(user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_trades_timestamp ON trades(timestamp)`,
	`CREATE INDEX IF NOT EXISTS idx_pnl_snapshots_user_time ON pnl_snapshots(user_id, timestamp)`,

	// Personas table
	`CREATE TABLE IF NOT EXISTS personas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT UNIQUE NOT NULL,
		display_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Add persona_id column to users table (nullable for backwards compatibility)
	`ALTER TABLE users ADD COLUMN persona_id INTEGER REFERENCES personas(id)`,

	// Index for user-persona relationship
	`CREATE INDEX IF NOT EXISTS idx_users_persona ON users(persona_id)`,

	// Add profile_image column to users table
	`ALTER TABLE users ADD COLUMN profile_image TEXT`,

	// Add image column to personas table
	`ALTER TABLE personas ADD COLUMN image TEXT`,

	// Add official PnL columns to users table (scraped from Polymarket profile page)
	`ALTER TABLE users ADD COLUMN official_pnl REAL`,
	`ALTER TABLE users ADD COLUMN official_volume REAL`,
}

// runMigrations executes all database migrations
func runMigrations(ctx context.Context, db *sql.DB) error {
	// Create migrations tracking table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	var currentVersion int
	err = db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	// Apply pending migrations
	for i := currentVersion; i < len(migrations); i++ {
		version := i + 1
		migration := migrations[i]

		// Start transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", version, err)
		}

		// Execute migration
		if _, err := tx.ExecContext(ctx, migration); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("failed to execute migration %d: %w (rollback error: %v)", version, err, rbErr)
			}
			return fmt.Errorf("failed to execute migration %d: %w", version, err)
		}

		// Record migration
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("failed to record migration %d: %w (rollback error: %v)", version, err, rbErr)
			}
			return fmt.Errorf("failed to record migration %d: %w", version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", version, err)
		}
	}

	return nil
}
