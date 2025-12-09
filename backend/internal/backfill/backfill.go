package backfill

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/samcm/pyre/internal/storage"
	"github.com/sirupsen/logrus"
)

// Result contains the results of a backfill operation
type Result struct {
	Username         string
	TradesProcessed  int
	SnapshotsCreated int
	TotalRealizedPnl float64
	OldestTradeDate  *time.Time
	NewestTradeDate  *time.Time
}

// Service provides PnL backfill functionality
type Service interface {
	BackfillUser(ctx context.Context, username string) (*Result, error)
}

// service implements the backfill Service
type service struct {
	storage storage.Storage
	log     logrus.FieldLogger
}

var _ Service = (*service)(nil)

// NewService creates a new backfill service
func NewService(storage storage.Storage, log logrus.FieldLogger) Service {
	return &service{
		storage: storage,
		log:     log.WithField("package", "backfill"),
	}
}

// lot represents a single buy lot for FIFO cost basis tracking
type lot struct {
	price float64
	size  float64
}

// positionKey uniquely identifies a position by condition and outcome
type positionKey struct {
	conditionID string
	outcome     string
}

// BackfillUser reconstructs PnL history from trade data for a user
func (s *service) BackfillUser(ctx context.Context, username string) (*Result, error) {
	s.log.WithField("username", username).Info("starting backfill")

	// Get user
	user, err := s.storage.GetUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get all trades sorted chronologically
	trades, err := s.storage.GetUserTradesChronological(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	if len(trades) == 0 {
		return &Result{
			Username:         username,
			TradesProcessed:  0,
			SnapshotsCreated: 0,
			TotalRealizedPnl: 0,
		}, nil
	}

	// Delete existing PnL snapshots
	if err := s.storage.DeleteUserPnlSnapshots(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to delete existing snapshots: %w", err)
	}

	// Track cost basis per position using FIFO
	// Key: conditionID + outcome
	costBasis := make(map[positionKey][]lot, 32)

	// Track cumulative realized PnL
	var cumulativeRealizedPnl float64

	// Track daily PnL for snapshots
	// Key: date (truncated to day)
	dailyPnl := make(map[time.Time]float64, 64)

	var oldestDate, newestDate *time.Time

	for _, trade := range trades {
		if trade.Timestamp == nil || trade.ConditionID == nil || trade.Outcome == nil ||
			trade.Side == nil || trade.Price == nil || trade.Size == nil {
			continue
		}

		key := positionKey{
			conditionID: *trade.ConditionID,
			outcome:     *trade.Outcome,
		}

		price := *trade.Price
		size := *trade.Size
		timestamp := *trade.Timestamp
		day := timestamp.Truncate(24 * time.Hour)

		// Track date range
		if oldestDate == nil || timestamp.Before(*oldestDate) {
			oldestDate = &timestamp
		}
		if newestDate == nil || timestamp.After(*newestDate) {
			newestDate = &timestamp
		}

		switch *trade.Side {
		case "BUY":
			// Add lot to FIFO queue
			if _, exists := costBasis[key]; !exists {
				costBasis[key] = make([]lot, 0, 8)
			}
			costBasis[key] = append(costBasis[key], lot{price: price, size: size})

		case "SELL":
			// Calculate realized PnL using FIFO
			realizedPnl := s.calculateRealizedPnlFIFO(costBasis, key, price, size)
			cumulativeRealizedPnl += realizedPnl

			// Record in daily map
			dailyPnl[day] = cumulativeRealizedPnl
		}
	}

	// Also record the final state for days with only buys
	for _, trade := range trades {
		if trade.Timestamp == nil || trade.Side == nil {
			continue
		}
		day := trade.Timestamp.Truncate(24 * time.Hour)
		// Only set if not already set (preserve sell-day values)
		if _, exists := dailyPnl[day]; !exists {
			// Find cumulative PnL up to this point
			// This is a simplification - we use the current cumulative value
			// In a more sophisticated implementation, we'd track this per-day
		}
	}

	// Create snapshots from daily PnL data
	snapshots := s.createSnapshots(user.ID, dailyPnl)

	// Sort snapshots by timestamp
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
	})

	// Bulk insert snapshots
	if err := s.storage.BulkInsertPnlSnapshots(ctx, snapshots); err != nil {
		return nil, fmt.Errorf("failed to insert snapshots: %w", err)
	}

	result := &Result{
		Username:         username,
		TradesProcessed:  len(trades),
		SnapshotsCreated: len(snapshots),
		TotalRealizedPnl: cumulativeRealizedPnl,
		OldestTradeDate:  oldestDate,
		NewestTradeDate:  newestDate,
	}

	s.log.WithFields(logrus.Fields{
		"username":          username,
		"trades_processed":  result.TradesProcessed,
		"snapshots_created": result.SnapshotsCreated,
		"total_realized":    result.TotalRealizedPnl,
	}).Info("backfill completed")

	return result, nil
}

// calculateRealizedPnlFIFO calculates realized PnL for a sell using FIFO cost basis
func (s *service) calculateRealizedPnlFIFO(costBasis map[positionKey][]lot, key positionKey, sellPrice, sellSize float64) float64 {
	lots, exists := costBasis[key]
	if !exists || len(lots) == 0 {
		// No cost basis - selling something we didn't buy (possibly from before tracking)
		// Assume zero cost basis
		return sellPrice * sellSize
	}

	var realizedPnl float64
	remainingToSell := sellSize

	for remainingToSell > 0 && len(lots) > 0 {
		currentLot := &lots[0]

		if currentLot.size <= remainingToSell {
			// Use entire lot
			pnl := (sellPrice - currentLot.price) * currentLot.size
			realizedPnl += pnl
			remainingToSell -= currentLot.size

			// Remove lot from queue
			lots = lots[1:]
		} else {
			// Partial lot usage
			pnl := (sellPrice - currentLot.price) * remainingToSell
			realizedPnl += pnl
			currentLot.size -= remainingToSell
			remainingToSell = 0
		}
	}

	// Update the cost basis map
	costBasis[key] = lots

	// If we still have shares to sell with no cost basis, assume zero cost
	if remainingToSell > 0 {
		realizedPnl += sellPrice * remainingToSell
	}

	return realizedPnl
}

// createSnapshots creates PnL snapshots from daily PnL data
func (s *service) createSnapshots(userID int64, dailyPnl map[time.Time]float64) []*storage.PnlSnapshot {
	snapshots := make([]*storage.PnlSnapshot, 0, len(dailyPnl))

	for day, realizedPnl := range dailyPnl {
		pnl := realizedPnl
		// For backfilled data, unrealized is 0 since we're only tracking realized
		zero := 0.0
		snapshots = append(snapshots, &storage.PnlSnapshot{
			UserID:        userID,
			Timestamp:     day,
			TotalPnl:      &pnl,
			RealizedPnl:   &pnl,
			UnrealizedPnl: &zero,
		})
	}

	return snapshots
}
