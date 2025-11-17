# Project Structure

## Directory Layout

```
llm-gate/
├── cmd/
│   └── gateway/
│       └── main.go                 # Application entry point
│
├── internal/                       # Private application code
│   ├── api/                        # HTTP handlers and routing
│   │   ├── router.go              # Main router setup
│   │   ├── middleware/            # Custom middleware
│   │   │   ├── auth.go
│   │   │   ├── ratelimit.go
│   │   │   └── logging.go
│   │   └── handlers/              # HTTP handlers
│   │       ├── chat.go
│   │       ├── embeddings.go
│   │       ├── health.go
│   │       └── admin.go
│   │
│   ├── core/                      # Business logic
│   │   ├── models/                # Domain models
│   │   │   ├── request.go
│   │   │   ├── response.go
│   │   │   ├── provider.go
│   │   │   └── cache.go
│   │   ├── services/              # Business services
│   │   │   ├── completion.go
│   │   │   ├── embedding.go
│   │   │   └── cost_tracker.go
│   │   └── interfaces/            # Interface definitions
│   │       └── provider.go
│   │
│   ├── providers/                 # LLM provider implementations
│   │   ├── registry.go            # Provider registry
│   │   ├── base.go                # Base provider interface
│   │   ├── openai/
│   │   │   ├── client.go
│   │   │   ├── chat.go
│   │   │   ├── embeddings.go
│   │   │   └── models.go
│   │   ├── anthropic/
│   │   │   ├── client.go
│   │   │   ├── chat.go
│   │   │   └── models.go
│   │   ├── azure/
│   │   │   └── ...
│   │   ├── bedrock/
│   │   │   └── ...
│   │   └── vertex/
│   │       └── ...
│   │
│   ├── cache/                     # Caching layer
│   │   ├── cache.go               # Cache interface
│   │   ├── redis.go               # Redis implementation
│   │   ├── semantic.go            # Semantic caching logic
│   │   ├── embedder.go            # Embedding generation
│   │   └── similarity.go          # Vector similarity matching
│   │
│   ├── routing/                   # Request routing logic
│   │   ├── router.go              # Main routing logic
│   │   ├── strategies/            # Routing strategies
│   │   │   ├── roundrobin.go
│   │   │   ├── least_latency.go
│   │   │   ├── cost_optimized.go
│   │   │   └── intelligent.go
│   │   ├── analyzer.go            # Query analysis
│   │   ├── loadbalancer.go        # Load balancing
│   │   └── fallback.go            # Fallback logic
│   │
│   ├── ratelimit/                 # Rate limiting
│   │   ├── limiter.go             # Rate limiter interface
│   │   ├── token_bucket.go        # Token bucket implementation
│   │   └── distributed.go         # Redis-based distributed limiter
│   │
│   ├── telemetry/                 # Observability
│   │   ├── tracing.go             # OpenTelemetry tracing
│   │   ├── metrics.go             # Prometheus metrics
│   │   ├── logging.go             # Structured logging
│   │   └── exporter.go            # Metrics exporters
│   │
│   ├── storage/                   # Data persistence
│   │   ├── postgres.go            # PostgreSQL client
│   │   ├── repositories/          # Data repositories
│   │   │   ├── prompt.go
│   │   │   ├── user.go
│   │   │   └── analytics.go
│   │   └── migrations/            # Database migrations
│   │       └── ...
│   │
│   ├── prompts/                   # Prompt management
│   │   ├── manager.go             # Prompt lifecycle manager
│   │   ├── template.go            # Template engine
│   │   ├── versioning.go          # Version control
│   │   └── abtesting.go           # A/B testing framework
│   │
│   ├── security/                  # Security features
│   │   ├── auth.go                # Authentication
│   │   ├── rbac.go                # Role-based access control
│   │   ├── pii.go                 # PII detection & redaction
│   │   └── encryption.go          # Encryption utilities
│   │
│   └── config/                    # Configuration management
│       ├── config.go              # Config structure
│       ├── loader.go              # Config loader (Viper)
│       └── validator.go           # Config validation
│
├── pkg/                           # Public packages (can be imported)
│   ├── errors/                    # Custom error types
│   │   └── errors.go
│   ├── utils/                     # Utility functions
│   │   ├── retry.go
│   │   ├── http.go
│   │   └── json.go
│   └── client/                    # Go SDK for the gateway
│       └── client.go
│
├── api/                           # API definitions
│   ├── openapi.yaml               # OpenAPI/Swagger spec
│   └── protobuf/                  # Protocol buffer definitions (if using gRPC)
│       └── ...
│
├── configs/                       # Configuration files
│   ├── config.yaml                # Main configuration
│   ├── config.dev.yaml            # Development config
│   ├── config.prod.yaml           # Production config
│   ├── prometheus.yml             # Prometheus config
│   └── grafana/                   # Grafana dashboards
│       ├── dashboards/
│       │   ├── overview.json
│       │   ├── providers.json
│       │   └── costs.json
│       └── datasources/
│           └── prometheus.yaml
│
├── scripts/                       # Build and deployment scripts
│   ├── build.sh                   # Build script
│   ├── test.sh                    # Test runner
│   ├── deploy.sh                  # Deployment script
│   ├── migrate.sh                 # Database migration script
│   └── load_test.sh               # Load testing script
│
├── tests/                         # Test files
│   ├── integration/               # Integration tests
│   │   ├── cache_test.go
│   │   ├── providers_test.go
│   │   └── routing_test.go
│   ├── e2e/                       # End-to-end tests
│   │   └── gateway_test.go
│   └── fixtures/                  # Test fixtures
│       ├── requests.json
│       └── responses.json
│
├── deployments/                   # Deployment configurations
│   ├── docker/
│   │   └── Dockerfile
│   ├── kubernetes/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── configmap.yaml
│   │   ├── secret.yaml
│   │   └── ingress.yaml
│   ├── helm/                      # Helm charts
│   │   └── llm-gateway/
│   │       ├── Chart.yaml
│   │       ├── values.yaml
│   │       └── templates/
│   └── terraform/                 # Infrastructure as Code
│       └── ...
│
├── docs/                          # Documentation
│   ├── architecture.md            # Architecture overview
│   ├── api.md                     # API documentation
│   ├── deployment.md              # Deployment guide
│   ├── development.md             # Development guide
│   └── runbooks/                  # Operational runbooks
│       ├── high_latency.md
│       ├── cache_miss.md
│       └── provider_down.md
│
├── examples/                      # Example usage
│   ├── basic_request.go
│   ├── streaming.go
│   ├── semantic_cache.go
│   └── multi_provider.go
│
├── .github/                       # GitHub configuration
│   ├── workflows/                 # GitHub Actions
│   │   ├── ci.yml
│   │   ├── cd.yml
│   │   └── security.yml
│   └── ISSUE_TEMPLATE/
│       └── bug_report.md
│
├── docker-compose.yml             # Local development environment
├── Makefile                       # Build automation
├── go.mod                         # Go module definition
├── go.sum                         # Go module checksums
├── .env.example                   # Example environment variables
├── .gitignore                     # Git ignore rules
├── .golangci.yml                  # Linter configuration
├── README.md                      # Project README
├── ROADMAP.md                     # Project roadmap
├── GETTING_STARTED.md             # Getting started guide
├── CHANGELOG.md                   # Change log
└── LICENSE                        # License file
```

