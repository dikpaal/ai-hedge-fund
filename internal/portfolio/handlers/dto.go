package handlers

import "time"

// Request DTOs

type CreatePortfolioRequest struct {
	UserID      int     `json:"user_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	InitialCash float64 `json:"initial_cash" binding:"required,gt=0"`
}

type UpdatePortfolioRequest struct {
	Cash float64 `json:"cash" binding:"gte=0"`
}

type TradeRequest struct {
	Symbol    string `json:"symbol" binding:"required"`
	Side      string `json:"side" binding:"required,oneof=buy sell"`
	Quantity  int64  `json:"quantity" binding:"required,gt=0"`
	OrderType string `json:"order_type" binding:"required,oneof=market limit"`
	Price     float64 `json:"price"` // Only for limit orders
}

type RebalanceRequest struct {
	TargetAllocations map[string]float64 `json:"target_allocations" binding:"required"`
}

// Response DTOs

type PortfolioResponse struct {
	ID               int                `json:"id"`
	UserID           int                `json:"user_id"`
	Name             string             `json:"name"`
	Cash             float64            `json:"cash"`
	MarginUsed       float64            `json:"margin_used"`
	MarginAvailable  float64            `json:"margin_available"`
	TotalValue       float64            `json:"total_value"`
	UnrealizedPnL    float64            `json:"unrealized_pnl"`
	RealizedPnL      float64            `json:"realized_pnl"`
	DayPnL           float64            `json:"day_pnl"`
	Positions        []PositionResponse `json:"positions"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

type PositionResponse struct {
	ID            int       `json:"id"`
	PortfolioID   int       `json:"portfolio_id"`
	Symbol        string    `json:"symbol"`
	Quantity      int64     `json:"quantity"`
	Side          string    `json:"side"`
	EntryPrice    float64   `json:"entry_price"`
	CurrentPrice  float64   `json:"current_price"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	RealizedPnL   float64   `json:"realized_pnl"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TradeResponse struct {
	ID          int        `json:"id"`
	PortfolioID int        `json:"portfolio_id"`
	PositionID  int        `json:"position_id"`
	Symbol      string     `json:"symbol"`
	Quantity    int64      `json:"quantity"`
	Price       float64    `json:"price"`
	Side        string     `json:"side"`
	Type        string     `json:"type"`
	Status      string     `json:"status"`
	Fees        float64    `json:"fees"`
	ExecutedAt  *time.Time `json:"executed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type SummaryResponse struct {
	TotalValue     float64 `json:"total_value"`
	Cash           float64 `json:"cash"`
	PositionsValue float64 `json:"positions_value"`
	UnrealizedPnL  float64 `json:"unrealized_pnl"`
	RealizedPnL    float64 `json:"realized_pnl"`
	DayPnL         float64 `json:"day_pnl"`
	DayReturn      float64 `json:"day_return"`
	TotalReturn    float64 `json:"total_return"`
	PositionCount  int     `json:"position_count"`
}

type AllocationResponse struct {
	Symbol     string  `json:"symbol"`
	Percentage float64 `json:"percentage"`
	Value      float64 `json:"value"`
}

type RiskMetricsResponse struct {
	TotalValue            float64 `json:"total_value"`
	PositionCount         int     `json:"position_count"`
	MaxPositionPercent    float64 `json:"max_position_percent"`
	CashPercent           float64 `json:"cash_percent"`
	DiversificationScore  float64 `json:"diversification_score"`
}

type RebalanceRecommendation struct {
	Symbol          string  `json:"symbol"`
	CurrentPercent  float64 `json:"current_percent"`
	TargetPercent   float64 `json:"target_percent"`
	Difference      float64 `json:"difference"`
	TargetValue     float64 `json:"target_value"`
	CurrentValue    float64 `json:"current_value"`
	Action          string  `json:"action"` // "buy", "sell", "hold"
	EstimatedShares int64   `json:"estimated_shares"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
