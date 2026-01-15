package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"hedge-fund/internal/portfolio/domain"
	"hedge-fund/internal/portfolio/handlers"
	"hedge-fund/internal/portfolio/repository"
	"hedge-fund/internal/portfolio/service"
	"hedge-fund/pkg/shared/config"
	"hedge-fund/pkg/shared/database"
	"hedge-fund/pkg/shared/logger"
	"hedge-fund/pkg/shared/redis"
)

// PortfolioIntegrationTestSuite holds test dependencies
type PortfolioIntegrationTestSuite struct {
	suite.Suite
	db          *database.DB
	redisClient *redis.Client
	router      *gin.Engine
	service     *service.PortfolioService
	testUserID  int
}

// SetupSuite runs once before all tests
func (suite *PortfolioIntegrationTestSuite) SetupSuite() {
	// Set test environment
	os.Setenv("ENV", "test")
	os.Setenv("DATABASE_URL", "postgres://hedge_fund:password@localhost:5433/hedge_fund_test?sslmode=disable")
	os.Setenv("LOG_LEVEL", "error") // Reduce log noise in tests

	// Initialize logger
	err := logger.Init("error", "test")
	suite.Require().NoError(err)

	// Load test configuration
	cfg := config.Load()

	// Connect to test database
	db, err := database.Connect(cfg)
	suite.Require().NoError(err)
	suite.db = db

	// Connect to Redis
	redisClient, err := redis.Connect(cfg)
	suite.Require().NoError(err)
	suite.redisClient = redisClient

	// Get test user ID
	suite.testUserID = suite.getTestUserID()
}

// SetupTest runs before each test
func (suite *PortfolioIntegrationTestSuite) SetupTest() {
	// Clean database tables (in correct order due to foreign keys)
	suite.cleanDatabase()

	// Flush Redis cache
	suite.redisClient.FlushCache(context.Background())

	// Setup dependencies
	portfolioRepo := repository.NewPortfolioRepository(suite.db, logger.Logger)
	domainService := domain.NewPortfolioService()
	portfolioService := service.NewPortfolioService(portfolioRepo, domainService, logger.Logger)
	marketClient := handlers.NewMockMarketDataClient()
	portfolioHandler := handlers.NewPortfolioHandler(portfolioService, marketClient, logger.Logger)

	suite.service = portfolioService

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/portfolios", portfolioHandler.CreatePortfolio)
		v1.GET("/portfolios/:id", portfolioHandler.GetPortfolio)
		v1.PUT("/portfolios/:id", portfolioHandler.UpdatePortfolio)
		v1.DELETE("/portfolios/:id", portfolioHandler.DeletePortfolio)
		v1.GET("/portfolios/user/:user_id", portfolioHandler.ListUserPortfolios)
		v1.GET("/portfolios/:id/positions", portfolioHandler.GetPositions)
		v1.GET("/portfolios/:id/summary", portfolioHandler.GetSummary)
		v1.GET("/portfolios/:id/allocation", portfolioHandler.GetAllocation)
		v1.GET("/portfolios/:id/risk", portfolioHandler.GetRiskMetrics)
		v1.POST("/portfolios/:id/trades", portfolioHandler.ExecuteTrade)
		v1.GET("/portfolios/:id/trades", portfolioHandler.GetTradeHistory)
		v1.POST("/portfolios/:id/rebalance", portfolioHandler.GetRebalanceRecommendations)
	}

	suite.router = router
}

// TearDownSuite runs once after all tests
func (suite *PortfolioIntegrationTestSuite) TearDownSuite() {
	suite.db.Close()
	suite.redisClient.Close()
}

// Helper methods

func (suite *PortfolioIntegrationTestSuite) getTestUserID() int {
	var userID int
	query := "SELECT id FROM users WHERE username = 'testuser' LIMIT 1"
	err := suite.db.QueryRowContext(context.Background(), query).Scan(&userID)
	suite.Require().NoError(err)
	return userID
}

func (suite *PortfolioIntegrationTestSuite) cleanDatabase() {
	ctx := context.Background()
	suite.db.ExecContext(ctx, "DELETE FROM trades")
	suite.db.ExecContext(ctx, "DELETE FROM positions")
	suite.db.ExecContext(ctx, "DELETE FROM portfolios")
}

func (suite *PortfolioIntegrationTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// Test Cases

func (suite *PortfolioIntegrationTestSuite) TestCreatePortfolio() {
	reqBody := handlers.CreatePortfolioRequest{
		UserID:      suite.testUserID,
		Name:        "Test Portfolio",
		InitialCash: 100000.00,
	}

	w := suite.makeRequest("POST", "/api/v1/portfolios", reqBody)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response handlers.PortfolioResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testUserID, response.UserID)
	assert.Equal(suite.T(), "Test Portfolio", response.Name)
	assert.Equal(suite.T(), 100000.00, response.Cash)
	assert.NotZero(suite.T(), response.ID)
}

