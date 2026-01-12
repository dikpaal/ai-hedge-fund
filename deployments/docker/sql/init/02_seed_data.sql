-- Seed data for development environment
-- This file creates initial data for testing and development

-- Insert default admin user
INSERT INTO users (username, email, password_hash, full_name, role) VALUES
('admin', 'admin@hedgefund.com', '$2a$10$N.zmdr9VgKs1HY.T9V1YG.FxA.h9XOCR.qg5XTI5bgtGIDEL7C4pu', 'Administrator', 'admin'),
('trader1', 'trader1@hedgefund.com', '$2a$10$N.zmdr9VgKs1HY.T9V1YG.FxA.h9XOCR.qg5XTI5bgtGIDEL7C4pu', 'John Trader', 'trader'),
('analyst1', 'analyst1@hedgefund.com', '$2a$10$N.zmdr9VgKs1HY.T9V1YG.FxA.h9XOCR.qg5XTI5bgtGIDEL7C4pu', 'Jane Analyst', 'analyst');

-- Note: Password hash is for 'password123' - DO NOT use in production

-- Create default portfolios
INSERT INTO portfolios (user_id, name, cash, margin_available) VALUES
((SELECT id FROM users WHERE username = 'admin'), 'Admin Portfolio', 1000000.00, 500000.00),
((SELECT id FROM users WHERE username = 'trader1'), 'Main Trading Portfolio', 500000.00, 250000.00),
((SELECT id FROM users WHERE username = 'analyst1'), 'Analysis Portfolio', 100000.00, 50000.00);

-- Insert default risk limits
INSERT INTO risk_limits (user_id, max_position_size, max_daily_loss, max_portfolio_risk, max_leverage, max_concentration, stop_loss_percentage) VALUES
((SELECT id FROM users WHERE username = 'admin'), 100000.00, 50000.00, 0.20, 2.0, 0.15, 0.10),
((SELECT id FROM users WHERE username = 'trader1'), 50000.00, 25000.00, 0.15, 1.5, 0.10, 0.08),
((SELECT id FROM users WHERE username = 'analyst1'), 10000.00, 5000.00, 0.10, 1.2, 0.05, 0.05);

-- Add some popular stocks to watchlists
INSERT INTO watchlists (user_id, symbol, name, alert_enabled) VALUES
((SELECT id FROM users WHERE username = 'admin'), 'AAPL', 'Apple Inc.', false),
((SELECT id FROM users WHERE username = 'admin'), 'GOOGL', 'Alphabet Inc.', false),
((SELECT id FROM users WHERE username = 'admin'), 'MSFT', 'Microsoft Corp.', false),
((SELECT id FROM users WHERE username = 'admin'), 'NVDA', 'NVIDIA Corp.', false),
((SELECT id FROM users WHERE username = 'admin'), 'TSLA', 'Tesla Inc.', false),
((SELECT id FROM users WHERE username = 'trader1'), 'AAPL', 'Apple Inc.', false),
((SELECT id FROM users WHERE username = 'trader1'), 'MSFT', 'Microsoft Corp.', false),
((SELECT id FROM users WHERE username = 'trader1'), 'NVDA', 'NVIDIA Corp.', false);

-- Insert some sample market data
INSERT INTO market_prices (symbol, open, high, low, close, volume, timestamp, source) VALUES
('AAPL', 185.50, 187.25, 184.80, 186.95, 45123456, NOW() - INTERVAL '1 day', 'api'),
('GOOGL', 145.20, 147.80, 144.50, 146.75, 23456789, NOW() - INTERVAL '1 day', 'api'),
('MSFT', 378.90, 382.15, 377.25, 380.50, 34567890, NOW() - INTERVAL '1 day', 'api'),
('NVDA', 725.40, 742.80, 720.15, 738.65, 67890123, NOW() - INTERVAL '1 day', 'api'),
('TSLA', 248.75, 255.20, 246.30, 252.80, 78901234, NOW() - INTERVAL '1 day', 'api');

-- Insert today's prices
INSERT INTO market_prices (symbol, open, high, low, close, volume, timestamp, source) VALUES
('AAPL', 186.95, 189.40, 185.60, 188.25, 42345678, NOW(), 'api'),
('GOOGL', 146.75, 149.20, 145.80, 147.90, 21234567, NOW(), 'api'),
('MSFT', 380.50, 384.75, 379.20, 382.30, 32123456, NOW(), 'api'),
('NVDA', 738.65, 755.90, 735.20, 748.40, 65432109, NOW(), 'api'),
('TSLA', 252.80, 259.45, 250.15, 256.70, 76543210, NOW(), 'api');

-- Insert some sample news
INSERT INTO news_items (symbol, title, summary, source, sentiment, sentiment_score, published_at) VALUES
('AAPL', 'Apple Reports Strong Q4 Earnings', 'Apple Inc. reported better-than-expected earnings for Q4, driven by strong iPhone sales and services revenue.', 'Reuters', 'positive', 0.75, NOW() - INTERVAL '2 hours'),
('GOOGL', 'Alphabet Invests Heavily in AI Infrastructure', 'Google parent company increases AI spending, focusing on cloud services and machine learning capabilities.', 'Bloomberg', 'positive', 0.60, NOW() - INTERVAL '4 hours'),
('TSLA', 'Tesla Faces Production Challenges', 'Tesla reports lower than expected delivery numbers for the quarter due to supply chain issues.', 'Wall Street Journal', 'negative', -0.45, NOW() - INTERVAL '6 hours'),
('MSFT', 'Microsoft Azure Growth Continues', 'Microsoft cloud services show continued strong growth, beating analyst expectations.', 'CNBC', 'positive', 0.65, NOW() - INTERVAL '8 hours'),
('NVDA', 'NVIDIA AI Chip Demand Remains Strong', 'Demand for NVIDIA AI chips continues to outpace supply, driving revenue growth.', 'TechCrunch', 'positive', 0.80, NOW() - INTERVAL '10 hours');

-- Initialize agent performance tracking
INSERT INTO agent_performance (agent_name, symbol, period, total_signals, correct_signals, accuracy, avg_return) VALUES
('warren_buffett', 'AAPL', '1m', 15, 12, 0.80, 0.045),
('warren_buffett', 'MSFT', '1m', 12, 10, 0.83, 0.052),
('michael_burry', 'TSLA', '1m', 8, 6, 0.75, -0.023),
('cathie_wood', 'NVDA', '1m', 20, 16, 0.80, 0.087),
('technical_analyst', 'GOOGL', '1m', 25, 18, 0.72, 0.031);

-- Sample AI signals
INSERT INTO ai_signals (agent_name, symbol, signal, confidence, reasoning, price) VALUES
('warren_buffett', 'AAPL', 'hold', 75.0, 'Strong fundamentals but fairly valued at current levels', 188.25),
('michael_burry', 'TSLA', 'sell', 85.0, 'Overvalued based on traditional metrics, market correction likely', 256.70),
('cathie_wood', 'NVDA', 'buy', 90.0, 'AI revolution is just beginning, strong growth potential', 748.40),
('technical_analyst', 'MSFT', 'buy', 65.0, 'Bullish technical indicators, breaking resistance levels', 382.30),
('warren_buffett', 'GOOGL', 'buy', 70.0, 'Strong moat in search and growing cloud business', 147.90);