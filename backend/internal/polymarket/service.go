package polymarket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/samcm/pyre/internal/storage"
	"github.com/sirupsen/logrus"
)

// Service defines the interface for the sync service
type Service interface {
	Start(ctx context.Context) error
	Stop() error
	TriggerSync(ctx context.Context) error
}

// service implements the sync service
type service struct {
	client   Client
	storage  storage.Storage
	users    map[string][]string // username -> addresses
	interval time.Duration
	log      logrus.FieldLogger

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	done   chan struct{}
}

var _ Service = (*service)(nil)

// NewService creates a new sync service
func NewService(client Client, storage storage.Storage, users map[string][]string, intervalMinutes int, log logrus.FieldLogger) Service {
	return &service{
		client:   client,
		storage:  storage,
		users:    users,
		interval: time.Duration(intervalMinutes) * time.Minute,
		log:      log.WithField("package", "polymarket-service"),
		done:     make(chan struct{}),
	}
}

// Start begins the sync service
func (s *service) Start(ctx context.Context) error {
	s.log.Info("starting polymarket sync service")

	s.ctx, s.cancel = context.WithCancel(ctx)

	// Ensure all users exist in database
	if err := s.ensureUsers(s.ctx); err != nil {
		return fmt.Errorf("failed to ensure users: %w", err)
	}

	// Perform initial sync
	s.log.Info("performing initial sync")
	if err := s.syncAll(s.ctx); err != nil {
		s.log.WithError(err).Error("initial sync failed")
	}

	// Start background sync goroutine
	s.wg.Add(1)
	go s.syncLoop()

	s.log.WithField("interval", s.interval).Info("polymarket sync service started")
	return nil
}

// Stop stops the sync service
func (s *service) Stop() error {
	s.log.Info("stopping polymarket sync service")

	close(s.done)
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()

	s.log.Info("polymarket sync service stopped")
	return nil
}

// TriggerSync manually triggers a sync
func (s *service) TriggerSync(ctx context.Context) error {
	s.log.Info("manual sync triggered")
	return s.syncAll(ctx)
}

// syncLoop runs periodic syncs
func (s *service) syncLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.log.Info("starting scheduled sync")
			if err := s.syncAll(s.ctx); err != nil {
				s.log.WithError(err).Error("scheduled sync failed")
			}
		}
	}
}

// ensureUsers ensures all configured users exist in the database
func (s *service) ensureUsers(ctx context.Context) error {
	for username, addresses := range s.users {
		_, err := s.storage.GetUser(ctx, username)
		if err != nil {
			// User doesn't exist, create it
			s.log.WithField("username", username).Info("creating user")
			if _, err := s.storage.CreateUser(ctx, username, addresses); err != nil {
				return fmt.Errorf("failed to create user %s: %w", username, err)
			}
		}
	}
	return nil
}

// syncAll syncs data for all configured users
func (s *service) syncAll(ctx context.Context) error {
	s.log.WithField("users", len(s.users)).Info("syncing all users")

	for username, addresses := range s.users {
		if err := s.syncUser(ctx, username, addresses); err != nil {
			s.log.WithError(err).WithField("username", username).Error("failed to sync user")
			// Continue with other users even if one fails
			continue
		}
	}

	s.log.Info("sync completed for all users")
	return nil
}

