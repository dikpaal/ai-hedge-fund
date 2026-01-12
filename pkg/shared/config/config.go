package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	// Database
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	RedisURL    string `mapstructure:"REDIS_URL"`

	// API Keys
	OpenAIAPIKey              string `mapstructure:"OPENAI_API_KEY"`
	FinancialDatasetsAPIKey   string `mapstructure:"FINANCIAL_DATASETS_API_KEY"`
	AnthropicAPIKey           string `mapstructure:"ANTHROPIC_API_KEY"`

	// Service Ports
	APIGatewayPort      string `mapstructure:"API_GATEWAY_PORT"`
	PortfolioServicePort string `mapstructure:"PORTFOLIO_SERVICE_PORT"`
	RiskServicePort     string `mapstructure:"RISK_SERVICE_PORT"`
	MarketDataServicePort string `mapstructure:"MARKET_DATA_SERVICE_PORT"`
	AIServicePort       string `mapstructure:"AI_SERVICE_PORT"`

	// JWT
	JWTSecret string `mapstructure:"JWT_SECRET"`

	// Application
	LogLevel string `mapstructure:"LOG_LEVEL"`
	Env      string `mapstructure:"ENV"`

	// Monitoring
	PrometheusPort string `mapstructure:"PROMETHEUS_PORT"`
	GrafanaPort    string `mapstructure:"GRAFANA_PORT"`
	JaegerPort     string `mapstructure:"JAEGER_PORT"`
}

func Load() *Config {
	config := &Config{}

	// Set default values
	viper.SetDefault("DATABASE_URL", "postgres://hedge_fund:password@localhost:5432/hedge_fund_db?sslmode=disable")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	viper.SetDefault("API_GATEWAY_PORT", "8080")
	viper.SetDefault("PORTFOLIO_SERVICE_PORT", "8081")
	viper.SetDefault("RISK_SERVICE_PORT", "8082")
	viper.SetDefault("MARKET_DATA_SERVICE_PORT", "8083")
	viper.SetDefault("AI_SERVICE_PORT", "8084")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("PROMETHEUS_PORT", "9090")
	viper.SetDefault("GRAFANA_PORT", "3000")
	viper.SetDefault("JAEGER_PORT", "16686")

	// Read config from environment variables
	viper.AutomaticEnv()

	// Try to read from .env file if it exists
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %v", err)
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}

	// Validate required configuration
	if config.Env == "production" {
		validateProductionConfig(config)
	}

	return config
}

func validateProductionConfig(config *Config) {
	required := map[string]string{
		"DATABASE_URL": config.DatabaseURL,
		"REDIS_URL":    config.RedisURL,
		"JWT_SECRET":   config.JWTSecret,
	}

	for key, value := range required {
		if value == "" {
			log.Fatalf("Required configuration %s is not set", key)
			os.Exit(1)
		}
	}
}