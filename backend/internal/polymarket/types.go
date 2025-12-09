package polymarket

import "time"

// PositionResponse represents a position from the Polymarket API
type PositionResponse struct {
	Asset        string   `json:"asset"`
	ConditionID  string   `json:"conditionId"`
	Outcome      string   `json:"outcome"`
	Size         *float64 `json:"size"`
	AvgPrice     *float64 `json:"avgPrice"`
	CurrentPrice *float64 `json:"curPrice"`
	InitialValue *float64 `json:"initialValue"`
	CurrentValue *float64 `json:"currentValue"`
	// cashPnl is the unrealized PnL
	UnrealizedPnl *float64 `json:"cashPnl"`
	// percentPnl is the unrealized PnL percent
	UnrealizedPnlPercent *float64 `json:"percentPnl"`
	RealizedPnl          *float64 `json:"realizedPnl"`
	// Market info is inline, not nested
	Title   string `json:"title"`
	Slug    string `json:"slug"`
	EndDate string `json:"endDate"` // Date string like "2025-12-10"
}

// Market represents market information from Polymarket (for compatibility)
type Market struct {
	Title   string
	Slug    string
	EndDate *time.Time
}

// TradeResponse represents a trade from the Polymarket API
type TradeResponse struct {
	ID          string   `json:"id"`
	ConditionID string   `json:"conditionId"`
	Outcome     string   `json:"outcome"`
	Side        string   `json:"side"` // BUY or SELL
	Price       *float64 `json:"price"`
	Size        *float64 `json:"size"`
	// Timestamp is a Unix timestamp
	Timestamp int64 `json:"timestamp"`
	// Market info is inline
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	EventSlug string `json:"eventSlug"`
}

// ActivityResponse represents activity from the Polymarket API
type ActivityResponse struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	ConditionID string   `json:"conditionId"`
	Outcome     string   `json:"outcome"`
	Side        string   `json:"side"`
	Price       *float64 `json:"price"`
	Size        *float64 `json:"size"`
	Timestamp   int64    `json:"timestamp"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
}

// PositionsResponse is a list of positions
type PositionsResponse []PositionResponse

// TradesResponse is a list of trades
type TradesResponse []TradeResponse

// ActivitiesResponse is a list of activities
type ActivitiesResponse []ActivityResponse

// ProfileResponse represents user profile data from the Polymarket API
type ProfileResponse struct {
	Name                  string `json:"name"`
	Pseudonym             string `json:"pseudonym"`
	Bio                   string `json:"bio"`
	ProfileImage          string `json:"profileImage"`
	ProfileImageOptimized string `json:"profileImageOptimized"`
}

// PortfolioStats represents the all-time portfolio statistics from Polymarket
type PortfolioStats struct {
	TotalPnl      float64 `json:"pnl"`
	TotalVolume   float64 `json:"amount"`
	RealizedPnl   float64 `json:"realized"`
	UnrealizedPnl float64 `json:"unrealized"`
}
