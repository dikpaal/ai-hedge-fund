package models

import "time"

// AISignal represents an AI agent's investment recommendation
type AISignal struct {
	ID         int       `json:"id"`
	AgentName  string    `json:"agent_name"`  // "warren_buffett", "michael_burry", etc.
	Symbol     string    `json:"symbol"`
	Signal     string    `json:"signal"`      // "buy", "sell", "hold"
	Confidence float64   `json:"confidence"`  // 0-100
	Reasoning  string    `json:"reasoning"`
	Price      float64   `json:"price"`       // Price at time of signal
	CreatedAt  time.Time `json:"created_at"`
}

// AIAnalysisRequest represents a request for AI analysis
type AIAnalysisRequest struct {
	Symbol    string            `json:"symbol"`
	Agents    []string          `json:"agents"`    // List of agent names to run
	StartDate *time.Time        `json:"start_date,omitempty"`
	EndDate   *time.Time        `json:"end_date,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"` // Additional options
}

// AIAnalysisResponse represents the response from AI analysis
type AIAnalysisResponse struct {
	RequestID      string            `json:"request_id"`
	Symbol         string            `json:"symbol"`
	Signals        []AISignal        `json:"signals"`
	ConsensusSignal string           `json:"consensus_signal"` // Overall consensus
	ConsensusConfidence float64      `json:"consensus_confidence"`
	MarketData     *MarketData       `json:"market_data,omitempty"`
	RiskMetrics    *RiskMetrics      `json:"risk_metrics,omitempty"`
	ProcessingTime float64           `json:"processing_time_ms"`
	CompletedAt    time.Time         `json:"completed_at"`
}

// AgentConfig represents configuration for an AI agent
type AgentConfig struct {
	Name            string                 `json:"name"`
	DisplayName     string                 `json:"display_name"`
	Description     string                 `json:"description"`
	InvestingStyle  string                 `json:"investing_style"`
	Enabled         bool                   `json:"enabled"`
	Parameters      map[string]interface{} `json:"parameters"`
	ModelProvider   string                 `json:"model_provider"`   // "openai", "anthropic", etc.
	ModelName       string                 `json:"model_name"`       // "gpt-4", "claude-3", etc.
	Temperature     float64                `json:"temperature"`
	MaxTokens       int                    `json:"max_tokens"`
}

// AgentPerformance tracks how well an agent's signals perform
type AgentPerformance struct {
	ID            int       `json:"id" db:"id"`
	AgentName     string    `json:"agent_name" db:"agent_name"`
	Symbol        string    `json:"symbol" db:"symbol"`
	Period        string    `json:"period" db:"period"`        // "1d", "1w", "1m", "3m", "1y"
	TotalSignals  int       `json:"total_signals" db:"total_signals"`
	CorrectSignals int      `json:"correct_signals" db:"correct_signals"`
	Accuracy      float64   `json:"accuracy" db:"accuracy"`    // % of correct signals
	AvgReturn     float64   `json:"avg_return" db:"avg_return"` // Average return per signal
	SharpeRatio   float64   `json:"sharpe_ratio" db:"sharpe_ratio"`
	MaxDrawdown   float64   `json:"max_drawdown" db:"max_drawdown"`
	LastUpdated   time.Time `json:"last_updated" db:"last_updated"`
}

// WorkflowStatus represents the status of an AI workflow execution
type WorkflowStatus struct {
	RequestID       string                 `json:"request_id"`
	Status          string                 `json:"status"`          // "pending", "running", "completed", "failed"
	CurrentStep     string                 `json:"current_step"`
	CompletedSteps  []string               `json:"completed_steps"`
	Progress        float64                `json:"progress"`        // 0-100
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Result          *AIAnalysisResponse    `json:"result,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
}

// AIAgentMetrics represents performance metrics for an AI agent
type AIAgentMetrics struct {
	AgentName       string    `json:"agent_name"`
	TotalRequests   int       `json:"total_requests"`
	SuccessfulRequests int    `json:"successful_requests"`
	FailedRequests  int       `json:"failed_requests"`
	AvgResponseTime float64   `json:"avg_response_time_ms"`
	AvgConfidence   float64   `json:"avg_confidence"`
	LastRequest     time.Time `json:"last_request"`
	LastSuccess     time.Time `json:"last_success"`
}