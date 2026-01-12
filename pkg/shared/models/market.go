package models

import "time"

// Price represents market price data
type Price struct {
	Symbol    string    `json:"symbol" db:"symbol"`
	Open      float64   `json:"open" db:"open"`
	High      float64   `json:"high" db:"high"`
	Low       float64   `json:"low" db:"low"`
	Close     float64   `json:"close" db:"close"`
	Volume    int64     `json:"volume" db:"volume"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Source    string    `json:"source" db:"source"` // API source identifier
}

// Quote represents real-time quote data
type Quote struct {
	Symbol    string    `json:"symbol"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	BidSize   int64     `json:"bid_size"`
	AskSize   int64     `json:"ask_size"`
	Last      float64   `json:"last"`
	Volume    int64     `json:"volume"`
	Change    float64   `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Timestamp time.Time `json:"timestamp"`
}

// NewsItem represents financial news
type NewsItem struct {
	ID          string    `json:"id" db:"id"`
	Symbol      string    `json:"symbol" db:"symbol"`
	Title       string    `json:"title" db:"title"`
	Summary     string    `json:"summary" db:"summary"`
	URL         string    `json:"url" db:"url"`
	Source      string    `json:"source" db:"source"`
	Sentiment   string    `json:"sentiment" db:"sentiment"` // "positive", "negative", "neutral"
	SentimentScore float64 `json:"sentiment_score" db:"sentiment_score"` // -1.0 to 1.0
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// MarketData aggregates various market data for a symbol
type MarketData struct {
	Symbol        string     `json:"symbol"`
	CurrentPrice  float64    `json:"current_price"`
	Quote         *Quote     `json:"quote,omitempty"`
	DailyBar      *Price     `json:"daily_bar,omitempty"`
	Volume        int64      `json:"volume"`
	MarketCap     float64    `json:"market_cap,omitempty"`
	PERatio       float64    `json:"pe_ratio,omitempty"`
	DividendYield float64    `json:"dividend_yield,omitempty"`
	Beta          float64    `json:"beta,omitempty"`
	AvgVolume     int64      `json:"avg_volume,omitempty"`
	RecentNews    []NewsItem `json:"recent_news,omitempty"`
	LastUpdated   time.Time  `json:"last_updated"`
}

// TechnicalIndicators represents calculated technical analysis indicators
type TechnicalIndicators struct {
	Symbol         string    `json:"symbol"`
	SMA20          float64   `json:"sma_20"`          // 20-period Simple Moving Average
	SMA50          float64   `json:"sma_50"`          // 50-period Simple Moving Average
	SMA200         float64   `json:"sma_200"`         // 200-period Simple Moving Average
	EMA20          float64   `json:"ema_20"`          // 20-period Exponential Moving Average
	RSI            float64   `json:"rsi"`             // Relative Strength Index
	MACD           float64   `json:"macd"`            // MACD Line
	MACDSignal     float64   `json:"macd_signal"`     // MACD Signal Line
	MACDHistogram  float64   `json:"macd_histogram"`  // MACD Histogram
	BollingerUpper float64   `json:"bollinger_upper"` // Upper Bollinger Band
	BollingerLower float64   `json:"bollinger_lower"` // Lower Bollinger Band
	BollingerMid   float64   `json:"bollinger_mid"`   // Middle Bollinger Band
	ATR            float64   `json:"atr"`             // Average True Range
	StochK         float64   `json:"stoch_k"`         // Stochastic %K
	StochD         float64   `json:"stoch_d"`         // Stochastic %D
	WilliamsR      float64   `json:"williams_r"`      // Williams %R
	CalculatedAt   time.Time `json:"calculated_at"`
}

// WatchlistItem represents a symbol in a user's watchlist
type WatchlistItem struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Symbol       string    `json:"symbol" db:"symbol"`
	Name         string    `json:"name" db:"name"`
	CurrentPrice float64   `json:"current_price"`
	Change       float64   `json:"change"`
	ChangePercent float64  `json:"change_percent"`
	AlertPrice   *float64  `json:"alert_price" db:"alert_price"`
	AlertEnabled bool      `json:"alert_enabled" db:"alert_enabled"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// MarketIndex represents major market indices
type MarketIndex struct {
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	Value         float64   `json:"value"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
	LastUpdated   time.Time `json:"last_updated"`
}