func (suite *PortfolioIntegrationTestSuite) TestGetPortfolio() {
	// Create test portfolio
	portfolio, err := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "My Portfolio", 50000.00)
	suite.Require().NoError(err)

	// Get portfolio
	path := fmt.Sprintf("/api/v1/portfolios/%d", portfolio.ID)
	w := suite.makeRequest("GET", path, nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.PortfolioResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), portfolio.ID, response.ID)
	assert.Equal(suite.T(), "My Portfolio", response.Name)
	assert.Equal(suite.T(), 50000.00, response.Cash)
}

func (suite *PortfolioIntegrationTestSuite) TestExecuteTradeBuy() {
	portfolio, _ := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "Trading Portfolio", 100000.00)

	tradeReq := handlers.TradeRequest{
		Symbol:    "AAPL",
		Side:      "buy",
		Quantity:  10,
		OrderType: "market",
	}

	path := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
	w := suite.makeRequest("POST", path, tradeReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.TradeResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "AAPL", response.Symbol)
	assert.Equal(suite.T(), "filled", response.Status)
	assert.NotZero(suite.T(), response.Price)
	assert.Equal(suite.T(), int64(10), response.Quantity)
}

func (suite *PortfolioIntegrationTestSuite) TestExecuteTradeSell() {
	portfolio, _ := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "Trading Portfolio", 100000.00)

	// First buy shares
	buyReq := handlers.TradeRequest{
		Symbol:    "AAPL",
		Side:      "buy",
		Quantity:  10,
		OrderType: "market",
	}
	path := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
	suite.makeRequest("POST", path, buyReq)

	// Now sell
	sellReq := handlers.TradeRequest{
		Symbol:    "AAPL",
		Side:      "sell",
		Quantity:  5,
		OrderType: "market",
	}
	w := suite.makeRequest("POST", path, sellReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.TradeResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "sell", response.Side)
	assert.Equal(suite.T(), int64(5), response.Quantity)
}

func (suite *PortfolioIntegrationTestSuite) TestGetSummary() {
	portfolio, _ := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "Summary Portfolio", 100000.00)

	// Execute a trade
	tradeReq := handlers.TradeRequest{
		Symbol:    "AAPL",
		Side:      "buy",
		Quantity:  10,
		OrderType: "market",
	}
	tradePath := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
	suite.makeRequest("POST", tradePath, tradeReq)

	// Get summary
	path := fmt.Sprintf("/api/v1/portfolios/%d/summary", portfolio.ID)
	w := suite.makeRequest("GET", path, nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.SummaryResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Greater(suite.T(), response.TotalValue, 0.0)
	assert.Equal(suite.T(), 1, response.PositionCount)
	assert.Greater(suite.T(), response.PositionsValue, 0.0)
}

func (suite *PortfolioIntegrationTestSuite) TestGetPositions() {
	portfolio, _ := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "Positions Portfolio", 100000.00)

	// Create multiple positions
	symbols := []string{"AAPL", "GOOGL"}
	for _, symbol := range symbols {
		tradeReq := handlers.TradeRequest{
			Symbol:    symbol,
			Side:      "buy",
			Quantity:  10,
			OrderType: "market",
		}
		tradePath := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
		suite.makeRequest("POST", tradePath, tradeReq)
	}

	path := fmt.Sprintf("/api/v1/portfolios/%d/positions", portfolio.ID)
	w := suite.makeRequest("GET", path, nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var positions []handlers.PositionResponse
	json.Unmarshal(w.Body.Bytes(), &positions)
	assert.Len(suite.T(), positions, 2)
}

func (suite *PortfolioIntegrationTestSuite) TestGetTradeHistory() {
	portfolio, _ := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "History Portfolio", 100000.00)

	// Execute multiple trades
	for i := 0; i < 3; i++ {
		tradeReq := handlers.TradeRequest{
			Symbol:    "AAPL",
			Side:      "buy",
			Quantity:  1,
			OrderType: "market",
		}
		tradePath := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
		suite.makeRequest("POST", tradePath, tradeReq)
	}

	path := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
	w := suite.makeRequest("GET", path, nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var trades []handlers.TradeResponse
	json.Unmarshal(w.Body.Bytes(), &trades)
	assert.GreaterOrEqual(suite.T(), len(trades), 3)
}

func (suite *PortfolioIntegrationTestSuite) TestGetAllocation() {
	portfolio, _ := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "Allocation Portfolio", 100000.00)

	// Create diversified portfolio
	trades := []struct {
		symbol   string
		quantity int64
	}{
		{"AAPL", 10},
		{"GOOGL", 5},
		{"MSFT", 8},
	}

	for _, trade := range trades {
		tradeReq := handlers.TradeRequest{
			Symbol:    trade.symbol,
			Side:      "buy",
			Quantity:  trade.quantity,
			OrderType: "market",
		}
		tradePath := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
		suite.makeRequest("POST", tradePath, tradeReq)
	}

	path := fmt.Sprintf("/api/v1/portfolios/%d/allocation", portfolio.ID)
	w := suite.makeRequest("GET", path, nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var allocations []handlers.AllocationResponse
	json.Unmarshal(w.Body.Bytes(), &allocations)
	assert.NotEmpty(suite.T(), allocations)

	// Verify percentages add up to ~100
	totalPercent := 0.0
	for _, alloc := range allocations {
		totalPercent += alloc.Percentage
	}
	assert.InDelta(suite.T(), 100.0, totalPercent, 1.0)
}

