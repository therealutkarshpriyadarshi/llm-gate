# LLM Gateway 🚀

An intelligent, high-performance LLM gateway implementing semantic caching, multi-provider routing, prompt management, and cost optimization.

## 🎯 Key Features

- **Semantic Caching**: 40-60% cost reduction through embedding-based cache matching
- **Multi-Provider Support**: OpenAI, Anthropic, Azure, AWS Bedrock, Google Vertex AI
- **Intelligent Routing**: Query analysis, cost-based routing, load balancing
- **Prompt Management**: Versioning, A/B testing, rollback capabilities
- **Cost Optimization**: Token tracking, budget enforcement, spending alerts
- **Enterprise-Grade**: Rate limiting, PII redaction, full observability
- **High Performance**: Built in Go for maximum throughput

## 🏗️ Architecture

```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│      LLM Gateway (Port 8080)    │
│  ┌──────────────────────────┐   │
│  │   Semantic Cache Layer   │   │
│  └──────────────────────────┘   │
│  ┌──────────────────────────┐   │
│  │  Intelligent Router      │   │
│  └──────────────────────────┘   │
│  ┌──────────────────────────┐   │
│  │  Provider Abstraction    │   │
│  └──────────────────────────┘   │
└─────────────┬───────────────────┘
              │
    ┌─────────┼─────────┐
    ▼         ▼         ▼
┌────────┐ ┌────────┐ ┌────────┐
│ OpenAI │ │Claude  │ │ Azure  │
└────────┘ └────────┘ └────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- API keys for LLM providers

### Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/llm-gate.git
cd llm-gate

# Start infrastructure (Redis, PostgreSQL, Prometheus, Grafana)
docker-compose up -d

# Copy environment variables
cp .env.example .env
# Edit .env and add your API keys

# Install dependencies
go mod download

# Run the gateway
go run cmd/gateway/main.go
```

The gateway will start on `http://localhost:8080`.

### Test It

```bash
# Health check
curl http://localhost:8080/health

# Send a chat completion request
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## 📊 Dashboards

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686
- **RedisInsight**: http://localhost:8001

## 📚 Documentation

- [Roadmap](ROADMAP.md) - Development phases and timeline
- [Getting Started](GETTING_STARTED.md) - Detailed setup guide
- [Project Structure](PROJECT_STRUCTURE.md) - Code organization
- [API Documentation](docs/api.md) - API reference (coming soon)
- [Architecture](docs/architecture.md) - System design (coming soon)

## 🎓 Learning Outcomes

This project demonstrates expertise in:

- High-throughput proxy architecture design
- Semantic similarity and vector search
- Multi-provider abstraction patterns
- Distributed caching strategies
- Rate limiting and cost optimization
- OpenTelemetry observability
- Production-ready Go services

## 🛣️ Roadmap

- [x] Project initialization
- [x] Phase 1: Foundation & Architecture (Week 1-2) ✅
- [x] Phase 2: Basic Proxy & OpenAI Provider (Week 2-3) ✅
- [x] Phase 3: Multi-Provider Support (Week 3-4) ✅
- [x] Phase 4: Semantic Caching (Week 4-6) ✅
- [x] Phase 5: Intelligent Routing (Week 6-7) ✅
- [x] Phase 6: Prompt Management (Week 7-8) ✅
- [x] Phase 7: Cost Optimization (Week 8-9) ✅
- [ ] Phase 8: Observability (Week 9-10)
- [ ] Phase 9: Security & Compliance (Week 10-11)
- [ ] Phase 10: Production Readiness (Week 11-12)

See [ROADMAP.md](ROADMAP.md) for detailed milestones.

## 🔧 Tech Stack

- **Language**: Go 1.21+
- **HTTP Framework**: Chi
- **Cache**: Redis Stack (with vector search)
- **Database**: PostgreSQL
- **Tracing**: OpenTelemetry + Jaeger
- **Metrics**: Prometheus
- **Dashboards**: Grafana
- **Deployment**: Docker, Kubernetes

## 🌟 Inspiration

This project is inspired by production-grade LLM infrastructure:

- **Kong AI Gateway**: Load balancing algorithms
- **LiteLLM**: Multi-provider abstraction patterns
- **Portkey**: Prompt management and governance

## 📈 Performance Targets

- **Latency**: p50 < 100ms overhead, p99 < 500ms
- **Throughput**: 10,000+ req/sec
- **Availability**: 99.9% uptime
- **Cache Hit Rate**: >40% (target 60%)
- **Cost Reduction**: 40-60% through intelligent routing

## 🤝 Contributing

Contributions are welcome! Please check out the [ROADMAP.md](ROADMAP.md) for areas to contribute.

## 📝 License

MIT License - see LICENSE file for details

## 🙋 Questions?

- Review the [Getting Started Guide](GETTING_STARTED.md)
- Check the [Roadmap](ROADMAP.md) for implementation details
- Open an issue for bugs or feature requests

---

**Built with ❤️ for the LLM infrastructure community**
