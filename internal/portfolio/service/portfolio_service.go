package service

import (
	"context"
	"fmt"

	"hedge-fund/internal/portfolio/domain"
	"hedge-fund/internal/portfolio/repository"
	"hedge-fund/pkg/shared/models"
	"go.uber.org/zap"
)

type PortfolioService struct {
	repo   *repository.PortfolioRepository
	domain *domain.PortfolioService
	logger *zap.Logger
}

func NewPortfolioService(repo *repository.PortfolioRepository, domain *domain.PortfolioService, logger *zap.Logger) *PortfolioService {
	return &PortfolioService{
		repo:   repo,
		domain: domain,
		logger: logger,
	}
}

// Portfolio Operations

// CreatePortfolio creates a new portfolio with initial cash
func (s *PortfolioService) CreatePortfolio(ctx context.Context, userID int, initialCash float64) (*models.Portfolio, error) {
	portfolio := &models.Portfolio{
		UserID:           userID,
		Cash:             initialCash,
		MarginUsed:       0.0,
		MarginAvailable:  initialCash * 0.5, // 50% margin
		TotalValue:       initialCash,
		UnrealizedPnL:    0.0,
		RealizedPnL:      0.0,
		DayPnL:           0.0,
		Positions:        []models.Position{},
	}

	err := s.repo.CreatePortfolio(ctx, portfolio)
	if err != nil {
		s.logger.Error("Failed to create portfolio", zap.Error(err), zap.Int("user_id", userID))
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}

	s.logger.Info("Portfolio created successfully",
		zap.Int("portfolio_id", portfolio.ID),
		zap.Int("user_id", userID),
		zap.Float64("initial_cash", initialCash))

	return portfolio, nil
}

// GetPortfolio retrieves a portfolio by ID with all positions
func (s *PortfolioService) GetPortfolio(ctx context.Context, portfolioID int) (*models.Portfolio, error) {
	return s.repo.GetPortfolioByID(ctx, portfolioID)
}

// GetUserPortfolios retrieves all portfolios for a user
func (s *PortfolioService) GetUserPortfolios(ctx context.Context, userID int) ([]models.Portfolio, error) {
	return s.repo.GetPortfoliosByUserID(ctx, userID)
}

// CalculatePortfolioSummary generates a comprehensive portfolio summary with current market data
func (s *PortfolioService) CalculatePortfolioSummary(ctx context.Context, portfolioID int, currentPrices map[string]float64, previousDayPrices map[string]float64) (*models.PortfolioSummary, error) {
	portfolio, err := s.repo.GetPortfolioByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	summary := s.domain.CalculatePortfolioSummary(portfolio, currentPrices, previousDayPrices)
	return &summary, nil
}

// UpdatePortfolioWithMarketData updates portfolio positions with current market prices
func (s *PortfolioService) UpdatePortfolioWithMarketData(ctx context.Context, portfolioID int, currentPrices map[string]float64) error {
	portfolio, err := s.repo.GetPortfolioByID(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("failed to get portfolio: %w", err)
	}

	// Update portfolio with market data using domain logic
	s.domain.UpdatePortfolioWithMarketData(portfolio, currentPrices)

	// Save updated portfolio to database
	err = s.repo.UpdatePortfolio(ctx, portfolio)
	if err != nil {
		return fmt.Errorf("failed to update portfolio: %w", err)
	}

	s.logger.Info("Portfolio updated with market data",
		zap.Int("portfolio_id", portfolioID),
		zap.Float64("total_value", portfolio.TotalValue),
		zap.Float64("unrealized_pnl", portfolio.UnrealizedPnL))

	return nil
}

// Trading Operations

