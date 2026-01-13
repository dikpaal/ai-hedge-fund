package models

import "time"

// Job represents a background job
type Job struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Priority  int                    `json:"priority"`   // Higher number = higher priority
	MaxRetries int                   `json:"max_retries"`
	Retries   int                    `json:"retries"`
	CreatedAt time.Time              `json:"created_at"`
	ScheduledAt *time.Time           `json:"scheduled_at,omitempty"` // For delayed jobs
}

// AIAnalysisJob represents a job for AI analysis
type AIAnalysisJob struct {
	Job
	Symbol    string   `json:"symbol"`
	Agents    []string `json:"agents"`
	UserID    int      `json:"user_id"`
	RequestID string   `json:"request_id"`
}

// MarketDataUpdateJob represents a job for updating market data
type MarketDataUpdateJob struct {
	Job
	Symbols   []string `json:"symbols"`
	DataType  string   `json:"data_type"` // "prices", "news", "technicals"
	Source    string   `json:"source"`
	Immediate bool     `json:"immediate"`  // Skip rate limiting
}

// RiskCalculationJob represents a job for calculating risk metrics
type RiskCalculationJob struct {
	Job
	UserID      int      `json:"user_id"`
	PortfolioID int      `json:"portfolio_id"`
	Symbols     []string `json:"symbols"`
	RiskType    string   `json:"risk_type"` // "position", "portfolio", "var"
}

// NotificationJob represents a job for sending notifications
type NotificationJob struct {
	Job
	UserID   int                    `json:"user_id"`
	Type     string                 `json:"notification_type"` // "email", "slack", "webhook"
	Subject  string                 `json:"subject"`
	Message  string                 `json:"message"`
	Data     map[string]interface{} `json:"data"`
	Channels []string               `json:"channels"`
}

// ReportGenerationJob represents a job for generating reports
type ReportGenerationJob struct {
	Job
	UserID      int       `json:"user_id"`
	PortfolioID int       `json:"portfolio_id"`
	ReportType  string    `json:"report_type"` // "performance", "risk", "positions"
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Format      string    `json:"format"` // "pdf", "csv", "json"
	Recipients  []string  `json:"recipients"`
}

// JobStatus represents the status of a job execution
type JobStatus struct {
	JobID       string                 `json:"job_id"`
	Status      string                 `json:"status"` // "pending", "running", "completed", "failed"
	Progress    float64                `json:"progress"` // 0-100
	Message     string                 `json:"message"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    *time.Duration         `json:"duration,omitempty"`
}

// Queue constants
const (
	// High priority queues
	QueueAIAnalysis   = "queue:ai_analysis"
	QueueRiskCalc     = "queue:risk_calculation"
	QueueNotifications = "queue:notifications"

	// Medium priority queues
	QueueMarketData   = "queue:market_data"
	QueueReports      = "queue:reports"

	// Low priority queues
	QueueCleanup      = "queue:cleanup"
	QueueMaintenance  = "queue:maintenance"

	// Job types
	JobTypeAIAnalysis      = "ai_analysis"
	JobTypeMarketDataUpdate = "market_data_update"
	JobTypeRiskCalculation = "risk_calculation"
	JobTypeNotification    = "notification"
	JobTypeReportGeneration = "report_generation"
	JobTypeCleanup         = "cleanup"

	// Job statuses
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
	JobStatusRetrying  = "retrying"
)

// Event models for pub/sub
type Event struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// PriceUpdateEvent represents a real-time price update
type PriceUpdateEvent struct {
	Event
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Change float64 `json:"change"`
	Volume int64   `json:"volume"`
}

// TradeExecutedEvent represents a trade execution
type TradeExecutedEvent struct {
	Event
	TradeID  int     `json:"trade_id"`
	UserID   int     `json:"user_id"`
	Symbol   string  `json:"symbol"`
	Quantity int64   `json:"quantity"`
	Price    float64 `json:"price"`
	Side     string  `json:"side"`
}

// RiskAlertEvent represents a risk alert
type RiskAlertEvent struct {
	Event
	AlertID   int     `json:"alert_id"`
	UserID    int     `json:"user_id"`
	AlertType string  `json:"alert_type"`
	Severity  string  `json:"severity"`
	Symbol    string  `json:"symbol"`
	Message   string  `json:"message"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
}

// AISignalEvent represents an AI signal generation
type AISignalEvent struct {
	Event
	SignalID   int     `json:"signal_id"`
	AgentName  string  `json:"agent_name"`
	Symbol     string  `json:"symbol"`
	Signal     string  `json:"signal"`
	Confidence float64 `json:"confidence"`
	Price      float64 `json:"price"`
}

// Event channels for pub/sub
const (
	ChannelPriceUpdates = "events:price_updates"
	ChannelTradeEvents  = "events:trades"
	ChannelRiskAlerts   = "events:risk_alerts"
	ChannelAISignals    = "events:ai_signals"
	ChannelSystemEvents = "events:system"
)