---

## Key Design Decisions

### 1. Clean Architecture Layers

```
┌─────────────────────────────────────────┐
│          API Layer (cmd, api)           │  ← HTTP handlers, routing
├─────────────────────────────────────────┤
│       Business Logic (core)             │  ← Domain models, services
├─────────────────────────────────────────┤
│   Infrastructure (providers, cache,     │  ← External integrations
│    storage, telemetry)                  │
└─────────────────────────────────────────┘
```

**Benefits:**
- Clear separation of concerns
- Easy to test each layer independently
- Flexibility to swap implementations

### 2. Internal vs Pkg

**`internal/`:**
- Cannot be imported by external projects
- Contains application-specific logic
- Most of your code lives here

**`pkg/`:**
- Can be imported by other Go projects
- Contains reusable utilities
- SDK for users of your gateway

### 3. Provider Abstraction

Each provider has its own package with:
- `client.go`: HTTP client and auth
- `chat.go`: Chat completion implementation
- `embeddings.go`: Embedding implementation (if supported)
- `models.go`: Provider-specific request/response models

All providers implement the common interface defined in `internal/core/interfaces/provider.go`.

### 4. Configuration Management

Three-tier configuration:
1. **Default values** in `config.go`
2. **Config files** (YAML) for environment-specific settings
3. **Environment variables** for secrets and overrides

Priority: Environment Variables > Config File > Defaults

