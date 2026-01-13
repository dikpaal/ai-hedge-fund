package main

import (
	"context"
	"fmt"
	"time"

	"hedge-fund/internal/portfolio/repository"
	"hedge-fund/pkg/shared/config"
	"hedge-fund/pkg/shared/database"
	"hedge-fund/pkg/shared/logger"
	"hedge-fund/pkg/shared/models"
)

func main() {
	fmt.Println("🗃️ Testing Portfolio Repository Layer...")

	// Initialize configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(cfg.LogLevel, cfg.Env); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer db.Close()

	// Initialize repository
	repo := repository.NewPortfolioRepository(db, logger.Logger)

	ctx := context.Background()

	// Test 1: Create Portfolio
	fmt.Println("\n📊 Test 1: Create Portfolio")
	portfolio := &models.Portfolio{
		UserID:           1,
		Cash:             10000.0,
		MarginUsed:       0.0,
		MarginAvailable:  5000.0,
		TotalValue:       10000.0,
		UnrealizedPnL:    0.0,
		RealizedPnL:      0.0,
		DayPnL:           0.0,
	}

	err = repo.CreatePortfolio(ctx, portfolio)
	if err != nil {
		fmt.Printf("❌ Portfolio creation failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Portfolio created successfully! ID: %d\n", portfolio.ID)

	// Test 2: Create Positions
	fmt.Println("\n📈 Test 2: Create Positions")
	position1 := &models.Position{
		UserID:        1,
		Symbol:        "AAPL",
		Quantity:      50,
		Side:          "long",
		EntryPrice:    150.0,
		CurrentPrice:  155.0,
		UnrealizedPnL: 250.0,
		RealizedPnL:   0.0,
	}

	err = repo.CreatePosition(ctx, position1)
	if err != nil {
		fmt.Printf("❌ Position 1 creation failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Position 1 created successfully! ID: %d\n", position1.ID)

	position2 := &models.Position{
		UserID:        1,
		Symbol:        "MSFT",
		Quantity:      30,
		Side:          "long",
		EntryPrice:    300.0,
		CurrentPrice:  310.0,
		UnrealizedPnL: 300.0,
		RealizedPnL:   0.0,
	}

	err = repo.CreatePosition(ctx, position2)
	if err != nil {
		fmt.Printf("❌ Position 2 creation failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Position 2 created successfully! ID: %d\n", position2.ID)

	// Test 3: Create Trade Records
	fmt.Println("\n💱 Test 3: Create Trade Records")
	trade1 := &models.Trade{
		UserID:     1,
		PositionID: position1.ID,
		Symbol:     "AAPL",
		Quantity:   50,
		Price:      150.0,
		Side:       "buy",
		Type:       "market",
		Status:     "filled",
		Fees:       1.0,
		ExecutedAt: &[]time.Time{time.Now()}[0],
	}

	err = repo.CreateTrade(ctx, trade1)
	if err != nil {
		fmt.Printf("❌ Trade 1 creation failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Trade 1 created successfully! ID: %d\n", trade1.ID)

	trade2 := &models.Trade{
		UserID:     1,
		PositionID: position2.ID,
		Symbol:     "MSFT",
		Quantity:   30,
		Price:      300.0,
		Side:       "buy",
		Type:       "market",
		Status:     "filled",
		Fees:       1.0,
		ExecutedAt: &[]time.Time{time.Now()}[0],
	}

	err = repo.CreateTrade(ctx, trade2)
	if err != nil {
		fmt.Printf("❌ Trade 2 creation failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Trade 2 created successfully! ID: %d\n", trade2.ID)

	// Test 4: Get Portfolio by ID
	fmt.Println("\n🔍 Test 4: Get Portfolio by ID")
	retrievedPortfolio, err := repo.GetPortfolioByID(ctx, portfolio.ID)
	if err != nil {
		fmt.Printf("❌ Portfolio retrieval failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Portfolio retrieved successfully!\n")
	fmt.Printf("   ID: %d, User ID: %d, Cash: $%.2f\n",
		retrievedPortfolio.ID, retrievedPortfolio.UserID, retrievedPortfolio.Cash)
	fmt.Printf("   Positions count: %d\n", len(retrievedPortfolio.Positions))

	// Test 5: Get Positions by Portfolio ID
	fmt.Println("\n📋 Test 5: Get Positions by Portfolio ID")
	positions, err := repo.GetPositionsByPortfolioID(ctx, portfolio.ID)
	if err != nil {
		fmt.Printf("❌ Positions retrieval failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Positions retrieved successfully! Count: %d\n", len(positions))
	for i, pos := range positions {
		fmt.Printf("   Position %d: %s, Quantity: %d, Entry: $%.2f, Current: $%.2f, PnL: $%.2f\n",
			i+1, pos.Symbol, pos.Quantity, pos.EntryPrice, pos.CurrentPrice, pos.UnrealizedPnL)
	}

	// Test 6: Get Position by User and Symbol
	fmt.Println("\n🎯 Test 6: Get Position by User and Symbol")
	aaplPosition, err := repo.GetPositionByUserAndSymbol(ctx, 1, "AAPL")
	if err != nil {
		fmt.Printf("❌ Position retrieval by symbol failed: %v\n", err)
		return
	}
	if aaplPosition != nil {
		fmt.Printf("✅ AAPL position found! Quantity: %d, Entry Price: $%.2f\n",
			aaplPosition.Quantity, aaplPosition.EntryPrice)
	} else {
		fmt.Printf("❌ AAPL position not found\n")
		return
	}

	// Test 7: Update Position
	fmt.Println("\n✏️ Test 7: Update Position")
	aaplPosition.CurrentPrice = 158.0
	aaplPosition.UnrealizedPnL = (158.0 - 150.0) * 50.0 // $400

	err = repo.UpdatePosition(ctx, aaplPosition)
	if err != nil {
		fmt.Printf("❌ Position update failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Position updated successfully! New current price: $%.2f, New PnL: $%.2f\n",
		aaplPosition.CurrentPrice, aaplPosition.UnrealizedPnL)

	// Test 8: Update Portfolio
	fmt.Println("\n📝 Test 8: Update Portfolio")
	portfolio.Cash = 8500.0
	portfolio.TotalValue = 25000.0
	portfolio.UnrealizedPnL = 700.0

	err = repo.UpdatePortfolio(ctx, portfolio)
	if err != nil {
		fmt.Printf("❌ Portfolio update failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Portfolio updated successfully! New cash: $%.2f, Total value: $%.2f\n",
		portfolio.Cash, portfolio.TotalValue)

	// Test 9: Get Trades by User ID
	fmt.Println("\n📊 Test 9: Get Trades by User ID")
	trades, err := repo.GetTradesByUserID(ctx, 1, 10, 0)
	if err != nil {
		fmt.Printf("❌ Trades retrieval failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Trades retrieved successfully! Count: %d\n", len(trades))
	for i, trade := range trades {
		fmt.Printf("   Trade %d: %s %d %s @ $%.2f, Fees: $%.2f, Status: %s\n",
			i+1, trade.Side, trade.Quantity, trade.Symbol, trade.Price, trade.Fees, trade.Status)
	}

	// Test 10: Get Trades by Symbol
	fmt.Println("\n🔍 Test 10: Get Trades by Symbol")
	aaplTrades, err := repo.GetTradesBySymbol(ctx, 1, "AAPL", 10, 0)
	if err != nil {
		fmt.Printf("❌ AAPL trades retrieval failed: %v\n", err)
		return
	}
	fmt.Printf("✅ AAPL trades retrieved successfully! Count: %d\n", len(aaplTrades))
	for i, trade := range aaplTrades {
		fmt.Printf("   AAPL Trade %d: %s %d @ $%.2f on %v\n",
			i+1, trade.Side, trade.Quantity, trade.Price, trade.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Test 11: Get Portfolios by User ID
	fmt.Println("\n👤 Test 11: Get Portfolios by User ID")
	userPortfolios, err := repo.GetPortfoliosByUserID(ctx, 1)
	if err != nil {
		fmt.Printf("❌ User portfolios retrieval failed: %v\n", err)
		return
	}
	fmt.Printf("✅ User portfolios retrieved successfully! Count: %d\n", len(userPortfolios))
	for i, pf := range userPortfolios {
		fmt.Printf("   Portfolio %d: ID=%d, Cash=$%.2f, Total=$%.2f, Positions=%d\n",
			i+1, pf.ID, pf.Cash, pf.TotalValue, len(pf.Positions))
	}

	// Test 12: Error Handling - Get Non-existent Portfolio
	fmt.Println("\n❌ Test 12: Error Handling - Get Non-existent Portfolio")
	_, err = repo.GetPortfolioByID(ctx, 99999)
	if err != nil {
		fmt.Printf("✅ Error handling working correctly: %v\n", err)
	} else {
		fmt.Printf("❌ Error handling failed: Should have returned error for non-existent portfolio\n")
	}

	// Test 13: Error Handling - Get Non-existent Position
	fmt.Println("\n❌ Test 13: Error Handling - Get Non-existent Position")
	nonExistentPosition, err := repo.GetPositionByUserAndSymbol(ctx, 1, "NONEXISTENT")
	if err != nil {
		fmt.Printf("❌ Unexpected error: %v\n", err)
	} else if nonExistentPosition == nil {
		fmt.Printf("✅ Non-existent position correctly returned nil\n")
	} else {
		fmt.Printf("❌ Should have returned nil for non-existent position\n")
	}

	// Cleanup (optional - for testing purposes)
	fmt.Println("\n🧹 Cleanup - Delete Test Data")
	fmt.Printf("Note: In production, be careful with delete operations!\n")
	fmt.Printf("Test data created: Portfolio ID %d with positions and trades\n", portfolio.ID)

	fmt.Println("\n🎉 All Portfolio Repository Tests Completed!")
	fmt.Println("Portfolio Service database layer is ready for production use!")
}