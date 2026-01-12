# AI Hedge Fund - Hybrid Architecture

A modern AI-powered hedge fund system built with Go backend services, Python AI agents, and a rich CLI interface.

## Architecture

- **Go Backend Services**: High-performance portfolio, risk, and market data services
- **Python AI Service**: LangGraph-powered multi-agent investment analysis
- **Go CLI**: Rich terminal interface for interacting with the system
- **Cloud-Native**: Kubernetes deployment with monitoring and observability

## Services

- **API Gateway** (Go): Authentication, routing, rate limiting
- **Portfolio Service** (Go): Position tracking, P&L calculations
- **Risk Service** (Go): Volatility analysis, position sizing
- **Market Data Service** (Go): Real-time price feeds, caching
- **AI Agent Service** (Python): Multi-agent investment analysis using LangGraph

## Project Structure

```
hedge-fund/
├── cmd/cli/              # Go CLI application
├── internal/
│   ├── api/              # API Gateway
│   ├── portfolio/        # Portfolio service
│   ├── risk/            # Risk service
│   ├── market/          # Market data service
│   └── shared/          # Common Go packages
├── ai-service/          # Python AI service
├── deployments/         # K8s, Docker, Helm
├── monitoring/          # Prometheus, Grafana
└── scripts/            # Build and deployment scripts
```

## Getting Started

Coming soon...