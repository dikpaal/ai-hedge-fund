package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"hedge-fund/pkg/shared/database"
	"hedge-fund/pkg/shared/models"
	"go.uber.org/zap"
)

type PortfolioRepository struct {
	db     *database.DB
	logger *zap.Logger
}

func NewPortfolioRepository(db *database.DB, logger *zap.Logger) *PortfolioRepository {
	return &PortfolioRepository{
		db:     db,
		logger: logger,
	}
}

// Portfolio CRUD Operations

// CreatePortfolio creates a new portfolio
func (r *PortfolioRepository) CreatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	query := `
		INSERT INTO portfolios (user_id, cash, margin_used, margin_available, total_value,
		                       unrealized_pnl, realized_pnl, day_pnl, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		portfolio.UserID,
		portfolio.Cash,
		portfolio.MarginUsed,
		portfolio.MarginAvailable,
		portfolio.TotalValue,
		portfolio.UnrealizedPnL,
		portfolio.RealizedPnL,
		portfolio.DayPnL,
		now,
		now,
	).Scan(&portfolio.ID)

	if err != nil {
		r.logger.Error("Failed to create portfolio", zap.Error(err), zap.Int("user_id", portfolio.UserID))
		return fmt.Errorf("failed to create portfolio: %w", err)
	}

	portfolio.CreatedAt = now
	portfolio.UpdatedAt = now

	r.logger.Info("Portfolio created successfully",
		zap.Int("portfolio_id", portfolio.ID),
		zap.Int("user_id", portfolio.UserID))

	return nil
}

// GetPortfolioByID retrieves a portfolio by ID with all positions
func (r *PortfolioRepository) GetPortfolioByID(ctx context.Context, portfolioID int) (*models.Portfolio, error) {
	query := `
		SELECT id, user_id, cash, margin_used, margin_available, total_value,
		       unrealized_pnl, realized_pnl, day_pnl, created_at, updated_at
		FROM portfolios
		WHERE id = $1`

	portfolio := &models.Portfolio{}
	err := r.db.QueryRowContext(ctx, query, portfolioID).Scan(
		&portfolio.ID,
		&portfolio.UserID,
		&portfolio.Cash,
		&portfolio.MarginUsed,
		&portfolio.MarginAvailable,
		&portfolio.TotalValue,
		&portfolio.UnrealizedPnL,
		&portfolio.RealizedPnL,
		&portfolio.DayPnL,
		&portfolio.CreatedAt,
		&portfolio.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("portfolio not found: %d", portfolioID)
		}
		r.logger.Error("Failed to get portfolio", zap.Error(err), zap.Int("portfolio_id", portfolioID))
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	// Load positions
	positions, err := r.GetPositionsByPortfolioID(ctx, portfolioID)
	if err != nil {
		r.logger.Error("Failed to load positions for portfolio", zap.Error(err), zap.Int("portfolio_id", portfolioID))
		return nil, fmt.Errorf("failed to load positions: %w", err)
	}
	portfolio.Positions = positions

	return portfolio, nil
}

// GetPortfoliosByUserID retrieves all portfolios for a user
func (r *PortfolioRepository) GetPortfoliosByUserID(ctx context.Context, userID int) ([]models.Portfolio, error) {
	query := `
		SELECT id, user_id, cash, margin_used, margin_available, total_value,
		       unrealized_pnl, realized_pnl, day_pnl, created_at, updated_at
		FROM portfolios
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to get portfolios for user", zap.Error(err), zap.Int("user_id", userID))
		return nil, fmt.Errorf("failed to get portfolios: %w", err)
	}
	defer rows.Close()

	var portfolios []models.Portfolio
	for rows.Next() {
		portfolio := models.Portfolio{}
		err := rows.Scan(
			&portfolio.ID,
			&portfolio.UserID,
			&portfolio.Cash,
			&portfolio.MarginUsed,
			&portfolio.MarginAvailable,
			&portfolio.TotalValue,
			&portfolio.UnrealizedPnL,
			&portfolio.RealizedPnL,
			&portfolio.DayPnL,
			&portfolio.CreatedAt,
			&portfolio.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan portfolio", zap.Error(err))
			continue
		}

		// Load positions for each portfolio
		positions, err := r.GetPositionsByPortfolioID(ctx, portfolio.ID)
		if err != nil {
			r.logger.Error("Failed to load positions", zap.Error(err), zap.Int("portfolio_id", portfolio.ID))
			continue
		}
		portfolio.Positions = positions

		portfolios = append(portfolios, portfolio)
	}

	return portfolios, nil
}

