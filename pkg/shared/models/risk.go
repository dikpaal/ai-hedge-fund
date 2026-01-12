package models

import "time"

// RiskMetrics represents risk calculations for a position or portfolio
type RiskMetrics struct {
	Symbol              string    `json:"symbol"`
	Volatility          float64   `json:"volatility"`           // Annualized volatility
	VaR95               float64   `json:"var_95"`               // 95% Value at Risk
	VaR99               float64   `json:"var_99"`               // 99% Value at Risk
	MaxDrawdown         float64   `json:"max_drawdown"`         // Maximum historical drawdown
	SharpeRatio         float64   `json:"sharpe_ratio"`         // Risk-adjusted return
	Beta                float64   `json:"beta"`                 // Market beta
	PositionLimit       float64   `json:"position_limit"`       // Maximum position size
	RemainingLimit      float64   `json:"remaining_limit"`      // Remaining position capacity
	CorrelationToMarket float64   `json:"correlation_to_market"`
	CalculatedAt        time.Time `json:"calculated_at"`
}

// PortfolioRisk represents portfolio-level risk metrics
type PortfolioRisk struct {
	UserID               int                     `json:"user_id"`
	TotalVaR95           float64                 `json:"total_var_95"`
	TotalVaR99           float64                 `json:"total_var_99"`
	PortfolioVolatility  float64                 `json:"portfolio_volatility"`
	PortfolioBeta        float64                 `json:"portfolio_beta"`
	PortfolioSharpe      float64                 `json:"portfolio_sharpe"`
	ConcentrationRisk    float64                 `json:"concentration_risk"`    // Largest position as % of portfolio
	LeverageRatio        float64                 `json:"leverage_ratio"`        // Total exposure / equity
	MarginUtilization    float64                 `json:"margin_utilization"`    // Used margin / available margin
	PositionRisks        map[string]RiskMetrics  `json:"position_risks"`
	CorrelationMatrix    [][]float64             `json:"correlation_matrix"`
	CalculatedAt         time.Time               `json:"calculated_at"`
}

// RiskLimit represents risk limits for trading
type RiskLimit struct {
	ID                  int       `json:"id" db:"id"`
	UserID              int       `json:"user_id" db:"user_id"`
	Symbol              string    `json:"symbol" db:"symbol"`              // Empty for portfolio-level limits
	MaxPositionSize     float64   `json:"max_position_size" db:"max_position_size"`
	MaxDailyLoss        float64   `json:"max_daily_loss" db:"max_daily_loss"`
	MaxPortfolioRisk    float64   `json:"max_portfolio_risk" db:"max_portfolio_risk"`
	MaxLeverage         float64   `json:"max_leverage" db:"max_leverage"`
	MaxConcentration    float64   `json:"max_concentration" db:"max_concentration"`    // Max % in single position
	StopLossPercentage  float64   `json:"stop_loss_percentage" db:"stop_loss_percentage"`
	IsActive            bool      `json:"is_active" db:"is_active"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// RiskAlert represents a risk alert/warning
type RiskAlert struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	AlertType   string    `json:"alert_type" db:"alert_type"`     // "position_limit", "daily_loss", "var_breach"
	Severity    string    `json:"severity" db:"severity"`         // "warning", "critical"
	Symbol      string    `json:"symbol" db:"symbol"`
	Message     string    `json:"message" db:"message"`
	CurrentValue float64   `json:"current_value" db:"current_value"`
	ThresholdValue float64 `json:"threshold_value" db:"threshold_value"`
	IsResolved  bool      `json:"is_resolved" db:"is_resolved"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at" db:"resolved_at"`
}

// VolatilityData represents historical volatility calculations
type VolatilityData struct {
	Symbol           string    `json:"symbol"`
	Period           int       `json:"period"`           // Period in days
	DailyVolatility  float64   `json:"daily_volatility"`
	WeeklyVolatility float64   `json:"weekly_volatility"`
	MonthlyVolatility float64  `json:"monthly_volatility"`
	AnnualizedVolatility float64 `json:"annualized_volatility"`
	CalculatedAt     time.Time `json:"calculated_at"`
}