// syncUser syncs data for a single user
func (s *service) syncUser(ctx context.Context, username string, addresses []string) error {
	s.log.WithFields(logrus.Fields{
		"username":  username,
		"addresses": len(addresses),
	}).Info("syncing user")

	user, err := s.storage.GetUser(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Sync profile data from first address
	var polymarketUsername string
	if len(addresses) > 0 {
		profile, err := s.client.GetUserProfile(ctx, addresses[0])
		if err != nil {
			s.log.WithError(err).WithField("username", username).Warn("failed to fetch user profile")
		} else if profile != nil {
			// Get the correct Polymarket username (case-sensitive)
			// Use Name (public display name) which is used in profile URLs
			polymarketUsername = profile.Name
			if profile.ProfileImage != "" {
				if err := s.storage.UpdateUserProfileImage(ctx, user.ID, profile.ProfileImage); err != nil {
					s.log.WithError(err).WithField("username", username).Warn("failed to update user profile image")
				}
			}
		}
	}

	// Fetch official PnL from Polymarket profile page (all-time accurate data)
	// Use the Polymarket pseudonym (case-sensitive username) if available
	if polymarketUsername != "" && len(addresses) > 0 {
		portfolioStats, err := s.client.GetPortfolioStats(ctx, polymarketUsername, addresses[0])
		if err != nil {
			s.log.WithError(err).WithFields(logrus.Fields{
				"username":           username,
				"polymarketUsername": polymarketUsername,
			}).Warn("failed to fetch portfolio stats")
		} else if portfolioStats != nil {
			if err := s.storage.UpdateUserOfficialPnl(ctx, user.ID, portfolioStats.TotalPnl, portfolioStats.TotalVolume); err != nil {
				s.log.WithError(err).WithField("username", username).Warn("failed to update official pnl")
			} else {
				s.log.WithFields(logrus.Fields{
					"username":           username,
					"polymarketUsername": polymarketUsername,
					"pnl":                portfolioStats.TotalPnl,
					"volume":             portfolioStats.TotalVolume,
				}).Info("updated official PnL from Polymarket")
			}
		}
	}

	// Clear existing positions (we'll replace with fresh data)
	if err := s.storage.DeleteUserPositions(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to delete existing positions: %w", err)
	}

	var totalPositions, totalTrades int

	// Sync each address
	for _, address := range addresses {
		positions, trades, err := s.syncAddress(ctx, user.ID, address)
		if err != nil {
			s.log.WithError(err).WithFields(logrus.Fields{
				"username": username,
				"address":  address,
			}).Error("failed to sync address")
			continue
		}
		totalPositions += positions
		totalTrades += trades
	}

	// Take PNL snapshot
	if err := s.takePnlSnapshot(ctx, user.ID); err != nil {
		s.log.WithError(err).WithField("username", username).Error("failed to take pnl snapshot")
	}

	// Update last synced timestamp
	if err := s.storage.UpdateUserLastSynced(ctx, user.ID, time.Now()); err != nil {
		return fmt.Errorf("failed to update last synced: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"username":  username,
		"positions": totalPositions,
		"trades":    totalTrades,
	}).Info("user sync completed")

	return nil
}

// syncAddress syncs data for a single address
func (s *service) syncAddress(ctx context.Context, userID int64, address string) (int, int, error) {
	s.log.WithField("address", address).Debug("syncing address")

	// Fetch positions
	positions, err := s.client.GetPositions(ctx, address)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch positions: %w", err)
	}

	// Store positions
	for _, pos := range positions {
		dbPos := &storage.Position{
			UserID:               userID,
			Address:              address,
			ConditionID:          pos.ConditionID,
			Asset:                pos.Asset,
			Size:                 pos.Size,
			AvgPrice:             pos.AvgPrice,
			CurrentPrice:         pos.CurrentPrice,
			InitialValue:         pos.InitialValue,
			CurrentValue:         pos.CurrentValue,
			UnrealizedPnl:        pos.UnrealizedPnl,
			UnrealizedPnlPercent: pos.UnrealizedPnlPercent,
			RealizedPnl:          pos.RealizedPnl,
		}

		// Market info is inline in the API response
		if pos.Title != "" {
			dbPos.MarketTitle = &pos.Title
		}
		if pos.Slug != "" {
			dbPos.MarketSlug = &pos.Slug
		}
		if pos.EndDate != "" {
			// Parse date string like "2025-12-10"
			if endDate, err := time.Parse("2006-01-02", pos.EndDate); err == nil {
				dbPos.EndDate = &endDate
			}
		}

		if pos.Outcome != "" {
			dbPos.Outcome = &pos.Outcome
		}

		if err := s.storage.UpsertPosition(ctx, dbPos); err != nil {
			s.log.WithError(err).WithField("condition_id", pos.ConditionID).Error("failed to upsert position")
		}
	}

	// Fetch trades (limit to last 100)
	trades, err := s.client.GetTrades(ctx, address, 100)
	if err != nil {
		return len(positions), 0, fmt.Errorf("failed to fetch trades: %w", err)
	}

	// Store trades
	for _, trade := range trades {
		dbTrade := &storage.Trade{
			UserID:  userID,
			Address: address,
			Price:   trade.Price,
			Size:    trade.Size,
		}

		if trade.ID != "" {
			dbTrade.TradeID = &trade.ID
		}
		if trade.ConditionID != "" {
			dbTrade.ConditionID = &trade.ConditionID
		}
		if trade.Outcome != "" {
			dbTrade.Outcome = &trade.Outcome
		}
		if trade.Side != "" {
			dbTrade.Side = &trade.Side
		}
		if trade.Timestamp > 0 {
			// Convert Unix timestamp to time.Time
			ts := time.Unix(trade.Timestamp, 0)
			dbTrade.Timestamp = &ts
		}

		// Market info is inline
		if trade.Title != "" {
			dbTrade.MarketTitle = &trade.Title
		}
		if trade.Slug != "" {
			dbTrade.MarketSlug = &trade.Slug
		}

		// Calculate value if not present
		if trade.Price != nil && trade.Size != nil {
			value := *trade.Price * *trade.Size
			dbTrade.Value = &value
		}

		if err := s.storage.InsertTrade(ctx, dbTrade); err != nil {
			// Ignore duplicate trade errors
			s.log.WithError(err).WithField("trade_id", trade.ID).Debug("failed to insert trade (likely duplicate)")
		}
	}

	s.log.WithFields(logrus.Fields{
		"address":   address,
		"positions": len(positions),
		"trades":    len(trades),
	}).Debug("address sync completed")

	return len(positions), len(trades), nil
}

// takePnlSnapshot takes a snapshot of current PNL for a user
func (s *service) takePnlSnapshot(ctx context.Context, userID int64) error {
	// Get all users and find the matching one
	users, err := s.storage.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	var username string
	for _, u := range users {
		if u.ID == userID {
			username = u.Username
			break
		}
	}

	if username == "" {
		return fmt.Errorf("user not found with id %d", userID)
	}

	stats, err := s.storage.GetUserStats(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to get user stats: %w", err)
	}

	snapshot := &storage.PnlSnapshot{
		UserID:        userID,
		Timestamp:     time.Now(),
		TotalPnl:      &stats.TotalPnl,
		RealizedPnl:   &stats.RealizedPnl,
		UnrealizedPnl: &stats.UnrealizedPnl,
	}

	if err := s.storage.InsertPnlSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("failed to insert pnl snapshot: %w", err)
	}

	return nil
}
