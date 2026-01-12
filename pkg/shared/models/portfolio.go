package models

import (
	"time"
)

// Position represents a trading position
type Position struct {
	ID               int       `json:"id" db:"id"`
	UserID           int       `json:"user_id" db:"user_id"`
	Symbol           string    `json:"symbol" db:"symbol"`
	Quantity         int64     `json:"quantity" db:"quantity"`
	Side             string    `json:"side" db:"side"` // "long" or "short"
	EntryPrice       float64   `json:"entry_price" db:"entry_price"`
	CurrentPrice     float64   `json:"current_price" db:"current_price"`
	UnrealizedPnL    float64   `json:"unrealized_pnl" db:"unrealized_pnl"`
	RealizedPnL      float64   `json:"realized_pnl" db:"realized_pnl"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// Portfolio represents a user's portfolio
type Portfolio struct {
	ID               int        `json:"id" db:"id"`
	UserID           int        `json:"user_id" db:"user_id"`
	Cash             float64    `json:"cash" db:"cash"`
	MarginUsed       float64    `json:"margin_used" db:"margin_used"`
	MarginAvailable  float64    `json:"margin_available" db:"margin_available"`
	TotalValue       float64    `json:"total_value" db:"total_value"`
	UnrealizedPnL    float64    `json:"unrealized_pnl" db:"unrealized_pnl"`
	RealizedPnL      float64    `json:"realized_pnl" db:"realized_pnl"`
	DayPnL           float64    `json:"day_pnl" db:"day_pnl"`
	Positions        []Position `json:"positions"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// Trade represents a trade transaction
type Trade struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	PositionID  int       `json:"position_id" db:"position_id"`
	Symbol      string    `json:"symbol" db:"symbol"`
	Quantity    int64     `json:"quantity" db:"quantity"`
	Price       float64   `json:"price" db:"price"`
	Side        string    `json:"side" db:"side"` // "buy" or "sell"
	Type        string    `json:"type" db:"type"` // "market", "limit", etc.
	Status      string    `json:"status" db:"status"` // "pending", "filled", "cancelled"
	Fees        float64   `json:"fees" db:"fees"`
	ExecutedAt  *time.Time `json:"executed_at" db:"executed_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PortfolioSummary provides a high-level view of portfolio performance
type PortfolioSummary struct {
	TotalValue      float64 `json:"total_value"`
	Cash            float64 `json:"cash"`
	PositionsValue  float64 `json:"positions_value"`
	UnrealizedPnL   float64 `json:"unrealized_pnl"`
	RealizedPnL     float64 `json:"realized_pnl"`
	DayPnL          float64 `json:"day_pnl"`
	DayReturn       float64 `json:"day_return"`
	TotalReturn     float64 `json:"total_return"`
	PositionCount   int     `json:"position_count"`
}

// PositionSummary provides aggregated position information
type PositionSummary struct {
	Symbol           string  `json:"symbol"`
	NetQuantity      int64   `json:"net_quantity"`
	LongQuantity     int64   `json:"long_quantity"`
	ShortQuantity    int64   `json:"short_quantity"`
	AveragePrice     float64 `json:"average_price"`
	CurrentPrice     float64 `json:"current_price"`
	MarketValue      float64 `json:"market_value"`
	UnrealizedPnL    float64 `json:"unrealized_pnl"`
	UnrealizedReturn float64 `json:"unrealized_return"`
}