### 5. Telemetry First

Every major component has built-in observability:
- **Traces**: Request flow across components
- **Metrics**: Performance and business metrics
- **Logs**: Structured JSON logs with context

---

## Module Organization Principles

### 1. Single Responsibility
Each package has one clear purpose:
- `cache/` handles caching, nothing else
- `routing/` handles request routing logic
- `providers/` implement LLM provider integrations

### 2. Dependency Direction
Dependencies flow inward:
```
api → core ← providers
            ← cache
            ← routing
```

Core has no knowledge of API layer or specific providers.

### 3. Interface-Based Design

Core defines interfaces:
```go
// internal/core/interfaces/provider.go
type Provider interface {
    SendRequest(ctx context.Context, req *Request) (*Response, error)
}
```

Providers implement interfaces:
```go
// internal/providers/openai/client.go
type Client struct { ... }

func (c *Client) SendRequest(ctx context.Context, req *Request) (*Response, error) {
    // Implementation
}
```

### 4. Error Handling

Custom error types in `pkg/errors/`:
```go
var (
    ErrProviderUnavailable = errors.New("provider unavailable")
    ErrRateLimitExceeded  = errors.New("rate limit exceeded")
    ErrCacheMiss          = errors.New("cache miss")
)
```

Wrap errors with context:
```go
return fmt.Errorf("failed to send request to %s: %w", provider, err)
```

---

## File Naming Conventions

- **Interfaces**: `interface.go` or named by domain (e.g., `provider.go`)
- **Implementations**: Named by what they do (e.g., `redis.go`, `openai.go`)
- **Tests**: `*_test.go` in the same package
- **Mocks**: `mock_*.go` or in `mocks/` subdirectory

---

## Testing Structure

```
internal/cache/
├── cache.go           # Interface
├── redis.go           # Implementation
├── redis_test.go      # Unit tests
└── mocks/
    └── cache.go       # Mock implementation

tests/
├── integration/
│   └── cache_test.go  # Integration tests with real Redis
└── e2e/
    └── flow_test.go   # Full end-to-end scenarios
```

**Test Types:**
1. **Unit tests**: Test individual functions, use mocks
2. **Integration tests**: Test with real dependencies (Redis, DB)
3. **E2E tests**: Test full request flow through gateway

---

## Configuration Files

### Development
`configs/config.dev.yaml`:
- Points to localhost services
- Verbose logging
- Short timeouts for fast feedback
- Test API keys

### Production
`configs/config.prod.yaml`:
- Points to production services
- Structured JSON logging
- Proper timeouts and retries
- Secrets from environment

---

## Docker Compose Services

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│  Redis   │  │ Postgres │  │ Gateway  │
│  :6379   │  │  :5432   │  │  :8080   │
└──────────┘  └──────────┘  └──────────┘
     │             │              │
     └─────────────┴──────────────┘
                   │
     ┌─────────────┴──────────────┐
     │                            │
┌──────────┐  ┌──────────┐  ┌──────────┐
│Prometheus│  │ Grafana  │  │  Jaeger  │
│  :9090   │  │  :3000   │  │  :16686  │
└──────────┘  └──────────┘  └──────────┘
```

---

## Build Artifacts

```
bin/
├── gateway              # Main binary
├── gateway-linux        # Linux binary
├── gateway-darwin       # macOS binary
└── gateway-windows.exe  # Windows binary

dist/
├── llm-gateway-v1.0.0-linux-amd64.tar.gz
└── llm-gateway-v1.0.0-darwin-arm64.tar.gz
```

---

## Makefile Targets

```makefile
.PHONY: build test run clean

build:
    go build -o bin/gateway cmd/gateway/main.go

test:
    go test -v ./...

integration-test:
    go test -v -tags=integration ./tests/integration/...

run:
    go run cmd/gateway/main.go

clean:
    rm -rf bin/ dist/

lint:
    golangci-lint run

docker-build:
    docker build -t llm-gateway:latest .

docker-up:
    docker-compose up -d

docker-down:
    docker-compose down

migrate-up:
    go run cmd/migrate/main.go up

migrate-down:
    go run cmd/migrate/main.go down
```

---

## Next Steps

1. Create the basic directory structure
2. Set up `go.mod` and initial dependencies
3. Implement `cmd/gateway/main.go`
4. Create configuration management
5. Build API router with health checks

Refer to GETTING_STARTED.md for detailed setup instructions!
