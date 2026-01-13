package domain

import (
	"fmt"
	"time"

	"hedge-fund/pkg/shared/models"
)

type PortfolioService struct{}

func NewPortfolioService() *PortfolioService {
	return &PortfolioService{}
}

// CalculatePortfolioValue calculates the total value of a portfolio
func (ps *PortfolioService) CalculatePortfolioValue(portfolio *models.Portfolio, currentPrices map[string]float64) float64 {
	totalValue := portfolio.Cash

	for _, position := range portfolio.Positions {
		if currentPrice, exists := currentPrices[position.Symbol]; exists {
			totalValue += float64(position.Quantity) * currentPrice
		}
	}

	return totalValue
}

// CalculateUnrealizedPnL calculates unrealized profit and loss for all positions
func (ps *PortfolioService) CalculateUnrealizedPnL(positions []models.Position, currentPrices map[string]float64) float64 {
	totalPnL := 0.0

	for _, position := range positions {
		if currentPrice, exists := currentPrices[position.Symbol]; exists {
			unrealizedPnL := (currentPrice - position.EntryPrice) * float64(position.Quantity)
			totalPnL += unrealizedPnL
		}
	}

	return totalPnL
}

// CalculatePositionSummary calculates detailed metrics for a specific position
func (ps *PortfolioService) CalculatePositionSummary(position *models.Position, currentPrice float64) models.PositionSummary {
	marketValue := float64(position.Quantity) * currentPrice
	unrealizedPnL := (currentPrice - position.EntryPrice) * float64(position.Quantity)
	unrealizedReturn := 0.0
	if position.EntryPrice > 0 {
		unrealizedReturn = (unrealizedPnL / (position.EntryPrice * float64(position.Quantity))) * 100
	}

	return models.PositionSummary{
		Symbol:           position.Symbol,
		NetQuantity:      position.Quantity,
		LongQuantity:     position.Quantity, // Assuming long positions for now
		ShortQuantity:    0,
		AveragePrice:     position.EntryPrice,
		CurrentPrice:     currentPrice,
		MarketValue:      marketValue,
		UnrealizedPnL:    unrealizedPnL,
		UnrealizedReturn: unrealizedReturn,
	}
}

// ValidateTradeOrder validates a trade order before execution
func (ps *PortfolioService) ValidateTradeOrder(trade *models.Trade, portfolio *models.Portfolio, currentPrice float64) error {
	if trade.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	if currentPrice <= 0 {
		return fmt.Errorf("invalid current price: %.4f", currentPrice)
	}

	if trade.Side == "buy" {
		// Check if sufficient cash for buy order
		orderValue := float64(trade.Quantity) * currentPrice
		fees := ps.calculateCommission(orderValue)
		totalCost := orderValue + fees

		if portfolio.Cash < totalCost {
			return fmt.Errorf("insufficient cash balance: need %.2f, have %.2f", totalCost, portfolio.Cash)
		}
	} else if trade.Side == "sell" {
		// Check if sufficient shares for sell order
		position := ps.findPosition(portfolio.Positions, trade.Symbol)
		if position == nil || position.Quantity < trade.Quantity {
			availableQuantity := int64(0)
			if position != nil {
				availableQuantity = position.Quantity
			}
			return fmt.Errorf("insufficient shares: need %d, have %d", trade.Quantity, availableQuantity)
		}
	} else {
		return fmt.Errorf("invalid order side: %s", trade.Side)
	}

	return nil
}

