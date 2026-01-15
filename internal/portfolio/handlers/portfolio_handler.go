package handlers

import (
	"net/http"
	"strconv"

	"hedge-fund/internal/portfolio/service"
	"hedge-fund/pkg/shared/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PortfolioHandler struct {
	service      *service.PortfolioService
	marketClient MarketDataClient
	logger       *zap.Logger
}

// MarketDataClient interface for getting market prices
type MarketDataClient interface {
	GetCurrentPrice(symbol string) (float64, error)
	GetCurrentPrices(symbols []string) (map[string]float64, error)
}

func NewPortfolioHandler(service *service.PortfolioService, marketClient MarketDataClient, logger *zap.Logger) *PortfolioHandler {
	return &PortfolioHandler{
		service:      service,
		marketClient: marketClient,
		logger:       logger,
	}
}

// CreatePortfolio godoc
// @Summary Create a new portfolio
// @Description Create a new portfolio for a user with initial cash
// @Tags portfolios
// @Accept json
// @Produce json
// @Param request body CreatePortfolioRequest true "Create Portfolio Request"
// @Success 201 {object} PortfolioResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios [post]
func (h *PortfolioHandler) CreatePortfolio(c *gin.Context) {
	var req CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Details: err.Error()})
		return
	}

	portfolio, err := h.service.CreatePortfolio(c.Request.Context(), req.UserID, req.Name, req.InitialCash)
	if err != nil {
		h.logger.Error("Failed to create portfolio", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create portfolio", Details: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, h.toPortfolioResponse(portfolio))
}

// GetPortfolio godoc
// @Summary Get portfolio by ID
// @Description Get portfolio details including positions
// @Tags portfolios
// @Produce json
// @Param id path int true "Portfolio ID"
// @Success 200 {object} PortfolioResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/portfolios/{id} [get]
func (h *PortfolioHandler) GetPortfolio(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	portfolio, err := h.service.GetPortfolio(c.Request.Context(), portfolioID)
	if err != nil {
		h.logger.Error("Failed to get portfolio", zap.Error(err), zap.Int("portfolio_id", portfolioID))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Portfolio not found", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.toPortfolioResponse(portfolio))
}

// UpdatePortfolio godoc
// @Summary Update portfolio
// @Description Update portfolio cash balance
// @Tags portfolios
// @Accept json
// @Produce json
// @Param id path int true "Portfolio ID"
// @Param request body UpdatePortfolioRequest true "Update Portfolio Request"
// @Success 200 {object} PortfolioResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/{id} [put]
func (h *PortfolioHandler) UpdatePortfolio(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	var req UpdatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Details: err.Error()})
		return
	}

	portfolio, err := h.service.GetPortfolio(c.Request.Context(), portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Portfolio not found"})
		return
	}

	portfolio.Cash = req.Cash
	if err := h.service.UpdatePortfolio(c.Request.Context(), portfolio); err != nil {
		h.logger.Error("Failed to update portfolio", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update portfolio", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.toPortfolioResponse(portfolio))
}

// DeletePortfolio godoc
// @Summary Delete portfolio
// @Description Delete a portfolio and all its positions
// @Tags portfolios
// @Param id path int true "Portfolio ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/{id} [delete]
func (h *PortfolioHandler) DeletePortfolio(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	if err := h.service.DeletePortfolio(c.Request.Context(), portfolioID); err != nil {
		h.logger.Error("Failed to delete portfolio", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete portfolio", Details: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListUserPortfolios godoc
// @Summary List user portfolios
// @Description Get all portfolios for a user
// @Tags portfolios
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {array} PortfolioResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/user/{user_id} [get]
func (h *PortfolioHandler) ListUserPortfolios(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	portfolios, err := h.service.GetUserPortfolios(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list portfolios", zap.Error(err), zap.Int("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list portfolios", Details: err.Error()})
		return
	}

	response := make([]PortfolioResponse, len(portfolios))
	for i, portfolio := range portfolios {
		response[i] = h.toPortfolioResponse(&portfolio)
	}

	c.JSON(http.StatusOK, response)
}

// GetPositions godoc
// @Summary Get portfolio positions
// @Description Get all positions for a portfolio
// @Tags portfolios
// @Produce json
// @Param id path int true "Portfolio ID"
// @Success 200 {array} PositionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/{id}/positions [get]
func (h *PortfolioHandler) GetPositions(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	positions, err := h.service.GetPositions(c.Request.Context(), portfolioID)
	if err != nil {
		h.logger.Error("Failed to get positions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get positions", Details: err.Error()})
		return
	}

	response := make([]PositionResponse, len(positions))
	for i, pos := range positions {
		response[i] = h.toPositionResponse(&pos)
	}

	c.JSON(http.StatusOK, response)
}

// GetSummary godoc
// @Summary Get portfolio summary
// @Description Get portfolio summary with current market prices
// @Tags portfolios
// @Produce json
// @Param id path int true "Portfolio ID"
// @Success 200 {object} SummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/{id}/summary [get]
func (h *PortfolioHandler) GetSummary(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	// Get portfolio
	portfolio, err := h.service.GetPortfolio(c.Request.Context(), portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Portfolio not found"})
		return
	}

	// Get current prices for all positions
	symbols := make([]string, len(portfolio.Positions))
	for i, pos := range portfolio.Positions {
		symbols[i] = pos.Symbol
	}

	currentPrices, err := h.marketClient.GetCurrentPrices(symbols)
	if err != nil {
		h.logger.Error("Failed to get current prices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get market prices"})
		return
	}

	// For now, use empty previous day prices (will be implemented with Market Data Service)
	previousDayPrices := make(map[string]float64)

	summary, err := h.service.CalculatePortfolioSummary(c.Request.Context(), portfolioID, currentPrices, previousDayPrices)
	if err != nil {
		h.logger.Error("Failed to calculate summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate summary", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.toSummaryResponse(summary))
}

// ExecuteTrade godoc
// @Summary Execute trade
// @Description Execute a buy or sell trade order
// @Tags portfolios
// @Accept json
// @Produce json
// @Param id path int true "Portfolio ID"
// @Param request body TradeRequest true "Trade Request"
// @Success 200 {object} TradeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/{id}/trades [post]
func (h *PortfolioHandler) ExecuteTrade(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	var req TradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Details: err.Error()})
		return
	}

	// Get portfolio to get user_id
	portfolio, err := h.service.GetPortfolio(c.Request.Context(), portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Portfolio not found"})
		return
	}

	// Get current price from market data
	currentPrice := req.Price
	if req.OrderType == "market" {
		currentPrice, err = h.marketClient.GetCurrentPrice(req.Symbol)
		if err != nil {
			h.logger.Error("Failed to get current price", zap.Error(err), zap.String("symbol", req.Symbol))
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get market price", Details: err.Error()})
			return
		}
	}

	// Create trade object
	trade := &models.Trade{
		UserID:   portfolio.UserID,
		Symbol:   req.Symbol,
		Quantity: req.Quantity,
		Side:     req.Side,
		Type:     req.OrderType,
		Status:   "pending",
	}

	// Execute trade
	position, err := h.service.ExecuteTrade(c.Request.Context(), portfolioID, trade, currentPrice)
	if err != nil {
		h.logger.Error("Failed to execute trade", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to execute trade", Details: err.Error()})
		return
	}

	h.logger.Info("Trade executed successfully",
		zap.Int("portfolio_id", portfolioID),
		zap.String("symbol", req.Symbol),
		zap.String("side", req.Side),
		zap.Int64("quantity", req.Quantity),
		zap.Float64("price", currentPrice))

	c.JSON(http.StatusOK, h.toTradeResponse(trade, position))
}

// GetTradeHistory godoc
// @Summary Get trade history
// @Description Get trade history for a portfolio
// @Tags portfolios
// @Produce json
// @Param id path int true "Portfolio ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} TradeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/{id}/trades [get]
func (h *PortfolioHandler) GetTradeHistory(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	// Get portfolio to get user_id
	portfolio, err := h.service.GetPortfolio(c.Request.Context(), portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Portfolio not found"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	trades, err := h.service.GetTradeHistory(c.Request.Context(), portfolio.UserID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get trade history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get trade history", Details: err.Error()})
		return
	}

	response := make([]TradeResponse, len(trades))
	for i, trade := range trades {
		response[i] = h.toTradeResponse(&trade, nil)
	}

	c.JSON(http.StatusOK, response)
}