// UpdatePortfolio updates an existing portfolio
func (r *PortfolioRepository) UpdatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	query := `
		UPDATE portfolios
		SET cash = $2, margin_used = $3, margin_available = $4, total_value = $5,
		    unrealized_pnl = $6, realized_pnl = $7, day_pnl = $8, updated_at = $9
		WHERE id = $1`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		portfolio.ID,
		portfolio.Cash,
		portfolio.MarginUsed,
		portfolio.MarginAvailable,
		portfolio.TotalValue,
		portfolio.UnrealizedPnL,
		portfolio.RealizedPnL,
		portfolio.DayPnL,
		now,
	)

	if err != nil {
		r.logger.Error("Failed to update portfolio", zap.Error(err), zap.Int("portfolio_id", portfolio.ID))
		return fmt.Errorf("failed to update portfolio: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("portfolio not found: %d", portfolio.ID)
	}

	portfolio.UpdatedAt = now

	r.logger.Info("Portfolio updated successfully", zap.Int("portfolio_id", portfolio.ID))
	return nil
}

// DeletePortfolio deletes a portfolio and all its positions
func (r *PortfolioRepository) DeletePortfolio(ctx context.Context, portfolioID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete positions first (foreign key constraint)
	_, err = tx.ExecContext(ctx, "DELETE FROM positions WHERE user_id = (SELECT user_id FROM portfolios WHERE id = $1)", portfolioID)
	if err != nil {
		r.logger.Error("Failed to delete positions", zap.Error(err), zap.Int("portfolio_id", portfolioID))
		return fmt.Errorf("failed to delete positions: %w", err)
	}

	// Delete portfolio
	result, err := tx.ExecContext(ctx, "DELETE FROM portfolios WHERE id = $1", portfolioID)
	if err != nil {
		r.logger.Error("Failed to delete portfolio", zap.Error(err), zap.Int("portfolio_id", portfolioID))
		return fmt.Errorf("failed to delete portfolio: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("portfolio not found: %d", portfolioID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Portfolio deleted successfully", zap.Int("portfolio_id", portfolioID))
	return nil
}

// Position CRUD Operations

// CreatePosition creates a new position
func (r *PortfolioRepository) CreatePosition(ctx context.Context, position *models.Position) error {
	query := `
		INSERT INTO positions (user_id, symbol, quantity, side, entry_price, current_price,
		                      unrealized_pnl, realized_pnl, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		position.UserID,
		position.Symbol,
		position.Quantity,
		position.Side,
		position.EntryPrice,
		position.CurrentPrice,
		position.UnrealizedPnL,
		position.RealizedPnL,
		now,
		now,
	).Scan(&position.ID)

	if err != nil {
		r.logger.Error("Failed to create position", zap.Error(err),
			zap.Int("user_id", position.UserID), zap.String("symbol", position.Symbol))
		return fmt.Errorf("failed to create position: %w", err)
	}

	position.CreatedAt = now
	position.UpdatedAt = now

	r.logger.Info("Position created successfully",
		zap.Int("position_id", position.ID),
		zap.String("symbol", position.Symbol),
		zap.Int64("quantity", position.Quantity))

	return nil
}

// GetPositionByID retrieves a position by ID
func (r *PortfolioRepository) GetPositionByID(ctx context.Context, positionID int) (*models.Position, error) {
	query := `
		SELECT id, user_id, symbol, quantity, side, entry_price, current_price,
		       unrealized_pnl, realized_pnl, created_at, updated_at
		FROM positions
		WHERE id = $1`

	position := &models.Position{}
	err := r.db.QueryRowContext(ctx, query, positionID).Scan(
		&position.ID,
		&position.UserID,
		&position.Symbol,
		&position.Quantity,
		&position.Side,
		&position.EntryPrice,
		&position.CurrentPrice,
		&position.UnrealizedPnL,
		&position.RealizedPnL,
		&position.CreatedAt,
		&position.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("position not found: %d", positionID)
		}
		r.logger.Error("Failed to get position", zap.Error(err), zap.Int("position_id", positionID))
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	return position, nil
}

// GetPositionsByPortfolioID retrieves all positions for a portfolio
func (r *PortfolioRepository) GetPositionsByPortfolioID(ctx context.Context, portfolioID int) ([]models.Position, error) {
	query := `
		SELECT p.id, p.user_id, p.symbol, p.quantity, p.side, p.entry_price, p.current_price,
		       p.unrealized_pnl, p.realized_pnl, p.created_at, p.updated_at
		FROM positions p
		JOIN portfolios pf ON p.user_id = pf.user_id
		WHERE pf.id = $1
		ORDER BY p.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		r.logger.Error("Failed to get positions for portfolio", zap.Error(err), zap.Int("portfolio_id", portfolioID))
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	defer rows.Close()

	var positions []models.Position
	for rows.Next() {
		position := models.Position{}
		err := rows.Scan(
			&position.ID,
			&position.UserID,
			&position.Symbol,
			&position.Quantity,
			&position.Side,
			&position.EntryPrice,
			&position.CurrentPrice,
			&position.UnrealizedPnL,
			&position.RealizedPnL,
			&position.CreatedAt,
			&position.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan position", zap.Error(err))
			continue
		}
		positions = append(positions, position)
	}

	return positions, nil
}

// GetPositionByUserAndSymbol retrieves a specific position by user and symbol
func (r *PortfolioRepository) GetPositionByUserAndSymbol(ctx context.Context, userID int, symbol string) (*models.Position, error) {
	query := `
		SELECT id, user_id, symbol, quantity, side, entry_price, current_price,
		       unrealized_pnl, realized_pnl, created_at, updated_at
		FROM positions
		WHERE user_id = $1 AND symbol = $2`

	position := &models.Position{}
	err := r.db.QueryRowContext(ctx, query, userID, symbol).Scan(
		&position.ID,
		&position.UserID,
		&position.Symbol,
		&position.Quantity,
		&position.Side,
		&position.EntryPrice,
		&position.CurrentPrice,
		&position.UnrealizedPnL,
		&position.RealizedPnL,
		&position.CreatedAt,
		&position.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Position doesn't exist, which is valid
		}
		r.logger.Error("Failed to get position by user and symbol",
			zap.Error(err), zap.Int("user_id", userID), zap.String("symbol", symbol))
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	return position, nil
}

// UpdatePosition updates an existing position
func (r *PortfolioRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	query := `
		UPDATE positions
		SET quantity = $2, side = $3, entry_price = $4, current_price = $5,
		    unrealized_pnl = $6, realized_pnl = $7, updated_at = $8
		WHERE id = $1`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		position.ID,
		position.Quantity,
		position.Side,
		position.EntryPrice,
		position.CurrentPrice,
		position.UnrealizedPnL,
		position.RealizedPnL,
		now,
	)

	if err != nil {
		r.logger.Error("Failed to update position", zap.Error(err), zap.Int("position_id", position.ID))
		return fmt.Errorf("failed to update position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("position not found: %d", position.ID)
	}

	position.UpdatedAt = now

	r.logger.Info("Position updated successfully",
		zap.Int("position_id", position.ID), zap.String("symbol", position.Symbol))
	return nil
}

// DeletePosition deletes a position
func (r *PortfolioRepository) DeletePosition(ctx context.Context, positionID int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM positions WHERE id = $1", positionID)
	if err != nil {
		r.logger.Error("Failed to delete position", zap.Error(err), zap.Int("position_id", positionID))
		return fmt.Errorf("failed to delete position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("position not found: %d", positionID)
	}

	r.logger.Info("Position deleted successfully", zap.Int("position_id", positionID))
	return nil
}

// Trade CRUD Operations

// CreateTrade creates a new trade record
func (r *PortfolioRepository) CreateTrade(ctx context.Context, trade *models.Trade) error {
	query := `
		INSERT INTO trades (user_id, position_id, symbol, quantity, price, side, type, status,
		                   fees, executed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		trade.UserID,
		trade.PositionID,
		trade.Symbol,
		trade.Quantity,
		trade.Price,
		trade.Side,
		trade.Type,
		trade.Status,
		trade.Fees,
		trade.ExecutedAt,
		now,
	).Scan(&trade.ID)

	if err != nil {
		r.logger.Error("Failed to create trade", zap.Error(err),
			zap.Int("user_id", trade.UserID), zap.String("symbol", trade.Symbol))
		return fmt.Errorf("failed to create trade: %w", err)
	}

	trade.CreatedAt = now

	r.logger.Info("Trade created successfully",
		zap.Int("trade_id", trade.ID),
		zap.String("symbol", trade.Symbol),
		zap.String("side", trade.Side),
		zap.Int64("quantity", trade.Quantity),
		zap.Float64("price", trade.Price))

	return nil
}

// GetTradesByUserID retrieves all trades for a user
func (r *PortfolioRepository) GetTradesByUserID(ctx context.Context, userID int, limit int, offset int) ([]models.Trade, error) {
	query := `
		SELECT id, user_id, position_id, symbol, quantity, price, side, type, status,
		       fees, executed_at, created_at
		FROM trades
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get trades for user", zap.Error(err), zap.Int("user_id", userID))
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		trade := models.Trade{}
		err := rows.Scan(
			&trade.ID,
			&trade.UserID,
			&trade.PositionID,
			&trade.Symbol,
			&trade.Quantity,
			&trade.Price,
			&trade.Side,
			&trade.Type,
			&trade.Status,
			&trade.Fees,
			&trade.ExecutedAt,
			&trade.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan trade", zap.Error(err))
			continue
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetTradesBySymbol retrieves all trades for a specific symbol
func (r *PortfolioRepository) GetTradesBySymbol(ctx context.Context, userID int, symbol string, limit int, offset int) ([]models.Trade, error) {
	query := `
		SELECT id, user_id, position_id, symbol, quantity, price, side, type, status,
		       fees, executed_at, created_at
		FROM trades
		WHERE user_id = $1 AND symbol = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, userID, symbol, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get trades for symbol", zap.Error(err),
			zap.Int("user_id", userID), zap.String("symbol", symbol))
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		trade := models.Trade{}
		err := rows.Scan(
			&trade.ID,
			&trade.UserID,
			&trade.PositionID,
			&trade.Symbol,
			&trade.Quantity,
			&trade.Price,
			&trade.Side,
			&trade.Type,
			&trade.Status,
			&trade.Fees,
			&trade.ExecutedAt,
			&trade.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan trade", zap.Error(err))
			continue
		}
		trades = append(trades, trade)
	}

	return trades, nil
}