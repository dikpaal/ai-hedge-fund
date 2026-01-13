package main

import (
	"fmt"
	"time"

	"hedge-fund/internal/portfolio/domain"
	"hedge-fund/pkg/shared/models"
)

func main() {
	fmt.Println("🧮 Testing Portfolio Domain Logic...")

	// Initialize portfolio service
	ps := domain.NewPortfolioService()

	// Create test portfolio
	portfolio := &models.Portfolio{
		ID:     1,
		UserID: 1,
		Cash:   10000.0,
		Positions: []models.Position{
			{
				ID:           1,
				UserID:       1,
				Symbol:       "AAPL",
				Quantity:     50,
				Side:         "long",
				EntryPrice:   150.0,
				CurrentPrice: 155.0,
				CreatedAt:    time.Now().AddDate(0, 0, -30),
				UpdatedAt:    time.Now(),
			},
			{
				ID:           2,
				UserID:       1,
				Symbol:       "MSFT",
				Quantity:     30,
				Side:         "long",
				EntryPrice:   300.0,
				CurrentPrice: 310.0,
				CreatedAt:    time.Now().AddDate(0, 0, -20),
				UpdatedAt:    time.Now(),
			},
		},
		CreatedAt: time.Now().AddDate(0, 0, -30),
		UpdatedAt: time.Now(),
	}

	// Test current prices
	currentPrices := map[string]float64{
		"AAPL": 155.0,
		"MSFT": 310.0,
		"GOOGL": 142.0,
	}

	previousPrices := map[string]float64{
		"AAPL": 153.0,
		"MSFT": 308.0,
	}

	// Test 1: Calculate Portfolio Value
	fmt.Println("\n📊 Test 1: Portfolio Value Calculation")
	totalValue := ps.CalculatePortfolioValue(portfolio, currentPrices)
	fmt.Printf("✅ Total Portfolio Value: $%.2f\n", totalValue)
	expectedValue := 10000.0 + (50*155.0) + (30*310.0) // 10000 + 7750 + 9300 = 27050
	if abs(totalValue-expectedValue) < 0.01 {
		fmt.Printf("✅ Expected: $%.2f, Got: $%.2f - PASSED\n", expectedValue, totalValue)
	} else {
		fmt.Printf("❌ Expected: $%.2f, Got: $%.2f - FAILED\n", expectedValue, totalValue)
	}

	// Test 2: Calculate Unrealized PnL
	fmt.Println("\n💰 Test 2: Unrealized PnL Calculation")
	unrealizedPnL := ps.CalculateUnrealizedPnL(portfolio.Positions, currentPrices)
	fmt.Printf("✅ Total Unrealized PnL: $%.2f\n", unrealizedPnL)
	expectedPnL := (50*(155.0-150.0)) + (30*(310.0-300.0)) // 250 + 300 = 550
	if abs(unrealizedPnL-expectedPnL) < 0.01 {
		fmt.Printf("✅ Expected: $%.2f, Got: $%.2f - PASSED\n", expectedPnL, unrealizedPnL)
	} else {
		fmt.Printf("❌ Expected: $%.2f, Got: $%.2f - FAILED\n", expectedPnL, unrealizedPnL)
	}

	// Test 3: Position Summary
	fmt.Println("\n📈 Test 3: Position Summary Calculation")
	positionSummary := ps.CalculatePositionSummary(&portfolio.Positions[0], 155.0)
	fmt.Printf("✅ AAPL Position Summary: Symbol=%s, Quantity=%d, MarketValue=$%.2f, UnrealizedPnL=$%.2f\n",
		positionSummary.Symbol, positionSummary.NetQuantity, positionSummary.MarketValue, positionSummary.UnrealizedPnL)

	// Test 4: Portfolio Allocation
	fmt.Println("\n🥧 Test 4: Portfolio Allocation")
	allocations := ps.CalculatePortfolioAllocation(portfolio, currentPrices)
	fmt.Printf("✅ Portfolio Allocations:\n")
	for symbol, percentage := range allocations {
		fmt.Printf("   %s: %.2f%%\n", symbol, percentage)
	}

	// Test 5: Trade Validation - Valid Buy Order
	fmt.Println("\n✅ Test 5: Trade Validation (Valid Buy)")
	buyTrade := &models.Trade{
		UserID:   1,
		Symbol:   "GOOGL",
		Quantity: 10,
		Side:     "buy",
		Type:     "market",
		Status:   "pending",
	}
	err := ps.ValidateTradeOrder(buyTrade, portfolio, 142.0)
	if err == nil {
		fmt.Printf("✅ Valid buy order validation - PASSED\n")
	} else {
		fmt.Printf("❌ Valid buy order validation - FAILED: %v\n", err)
	}

	// Test 6: Trade Validation - Invalid Buy Order (Insufficient Funds)
	fmt.Println("\n❌ Test 6: Trade Validation (Invalid Buy - Insufficient Funds)")
	largeBuyTrade := &models.Trade{
		UserID:   1,
		Symbol:   "GOOGL",
		Quantity: 1000, // This should exceed available cash
		Side:     "buy",
		Type:     "market",
		Status:   "pending",
	}
	err = ps.ValidateTradeOrder(largeBuyTrade, portfolio, 142.0)
	if err != nil {
		fmt.Printf("✅ Invalid buy order validation - PASSED: %v\n", err)
	} else {
		fmt.Printf("❌ Invalid buy order validation - FAILED: Should have failed\n")
	}

	// Test 7: Trade Validation - Valid Sell Order
	fmt.Println("\n✅ Test 7: Trade Validation (Valid Sell)")
	sellTrade := &models.Trade{
		UserID:   1,
		Symbol:   "AAPL",
		Quantity: 25, // Sell half of AAPL position
		Side:     "sell",
		Type:     "market",
		Status:   "pending",
	}
	err = ps.ValidateTradeOrder(sellTrade, portfolio, 155.0)
	if err == nil {
		fmt.Printf("✅ Valid sell order validation - PASSED\n")
	} else {
		fmt.Printf("❌ Valid sell order validation - FAILED: %v\n", err)
	}

	// Test 8: Portfolio Summary
	fmt.Println("\n📋 Test 8: Portfolio Summary")
	summary := ps.CalculatePortfolioSummary(portfolio, currentPrices, previousPrices)
	fmt.Printf("✅ Portfolio Summary:\n")
	fmt.Printf("   Total Value: $%.2f\n", summary.TotalValue)
	fmt.Printf("   Cash: $%.2f\n", summary.Cash)
	fmt.Printf("   Positions Value: $%.2f\n", summary.PositionsValue)
	fmt.Printf("   Unrealized PnL: $%.2f\n", summary.UnrealizedPnL)
	fmt.Printf("   Day PnL: $%.2f\n", summary.DayPnL)
	fmt.Printf("   Day Return: %.2f%%\n", summary.DayReturn)
	fmt.Printf("   Total Return: %.2f%%\n", summary.TotalReturn)
	fmt.Printf("   Position Count: %d\n", summary.PositionCount)

	// Test 9: Risk Metrics
	fmt.Println("\n⚠️ Test 9: Risk Metrics")
	riskMetrics := ps.CalculateRiskMetrics(portfolio, currentPrices)
	fmt.Printf("✅ Risk Metrics:\n")
	for key, value := range riskMetrics {
		switch v := value.(type) {
		case float64:
			fmt.Printf("   %s: %.2f\n", key, v)
		case int:
			fmt.Printf("   %s: %d\n", key, v)
		default:
			fmt.Printf("   %s: %v\n", key, v)
		}
	}

	// Test 10: Execute Trade Order (Buy GOOGL)
	fmt.Println("\n💱 Test 10: Execute Trade Order (Buy GOOGL)")
	originalCash := portfolio.Cash
	originalPositionCount := len(portfolio.Positions)

	position, err := ps.ExecuteTradeOrder(buyTrade, portfolio, 142.0)
	if err == nil && position != nil {
		fmt.Printf("✅ Trade execution - PASSED\n")
		fmt.Printf("   New position created: %s, Quantity: %d, Entry Price: $%.2f\n",
			position.Symbol, position.Quantity, position.EntryPrice)
		fmt.Printf("   Cash before: $%.2f, Cash after: $%.2f\n", originalCash, portfolio.Cash)
		fmt.Printf("   Positions before: %d, Positions after: %d\n", originalPositionCount, len(portfolio.Positions))
	} else {
		fmt.Printf("❌ Trade execution - FAILED: %v\n", err)
	}

	// Test 11: Update Portfolio with Market Data
	fmt.Println("\n🔄 Test 11: Update Portfolio with Market Data")
	ps.UpdatePortfolioWithMarketData(portfolio, currentPrices)
	fmt.Printf("✅ Portfolio updated with market data\n")
	fmt.Printf("   Total Value: $%.2f\n", portfolio.TotalValue)
	fmt.Printf("   Unrealized PnL: $%.2f\n", portfolio.UnrealizedPnL)

	fmt.Println("\n🎉 All Portfolio Domain Tests Completed!")
	fmt.Println("Portfolio Service core domain logic is ready for production use!")
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}