// ExecuteTradeOrder executes a validated trade order and updates portfolio state
func (ps *PortfolioService) ExecuteTradeOrder(trade *models.Trade, portfolio *models.Portfolio, currentPrice float64) (*models.Position, error) {
	trade.Price = currentPrice
	trade.Fees = ps.calculateCommission(float64(trade.Quantity) * currentPrice)
	trade.Status = "filled"
	executedAt := time.Now()
	trade.ExecutedAt = &executedAt

	tradeValue := float64(trade.Quantity) * currentPrice
	position := ps.findPositionByIndex(portfolio.Positions, trade.Symbol)

	if trade.Side == "buy" {
		// Update cash balance
		portfolio.Cash -= tradeValue + trade.Fees

		// Update or create position
		if position == -1 {
			// Create new position
			newPosition := models.Position{
				UserID:        trade.UserID,
				Symbol:        trade.Symbol,
				Quantity:      trade.Quantity,
				Side:          "long",
				EntryPrice:    currentPrice,
				CurrentPrice:  currentPrice,
				UnrealizedPnL: 0.0,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			portfolio.Positions = append(portfolio.Positions, newPosition)
			return &newPosition, nil
		} else {
			// Update existing position with weighted average cost
			pos := &portfolio.Positions[position]
			totalCost := (pos.EntryPrice * float64(pos.Quantity)) + tradeValue
			totalQuantity := pos.Quantity + trade.Quantity
			pos.EntryPrice = totalCost / float64(totalQuantity)
			pos.Quantity = totalQuantity
			pos.CurrentPrice = currentPrice
			pos.UnrealizedPnL = (currentPrice - pos.EntryPrice) * float64(totalQuantity)
			pos.UpdatedAt = time.Now()
			return pos, nil
		}
	} else { // sell
		if position == -1 {
			return nil, fmt.Errorf("position not found for symbol %s", trade.Symbol)
		}

		// Update cash balance
		portfolio.Cash += tradeValue - trade.Fees

		// Update position
		pos := &portfolio.Positions[position]
		pos.Quantity -= trade.Quantity
		pos.CurrentPrice = currentPrice

		if pos.Quantity == 0 {
			// Position fully closed - remove from portfolio
			portfolio.Positions = append(portfolio.Positions[:position], portfolio.Positions[position+1:]...)
			return nil, nil
		} else {
			// Partial sale - entry price remains the same
			pos.UnrealizedPnL = (currentPrice - pos.EntryPrice) * float64(pos.Quantity)
			pos.UpdatedAt = time.Now()
			return pos, nil
		}
	}
}

// CalculatePortfolioAllocation calculates allocation percentages for each position
func (ps *PortfolioService) CalculatePortfolioAllocation(portfolio *models.Portfolio, currentPrices map[string]float64) map[string]float64 {
	totalValue := ps.CalculatePortfolioValue(portfolio, currentPrices)
	allocations := make(map[string]float64)

	// Cash allocation
	if totalValue > 0 {
		allocations["CASH"] = (portfolio.Cash / totalValue) * 100
	}

	// Position allocations
	for _, position := range portfolio.Positions {
		if currentPrice, exists := currentPrices[position.Symbol]; exists {
			positionValue := float64(position.Quantity) * currentPrice
			if totalValue > 0 {
				allocations[position.Symbol] = (positionValue / totalValue) * 100
			}
		}
	}

	return allocations
}

// CalculatePortfolioSummary generates a comprehensive portfolio summary
func (ps *PortfolioService) CalculatePortfolioSummary(portfolio *models.Portfolio, currentPrices map[string]float64, previousDayPrices map[string]float64) models.PortfolioSummary {
	totalValue := ps.CalculatePortfolioValue(portfolio, currentPrices)
	positionsValue := totalValue - portfolio.Cash
	unrealizedPnL := ps.CalculateUnrealizedPnL(portfolio.Positions, currentPrices)

	// Calculate day PnL based on price changes
	dayPnL := 0.0
	for _, position := range portfolio.Positions {
		if currentPrice, exists := currentPrices[position.Symbol]; exists {
			if previousPrice, prevExists := previousDayPrices[position.Symbol]; prevExists {
				dayChange := (currentPrice - previousPrice) * float64(position.Quantity)
				dayPnL += dayChange
			}
		}
	}

	dayReturn := 0.0
	if totalValue > 0 {
		dayReturn = (dayPnL / totalValue) * 100
	}

	// Calculate total return based on initial investment
	totalReturn := 0.0
	if positionsValue > 0 {
		totalReturn = (unrealizedPnL / positionsValue) * 100
	}

	return models.PortfolioSummary{
		TotalValue:     totalValue,
		Cash:           portfolio.Cash,
		PositionsValue: positionsValue,
		UnrealizedPnL:  unrealizedPnL,
		RealizedPnL:    portfolio.RealizedPnL,
		DayPnL:         dayPnL,
		DayReturn:      dayReturn,
		TotalReturn:    totalReturn,
		PositionCount:  len(portfolio.Positions),
	}
}

// UpdatePortfolioWithMarketData updates portfolio positions with current market prices
func (ps *PortfolioService) UpdatePortfolioWithMarketData(portfolio *models.Portfolio, currentPrices map[string]float64) {
	totalUnrealizedPnL := 0.0
	totalValue := portfolio.Cash

	for i := range portfolio.Positions {
		position := &portfolio.Positions[i]
		if currentPrice, exists := currentPrices[position.Symbol]; exists {
			position.CurrentPrice = currentPrice
			position.UnrealizedPnL = (currentPrice - position.EntryPrice) * float64(position.Quantity)
			position.UpdatedAt = time.Now()

			totalUnrealizedPnL += position.UnrealizedPnL
			totalValue += float64(position.Quantity) * currentPrice
		}
	}

	portfolio.UnrealizedPnL = totalUnrealizedPnL
	portfolio.TotalValue = totalValue
	portfolio.UpdatedAt = time.Now()
}

// CalculateRiskMetrics calculates basic risk metrics for positions
func (ps *PortfolioService) CalculateRiskMetrics(portfolio *models.Portfolio, currentPrices map[string]float64) map[string]interface{} {
	totalValue := ps.CalculatePortfolioValue(portfolio, currentPrices)
	metrics := make(map[string]interface{})

	// Concentration risk - largest position percentage
	maxPositionPercent := 0.0
	positionCount := len(portfolio.Positions)

	for _, position := range portfolio.Positions {
		if currentPrice, exists := currentPrices[position.Symbol]; exists {
			positionValue := float64(position.Quantity) * currentPrice
			positionPercent := (positionValue / totalValue) * 100
			if positionPercent > maxPositionPercent {
				maxPositionPercent = positionPercent
			}
		}
	}

	// Cash percentage
	cashPercent := (portfolio.Cash / totalValue) * 100

	metrics["total_value"] = totalValue
	metrics["position_count"] = positionCount
	metrics["max_position_percent"] = maxPositionPercent
	metrics["cash_percent"] = cashPercent
	metrics["diversification_score"] = ps.calculateDiversificationScore(portfolio.Positions, totalValue, currentPrices)

	return metrics
}

// RebalanceRecommendations suggests portfolio rebalancing based on target allocations
func (ps *PortfolioService) RebalanceRecommendations(portfolio *models.Portfolio, targetAllocations map[string]float64, currentPrices map[string]float64) []map[string]interface{} {
	totalValue := ps.CalculatePortfolioValue(portfolio, currentPrices)
	currentAllocations := ps.CalculatePortfolioAllocation(portfolio, currentPrices)

	var recommendations []map[string]interface{}

	for symbol, targetPercent := range targetAllocations {
		currentPercent := currentAllocations[symbol]
		if currentPercent == 0 {
			currentPercent = 0
		}

		diff := targetPercent - currentPercent
		if abs(diff) > 1.0 { // Only recommend if difference > 1%
			targetValue := (targetPercent / 100) * totalValue
			currentValue := (currentPercent / 100) * totalValue

			if currentPrice, exists := currentPrices[symbol]; exists {
				recommendation := map[string]interface{}{
					"symbol":         symbol,
					"current_percent": currentPercent,
					"target_percent":  targetPercent,
					"difference":      diff,
					"target_value":    targetValue,
					"current_value":   currentValue,
					"action":          ps.getRebalanceAction(diff),
					"estimated_shares": int64((targetValue - currentValue) / currentPrice),
				}
				recommendations = append(recommendations, recommendation)
			}
		}
	}

	return recommendations
}

// Helper functions

func (ps *PortfolioService) calculateCommission(tradeValue float64) float64 {
	// Simple commission structure: $1 minimum, 0.1% of trade value
	commission := tradeValue * 0.001
	if commission < 1.0 {
		commission = 1.0
	}
	return commission
}

func (ps *PortfolioService) findPosition(positions []models.Position, symbol string) *models.Position {
	for i := range positions {
		if positions[i].Symbol == symbol {
			return &positions[i]
		}
	}
	return nil
}

func (ps *PortfolioService) findPositionByIndex(positions []models.Position, symbol string) int {
	for i, position := range positions {
		if position.Symbol == symbol {
			return i
		}
	}
	return -1
}

func (ps *PortfolioService) calculateDiversificationScore(positions []models.Position, totalValue float64, currentPrices map[string]float64) float64 {
	if len(positions) <= 1 {
		return 0.0
	}

	// Simple diversification score based on Herfindahl index
	sum := 0.0
	for _, position := range positions {
		if currentPrice, exists := currentPrices[position.Symbol]; exists {
			positionValue := float64(position.Quantity) * currentPrice
			weight := positionValue / totalValue
			sum += weight * weight
		}
	}

	// Convert to 0-100 scale (100 = perfectly diversified)
	return (1.0 - sum) * 100
}

func (ps *PortfolioService) getRebalanceAction(diff float64) string {
	if diff > 1.0 {
		return "buy"
	} else if diff < -1.0 {
		return "sell"
	}
	return "hold"
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}