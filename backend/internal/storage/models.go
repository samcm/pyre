package storage

import (
	"time"
)

// User represents a tracked user in the database
type User struct {
	ID             int64      `db:"id"`
	Username       string     `db:"username"`
	CreatedAt      time.Time  `db:"created_at"`
	LastSynced     *time.Time `db:"last_synced"`
	ProfileImage   *string    `db:"profile_image"`
	OfficialPnl    *float64   `db:"official_pnl"`    // All-time PnL from Polymarket profile page
	OfficialVolume *float64   `db:"official_volume"` // All-time volume from Polymarket profile page
}

// Address represents a wallet address associated with a user
type Address struct {
	ID      int64  `db:"id"`
	UserID  int64  `db:"user_id"`
	Address string `db:"address"`
}

// Position represents a current position in the database
type Position struct {
	ID                   int64      `db:"id"`
	UserID               int64      `db:"user_id"`
	Address              string     `db:"address"`
	ConditionID          string     `db:"condition_id"`
	Asset                string     `db:"asset"`
	MarketTitle          *string    `db:"market_title"`
	MarketSlug           *string    `db:"market_slug"`
	Outcome              *string    `db:"outcome"`
	Size                 *float64   `db:"size"`
	AvgPrice             *float64   `db:"avg_price"`
	CurrentPrice         *float64   `db:"current_price"`
	InitialValue         *float64   `db:"initial_value"`
	CurrentValue         *float64   `db:"current_value"`
	UnrealizedPnl        *float64   `db:"unrealized_pnl"`
	UnrealizedPnlPercent *float64   `db:"unrealized_pnl_percent"`
	RealizedPnl          *float64   `db:"realized_pnl"`
	EndDate              *time.Time `db:"end_date"`
	UpdatedAt            time.Time  `db:"updated_at"`
}

// Trade represents a historical trade in the database
type Trade struct {
	ID          int64      `db:"id"`
	UserID      int64      `db:"user_id"`
	Address     string     `db:"address"`
	TradeID     *string    `db:"trade_id"`
	ConditionID *string    `db:"condition_id"`
	MarketTitle *string    `db:"market_title"`
	MarketSlug  *string    `db:"market_slug"`
	Outcome     *string    `db:"outcome"`
	Side        *string    `db:"side"`
	Price       *float64   `db:"price"`
	Size        *float64   `db:"size"`
	Value       *float64   `db:"value"`
	Timestamp   *time.Time `db:"timestamp"`
	CreatedAt   time.Time  `db:"created_at"`
}

// TradeWithUsername represents a trade with the associated username
type TradeWithUsername struct {
	Trade
	Username string `db:"username"`
}

// TradeFilters represents filtering options for trades
type TradeFilters struct {
	Limit         int
	Offset        int
	Username      *string
	Side          *string
	MinValue      *float64
	SortBy        string
	SortDirection string
}

// PnlSnapshot represents a point-in-time PNL snapshot
type PnlSnapshot struct {
	ID            int64     `db:"id"`
	UserID        int64     `db:"user_id"`
	Timestamp     time.Time `db:"timestamp"`
	TotalPnl      *float64  `db:"total_pnl"`
	RealizedPnl   *float64  `db:"realized_pnl"`
	UnrealizedPnl *float64  `db:"unrealized_pnl"`
}

// UserStats represents aggregated statistics for a user
type UserStats struct {
	Username      string
	Addresses     []string
	ProfileImage  *string
	TotalPnl      float64
	RealizedPnl   float64
	UnrealizedPnl float64
	OpenPositions int
	TotalTrades   int
	WinRate       float64
	LastSynced    *time.Time
}

// Persona represents a real person mapped to multiple usernames
type Persona struct {
	ID          int64     `db:"id"`
	Slug        string    `db:"slug"`
	DisplayName string    `db:"display_name"`
	Image       *string   `db:"image"`
	CreatedAt   time.Time `db:"created_at"`
}

// PersonaStats represents aggregated statistics for a persona across all their users
type PersonaStats struct {
	Slug          string
	DisplayName   string
	Image         *string
	Usernames     []string
	TotalPnl      float64
	RealizedPnl   float64
	UnrealizedPnl float64
	OpenPositions int
	TotalTrades   int
	WinRate       float64
}

// PersonaAccount represents a user account belonging to a persona with individual stats
type PersonaAccount struct {
	Username      string
	Addresses     []string
	ProfileImage  *string
	TotalPnl      float64
	RealizedPnl   float64
	UnrealizedPnl float64
	OpenPositions int
	TotalTrades   int
	WinRate       float64
}

// BiggestTrade represents a trade with additional context for biggest wins/losses
type BiggestTrade struct {
	Trade
	Username    string  `db:"username"`
	RealizedPnl float64 `db:"realized_pnl"`
}

// PositionWithUsername represents a position with the associated username
type PositionWithUsername struct {
	Position
	Username string `db:"username"`
}

// PersonaInfo represents basic persona information for a user
type PersonaInfo struct {
	Slug        string
	DisplayName string
}

// Result represents a resolved position with win/loss information
type Result struct {
	ID             int64      `db:"id"`
	UserID         int64      `db:"user_id"`
	ConditionID    string     `db:"condition_id"`
	MarketTitle    *string    `db:"market_title"`
	MarketSlug     *string    `db:"market_slug"`
	Outcome        *string    `db:"outcome"`
	RealizedPnl    float64    `db:"realized_pnl"`
	InitialValue   *float64   `db:"initial_value"`
	EndDate        *time.Time `db:"end_date"`
	ResolutionDate *time.Time `db:"resolution_date"` // When position was closed or market ended
}

// ResultWithUsername represents a result with the associated username
type ResultWithUsername struct {
	Result
	Username string `db:"username"`
}

// fifoLot represents a single lot of shares for FIFO cost basis tracking
type fifoLot struct {
	Shares float64
	Price  float64 // Price per share
}