// GetAllocation godoc
// @Summary Get portfolio allocation
// @Description Get portfolio allocation percentages
// @Tags portfolios
// @Produce json
// @Param id path int true "Portfolio ID"
// @Success 200 {array} AllocationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/{id}/allocation [get]
func (h *PortfolioHandler) GetAllocation(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	// Get portfolio
	portfolio, err := h.service.GetPortfolio(c.Request.Context(), portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Portfolio not found"})
		return
	}

	// Get current prices
	symbols := make([]string, len(portfolio.Positions))
	for i, pos := range portfolio.Positions {
		symbols[i] = pos.Symbol
	}

	currentPrices, err := h.marketClient.GetCurrentPrices(symbols)
	if err != nil {
		h.logger.Error("Failed to get current prices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get market prices"})
		return
	}

	allocations, err := h.service.GetPortfolioAllocation(c.Request.Context(), portfolioID, currentPrices)
	if err != nil {
		h.logger.Error("Failed to get allocation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get allocation", Details: err.Error()})
		return
	}

	// Calculate total value
	totalValue := portfolio.Cash
	for _, pos := range portfolio.Positions {
		if price, ok := currentPrices[pos.Symbol]; ok {
			totalValue += float64(pos.Quantity) * price
		}
	}

	// Convert to response
	response := make([]AllocationResponse, 0, len(allocations))
	for symbol, percentage := range allocations {
		value := (percentage / 100) * totalValue
		response = append(response, AllocationResponse{
			Symbol:     symbol,
			Percentage: percentage,
			Value:      value,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetRiskMetrics godoc
// @Summary Get risk metrics
// @Description Get portfolio risk metrics
// @Tags portfolios
// @Produce json
// @Param id path int true "Portfolio ID"
// @Success 200 {object} RiskMetricsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/{id}/risk [get]
func (h *PortfolioHandler) GetRiskMetrics(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	// Get portfolio
	portfolio, err := h.service.GetPortfolio(c.Request.Context(), portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Portfolio not found"})
		return
	}

	// Get current prices
	symbols := make([]string, len(portfolio.Positions))
	for i, pos := range portfolio.Positions {
		symbols[i] = pos.Symbol
	}

	currentPrices, err := h.marketClient.GetCurrentPrices(symbols)
	if err != nil {
		h.logger.Error("Failed to get current prices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get market prices"})
		return
	}

	metrics, err := h.service.GetRiskMetrics(c.Request.Context(), portfolioID, currentPrices)
	if err != nil {
		h.logger.Error("Failed to get risk metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get risk metrics", Details: err.Error()})
		return
	}

	response := RiskMetricsResponse{
		TotalValue:           metrics["total_value"].(float64),
		PositionCount:        metrics["position_count"].(int),
		MaxPositionPercent:   metrics["max_position_percent"].(float64),
		CashPercent:          metrics["cash_percent"].(float64),
		DiversificationScore: metrics["diversification_score"].(float64),
	}

	c.JSON(http.StatusOK, response)
}

// GetRebalanceRecommendations godoc
// @Summary Get rebalancing recommendations
// @Description Get recommendations for rebalancing portfolio
// @Tags portfolios
// @Accept json
// @Produce json
// @Param id path int true "Portfolio ID"
// @Param request body RebalanceRequest true "Rebalance Request"
// @Success 200 {array} RebalanceRecommendation
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/portfolios/{id}/rebalance [post]
func (h *PortfolioHandler) GetRebalanceRecommendations(c *gin.Context) {
	portfolioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid portfolio ID"})
		return
	}

	var req RebalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Details: err.Error()})
		return
	}

	// Get portfolio
	portfolio, err := h.service.GetPortfolio(c.Request.Context(), portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Portfolio not found"})
		return
	}

	// Get current prices
	symbols := make([]string, len(portfolio.Positions))
	for i, pos := range portfolio.Positions {
		symbols[i] = pos.Symbol
	}

	// Add symbols from target allocations that might not be in portfolio
	for symbol := range req.TargetAllocations {
		found := false
		for _, s := range symbols {
			if s == symbol {
				found = true
				break
			}
		}
		if !found {
			symbols = append(symbols, symbol)
		}
	}

	currentPrices, err := h.marketClient.GetCurrentPrices(symbols)
	if err != nil {
		h.logger.Error("Failed to get current prices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get market prices"})
		return
	}

	recommendations, err := h.service.GetRebalanceRecommendations(c.Request.Context(), portfolioID, req.TargetAllocations, currentPrices)
	if err != nil {
		h.logger.Error("Failed to get rebalance recommendations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get recommendations", Details: err.Error()})
		return
	}

	// Convert to response
	response := make([]RebalanceRecommendation, len(recommendations))
	for i, rec := range recommendations {
		response[i] = RebalanceRecommendation{
			Symbol:          rec["symbol"].(string),
			CurrentPercent:  rec["current_percent"].(float64),
			TargetPercent:   rec["target_percent"].(float64),
			Difference:      rec["difference"].(float64),
			TargetValue:     rec["target_value"].(float64),
			CurrentValue:    rec["current_value"].(float64),
			Action:          rec["action"].(string),
			EstimatedShares: rec["estimated_shares"].(int64),
		}
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions to convert domain models to response DTOs

func (h *PortfolioHandler) toPortfolioResponse(portfolio *models.Portfolio) PortfolioResponse {
	positions := make([]PositionResponse, len(portfolio.Positions))
	for i, pos := range portfolio.Positions {
		positions[i] = h.toPositionResponse(&pos)
	}

	return PortfolioResponse{
		ID:              portfolio.ID,
		UserID:          portfolio.UserID,
		Name:            portfolio.Name,
		Cash:            portfolio.Cash,
		MarginUsed:      portfolio.MarginUsed,
		MarginAvailable: portfolio.MarginAvailable,
		TotalValue:      portfolio.TotalValue,
		UnrealizedPnL:   portfolio.UnrealizedPnL,
		RealizedPnL:     portfolio.RealizedPnL,
		DayPnL:          portfolio.DayPnL,
		Positions:       positions,
		CreatedAt:       portfolio.CreatedAt,
		UpdatedAt:       portfolio.UpdatedAt,
	}
}

func (h *PortfolioHandler) toPositionResponse(position *models.Position) PositionResponse {
	return PositionResponse{
		ID:            position.ID,
		PortfolioID:   position.PortfolioID,
		Symbol:        position.Symbol,
		Quantity:      position.Quantity,
		Side:          position.Side,
		EntryPrice:    position.EntryPrice,
		CurrentPrice:  position.CurrentPrice,
		UnrealizedPnL: position.UnrealizedPnL,
		RealizedPnL:   position.RealizedPnL,
		CreatedAt:     position.CreatedAt,
		UpdatedAt:     position.UpdatedAt,
	}
}

func (h *PortfolioHandler) toTradeResponse(trade *models.Trade, position *models.Position) TradeResponse {
	return TradeResponse{
		ID:          trade.ID,
		PortfolioID: trade.PortfolioID,
		PositionID:  trade.PositionID,
		Symbol:      trade.Symbol,
		Quantity:    trade.Quantity,
		Price:       trade.Price,
		Side:        trade.Side,
		Type:        trade.Type,
		Status:      trade.Status,
		Fees:        trade.Fees,
		ExecutedAt:  trade.ExecutedAt,
		CreatedAt:   trade.CreatedAt,
	}
}

func (h *PortfolioHandler) toSummaryResponse(summary *models.PortfolioSummary) SummaryResponse {
	return SummaryResponse{
		TotalValue:     summary.TotalValue,
		Cash:           summary.Cash,
		PositionsValue: summary.PositionsValue,
		UnrealizedPnL:  summary.UnrealizedPnL,
		RealizedPnL:    summary.RealizedPnL,
		DayPnL:         summary.DayPnL,
		DayReturn:      summary.DayReturn,
		TotalReturn:    summary.TotalReturn,
		PositionCount:  summary.PositionCount,
	}
}