// ExecuteTrade executes a trade order and updates portfolio state
func (s *PortfolioService) ExecuteTrade(ctx context.Context, portfolioID int, trade *models.Trade, currentPrice float64) (*models.Position, error) {
	// Get portfolio
	portfolio, err := s.repo.GetPortfolioByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	// Validate trade using domain logic
	err = s.domain.ValidateTradeOrder(trade, portfolio, currentPrice)
	if err != nil {
		s.logger.Warn("Trade validation failed",
			zap.Error(err),
			zap.Int("portfolio_id", portfolioID),
			zap.String("symbol", trade.Symbol),
			zap.String("side", trade.Side),
			zap.Int64("quantity", trade.Quantity))
		return nil, fmt.Errorf("trade validation failed: %w", err)
	}

	// Execute trade using domain logic (updates portfolio state in-memory)
	position, err := s.domain.ExecuteTradeOrder(trade, portfolio, currentPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to execute trade: %w", err)
	}

	// Set portfolio_id on trade
	trade.PortfolioID = portfolioID

	// Begin database transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Handle position operations FIRST (so we get the position ID)
	var finalPosition *models.Position
	if position != nil {
		// Set portfolio_id on position
		position.PortfolioID = portfolioID

		// Check if position already exists
		existingPosition, err := s.repo.GetPositionByUserAndSymbol(ctx, trade.UserID, trade.Symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing position: %w", err)
		}

		if existingPosition == nil {
			// Create new position in transaction
			err = s.repo.CreatePositionTx(ctx, tx, position)
			if err != nil {
				return nil, fmt.Errorf("failed to create position: %w", err)
			}
			finalPosition = position
		} else {
			// Update existing position in transaction
			position.ID = existingPosition.ID
			err = s.repo.UpdatePositionTx(ctx, tx, position)
			if err != nil {
				return nil, fmt.Errorf("failed to update position: %w", err)
			}
			finalPosition = position
		}

		// Set position_id on trade (now we have the position ID)
		trade.PositionID = finalPosition.ID
	} else {
		// Position was closed, need to get existing position for trade record
		existingPosition, err := s.repo.GetPositionByUserAndSymbol(ctx, trade.UserID, trade.Symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing position: %w", err)
		}

		if existingPosition != nil {
			// Set position_id before deletion
			trade.PositionID = existingPosition.ID

			// Delete the position in transaction
			err = s.repo.DeletePositionTx(ctx, tx, existingPosition.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to delete position: %w", err)
			}
		}
	}

	// Create trade record (position_id is now set)
	err = s.repo.CreateTradeTx(ctx, tx, trade)
	if err != nil {
		return nil, fmt.Errorf("failed to create trade record: %w", err)
	}

	// Update portfolio
	err = s.repo.UpdatePortfolioTx(ctx, tx, portfolio)
	if err != nil {
		return nil, fmt.Errorf("failed to update portfolio: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Trade executed successfully",
		zap.Int("trade_id", trade.ID),
		zap.Int("portfolio_id", portfolioID),
		zap.String("symbol", trade.Symbol),
		zap.String("side", trade.Side),
		zap.Int64("quantity", trade.Quantity),
		zap.Float64("price", trade.Price),
		zap.Float64("fees", trade.Fees))

	return finalPosition, nil
}

// GetTradeHistory retrieves trade history for a portfolio
func (s *PortfolioService) GetTradeHistory(ctx context.Context, userID int, limit, offset int) ([]models.Trade, error) {
	return s.repo.GetTradesByUserID(ctx, userID, limit, offset)
}

// GetSymbolTrades retrieves trades for a specific symbol
func (s *PortfolioService) GetSymbolTrades(ctx context.Context, userID int, symbol string, limit, offset int) ([]models.Trade, error) {
	return s.repo.GetTradesBySymbol(ctx, userID, symbol, limit, offset)
}

// Position Operations

// GetPositions retrieves all positions for a portfolio
func (s *PortfolioService) GetPositions(ctx context.Context, portfolioID int) ([]models.Position, error) {
	return s.repo.GetPositionsByPortfolioID(ctx, portfolioID)
}

// GetPosition retrieves a specific position
func (s *PortfolioService) GetPosition(ctx context.Context, userID int, symbol string) (*models.Position, error) {
	return s.repo.GetPositionByUserAndSymbol(ctx, userID, symbol)
}

// GetPositionSummary calculates detailed metrics for a specific position
func (s *PortfolioService) GetPositionSummary(ctx context.Context, positionID int, currentPrice float64) (*models.PositionSummary, error) {
	position, err := s.repo.GetPositionByID(ctx, positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	summary := s.domain.CalculatePositionSummary(position, currentPrice)
	return &summary, nil
}

// Analysis Operations

// GetPortfolioAllocation calculates allocation percentages for each position
func (s *PortfolioService) GetPortfolioAllocation(ctx context.Context, portfolioID int, currentPrices map[string]float64) (map[string]float64, error) {
	portfolio, err := s.repo.GetPortfolioByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	return s.domain.CalculatePortfolioAllocation(portfolio, currentPrices), nil
}

// GetRiskMetrics calculates basic risk metrics for the portfolio
func (s *PortfolioService) GetRiskMetrics(ctx context.Context, portfolioID int, currentPrices map[string]float64) (map[string]interface{}, error) {
	portfolio, err := s.repo.GetPortfolioByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	return s.domain.CalculateRiskMetrics(portfolio, currentPrices), nil
}

// GetRebalanceRecommendations suggests portfolio rebalancing based on target allocations
func (s *PortfolioService) GetRebalanceRecommendations(ctx context.Context, portfolioID int, targetAllocations map[string]float64, currentPrices map[string]float64) ([]map[string]interface{}, error) {
	portfolio, err := s.repo.GetPortfolioByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	return s.domain.RebalanceRecommendations(portfolio, targetAllocations, currentPrices), nil
}

// Portfolio Management

// UpdatePortfolio updates portfolio information
func (s *PortfolioService) UpdatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	err := s.repo.UpdatePortfolio(ctx, portfolio)
	if err != nil {
		return fmt.Errorf("failed to update portfolio: %w", err)
	}

	s.logger.Info("Portfolio updated",
		zap.Int("portfolio_id", portfolio.ID),
		zap.Float64("cash", portfolio.Cash),
		zap.Float64("total_value", portfolio.TotalValue))

	return nil
}

// DeletePortfolio deletes a portfolio and all its positions
func (s *PortfolioService) DeletePortfolio(ctx context.Context, portfolioID int) error {
	err := s.repo.DeletePortfolio(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("failed to delete portfolio: %w", err)
	}

	s.logger.Info("Portfolio deleted", zap.Int("portfolio_id", portfolioID))
	return nil
}