func (suite *PortfolioIntegrationTestSuite) TestGetRiskMetrics() {
	portfolio, _ := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "Risk Portfolio", 100000.00)

	// Create position
	tradeReq := handlers.TradeRequest{
		Symbol:    "AAPL",
		Side:      "buy",
		Quantity:  100,
		OrderType: "market",
	}
	tradePath := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
	suite.makeRequest("POST", tradePath, tradeReq)

	path := fmt.Sprintf("/api/v1/portfolios/%d/risk", portfolio.ID)
	w := suite.makeRequest("GET", path, nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var metrics handlers.RiskMetricsResponse
	json.Unmarshal(w.Body.Bytes(), &metrics)
	assert.Greater(suite.T(), metrics.TotalValue, 0.0)
	assert.Equal(suite.T(), 1, metrics.PositionCount)
	assert.Greater(suite.T(), metrics.MaxPositionPercent, 0.0)
}

func (suite *PortfolioIntegrationTestSuite) TestInsufficientFunds() {
	portfolio, _ := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "Low Cash Portfolio", 1000.00)

	tradeReq := handlers.TradeRequest{
		Symbol:    "AAPL",
		Side:      "buy",
		Quantity:  1000, // Too many shares
		OrderType: "market",
	}

	path := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
	w := suite.makeRequest("POST", path, tradeReq)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errResponse handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResponse)
	assert.Contains(suite.T(), errResponse.Error, "Failed to execute trade")
}

func (suite *PortfolioIntegrationTestSuite) TestInsufficientShares() {
	portfolio, _ := suite.service.CreatePortfolio(context.Background(), suite.testUserID, "Empty Portfolio", 100000.00)

	sellReq := handlers.TradeRequest{
		Symbol:    "AAPL",
		Side:      "sell",
		Quantity:  10, // Don't own any shares
		OrderType: "market",
	}

	path := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolio.ID)
	w := suite.makeRequest("POST", path, sellReq)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *PortfolioIntegrationTestSuite) TestEndToEndTradeFlow() {
	// Create portfolio
	createReq := handlers.CreatePortfolioRequest{
		UserID:      suite.testUserID,
		Name:        "E2E Portfolio",
		InitialCash: 100000.00,
	}
	w := suite.makeRequest("POST", "/api/v1/portfolios", createReq)
	var portfolioResp handlers.PortfolioResponse
	json.Unmarshal(w.Body.Bytes(), &portfolioResp)
	portfolioID := portfolioResp.ID

	// Buy shares
	buyReq := handlers.TradeRequest{Symbol: "AAPL", Side: "buy", Quantity: 10, OrderType: "market"}
	tradePath := fmt.Sprintf("/api/v1/portfolios/%d/trades", portfolioID)
	w = suite.makeRequest("POST", tradePath, buyReq)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Check portfolio updated
	portfolioPath := fmt.Sprintf("/api/v1/portfolios/%d", portfolioID)
	w = suite.makeRequest("GET", portfolioPath, nil)
	json.Unmarshal(w.Body.Bytes(), &portfolioResp)
	assert.Less(suite.T(), portfolioResp.Cash, 100000.00) // Cash reduced

	// Check positions
	positionsPath := fmt.Sprintf("/api/v1/portfolios/%d/positions", portfolioID)
	w = suite.makeRequest("GET", positionsPath, nil)
	var positions []handlers.PositionResponse
	json.Unmarshal(w.Body.Bytes(), &positions)
	assert.Len(suite.T(), positions, 1)
	assert.Equal(suite.T(), "AAPL", positions[0].Symbol)
	assert.Equal(suite.T(), int64(10), positions[0].Quantity)

	// Sell partial shares
	sellReq := handlers.TradeRequest{Symbol: "AAPL", Side: "sell", Quantity: 5, OrderType: "market"}
	w = suite.makeRequest("POST", tradePath, sellReq)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify position updated
	w = suite.makeRequest("GET", positionsPath, nil)
	json.Unmarshal(w.Body.Bytes(), &positions)
	assert.Equal(suite.T(), int64(5), positions[0].Quantity)

	// Check trade history
	w = suite.makeRequest("GET", tradePath, nil)
	var trades []handlers.TradeResponse
	json.Unmarshal(w.Body.Bytes(), &trades)
	assert.GreaterOrEqual(suite.T(), len(trades), 2) // Buy + Sell
}

// TestMain is the entry point for tests
func TestPortfolioIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PortfolioIntegrationTestSuite))
}
