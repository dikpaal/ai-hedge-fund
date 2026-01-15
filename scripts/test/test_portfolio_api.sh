#!/bin/bash

echo "=== Testing Portfolio Service API ==="
echo ""

echo "1. Health Check"
curl -s http://localhost:8081/health | jq
echo ""

echo "2. Create Portfolio"
PORTFOLIO=$(curl -s -X POST http://localhost:8081/api/v1/portfolios \
  -H 'Content-Type: application/json' \
  -d '{"user_id": 1, "name": "Test Portfolio", "initial_cash": 100000}')
echo $PORTFOLIO | jq
PORTFOLIO_ID=$(echo $PORTFOLIO | jq -r '.id')
echo "Created Portfolio ID: $PORTFOLIO_ID"
echo ""

echo "3. Buy AAPL (10 shares)"
curl -s -X POST http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/trades \
  -H 'Content-Type: application/json' \
  -d '{"symbol": "AAPL", "side": "buy", "quantity": 10, "order_type": "market"}' | jq
echo ""

echo "4. Buy GOOGL (5 shares)"
curl -s -X POST http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/trades \
  -H 'Content-Type: application/json' \
  -d '{"symbol": "GOOGL", "side": "buy", "quantity": 5, "order_type": "market"}' | jq
echo ""

echo "5. Buy MSFT (8 shares)"
curl -s -X POST http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/trades \
  -H 'Content-Type: application/json' \
  -d '{"symbol": "MSFT", "side": "buy", "quantity": 8, "order_type": "market"}' | jq
echo ""

echo "6. Get Positions"
curl -s http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/positions | jq
echo ""

echo "7. Get Summary"
curl -s http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/summary | jq
echo ""

echo "8. Get Allocation"
curl -s http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/allocation | jq
echo ""

echo "9. Get Risk Metrics"
curl -s http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/risk | jq
echo ""

echo "10. Sell AAPL (5 shares)"
curl -s -X POST http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/trades \
  -H 'Content-Type: application/json' \
  -d '{"symbol": "AAPL", "side": "sell", "quantity": 5, "order_type": "market"}' | jq
echo ""

echo "11. Get Updated Positions"
curl -s http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/positions | jq
echo ""

echo "12. Trade History"
curl -s http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/trades | jq
echo ""

echo "13. Get Rebalance Recommendations"
curl -s -X POST http://localhost:8081/api/v1/portfolios/$PORTFOLIO_ID/rebalance \
  -H 'Content-Type: application/json' \
  -d '{"target_allocations": {"AAPL": 30, "GOOGL": 30, "MSFT": 40}}' | jq
echo ""

echo "=== All Tests Complete ==="
