# Getting Started with LLM Gateway

## Immediate Next Steps (Your First Day)

### 1. Choose Your Language
**Recommendation: Go**

**Why Go?**
- ✅ Excellent performance (closer to Rust than Python/Node)
- ✅ Built-in concurrency (goroutines perfect for high-throughput proxy)
- ✅ Strong HTTP/networking libraries
- ✅ Great Redis, database, and cloud SDK support
- ✅ Easy deployment (single binary)
- ✅ Large LLM ecosystem (LiteLLM inspiration, many examples)

**If you prefer Rust:**
- ✅ Maximum performance
- ✅ Memory safety
- ❌ Steeper learning curve
- ❌ Slower development initially
- ❌ Smaller ecosystem for LLM tooling

**Decision:** Choose Go unless you're already Rust-experienced.

### 2. Set Up Development Environment

```bash
# Install Go (if not installed)
# Visit https://go.dev/dl/ or use your package manager

# Verify installation
go version  # Should be 1.21+

# Initialize the project
cd llm-gate
go mod init github.com/yourusername/llm-gate

# Create initial directory structure
mkdir -p cmd/gateway
mkdir -p internal/{api,core,providers,cache,routing,telemetry}
mkdir -p pkg
mkdir -p configs
mkdir -p scripts
mkdir -p tests
mkdir -p docs

# Install recommended tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/cosmtrek/air@latest  # Hot reload for development
```

### 3. Set Up Docker Environment

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  redis:
    image: redis/redis-stack:latest
    ports:
      - "6379:6379"
      - "8001:8001"  # RedisInsight
    volumes:
      - redis-data:/data
    environment:
      - REDIS_ARGS=--save 60 1000

  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=llmgate
      - POSTGRES_PASSWORD=dev_password
      - POSTGRES_DB=llmgate
    volumes:
      - postgres-data:/var/lib/postgresql/data

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana-data:/var/lib/grafana
      - ./configs/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./configs/grafana/datasources:/etc/grafana/provisioning/datasources

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "5775:5775/udp"
      - "6831:6831/udp"
      - "6832:6832/udp"
      - "5778:5778"
      - "16686:16686"  # Jaeger UI
      - "14268:14268"
      - "14250:14250"
      - "9411:9411"
    environment:
      - COLLECTOR_OTLP_ENABLED=true

volumes:
  redis-data:
  postgres-data:
  prometheus-data:
  grafana-data:
```

Start services:
```bash
docker-compose up -d
```

### 4. Create Initial Configuration

Create `configs/config.yaml`:

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

cache:
  enabled: true
  redis:
    host: "localhost:6379"
    db: 0
    pool_size: 10
  semantic:
    enabled: true
    similarity_threshold: 0.95
    embedding_model: "text-embedding-3-small"

providers:
  - name: openai
    enabled: true
    api_key_env: "OPENAI_API_KEY"
    base_url: "https://api.openai.com/v1"
    timeout: 30s
    max_retries: 3

  - name: anthropic
    enabled: false
    api_key_env: "ANTHROPIC_API_KEY"
    base_url: "https://api.anthropic.com/v1"
    timeout: 30s
    max_retries: 3

routing:
  strategy: "round_robin"  # round_robin, least_latency, cost_optimized
  enable_fallback: true
  fallback_chain:
    - openai
    - anthropic

rate_limiting:
  enabled: true
  default_rate: 100  # requests per minute
  burst: 20

telemetry:
  enabled: true
  service_name: "llm-gateway"
  jaeger:
    endpoint: "http://localhost:14268/api/traces"
  prometheus:
    enabled: true
    path: "/metrics"

logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, console
```

Create `.env`:

```bash
# Provider API Keys
OPENAI_API_KEY=your_openai_key_here
ANTHROPIC_API_KEY=your_anthropic_key_here
AZURE_OPENAI_KEY=your_azure_key_here
AWS_ACCESS_KEY_ID=your_aws_key_here
AWS_SECRET_ACCESS_KEY=your_aws_secret_here

# Database
DATABASE_URL=postgresql://llmgate:dev_password@localhost:5432/llmgate

# Redis
REDIS_URL=redis://localhost:6379/0

# Server
PORT=8080
ENVIRONMENT=development
```

### 5. Create Initial Project Files

**Create `cmd/gateway/main.go`:**

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/yourusername/llm-gate/internal/api"
    "github.com/yourusername/llm-gate/internal/config"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize router
    router := api.NewRouter(cfg)

    // Create HTTP server
    srv := &http.Server{
        Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
        Handler:      router,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
    }

    // Start server in goroutine
    go func() {
        log.Printf("Starting LLM Gateway on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited")
}
```

### 6. Install Initial Dependencies

```bash
go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
go get github.com/redis/go-redis/v9
go get github.com/spf13/viper
go get github.com/rs/zerolog
go get github.com/joho/godotenv
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/jaeger
go get github.com/prometheus/client_golang/prometheus
```

### 7. Create First API Endpoint

**Create `internal/api/router.go`:**

```go
package api

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
)

func NewRouter() *chi.Mux {
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: false,
        MaxAge:           300,
    }))

    // Health check
    r.Get("/health", healthCheck)
    r.Get("/ready", readyCheck)

    // API routes
    r.Route("/v1", func(r chi.Router) {
        r.Post("/chat/completions", handleChatCompletion)
    })

    return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
    })
}

func readyCheck(w http.ResponseWriter, r *http.Request) {
    // TODO: Check Redis, DB connections
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ready",
    })
}

func handleChatCompletion(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    w.WriteHeader(http.StatusNotImplemented)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "Not implemented yet",
    })
}
```

### 8. Run Your First Version

```bash
# Load environment variables
export $(cat .env | xargs)

# Run the server
go run cmd/gateway/main.go

# Test it
curl http://localhost:8080/health
# Should return: {"status":"ok"}
```

### 9. Set Up Version Control Best Practices

Create `.gitignore`:

```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test files
*.test
*.out
coverage.txt

# Go workspace file
go.work

# Environment variables
.env
.env.local
*.key
*.pem

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Data
data/
*.db
*.sqlite
```

Create `.golangci.yml`:

```yaml
linters:
  enable:
    - gofmt
    - golint
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - ineffassign
    - unused

linters-settings:
  govet:
    check-shadowing: true

run:
  timeout: 5m
```

### 10. Create Your First GitHub Issues

Create issues for Phase 1 tasks:

```bash
# Issue 1: Set up project structure
# Issue 2: Implement configuration management
# Issue 3: Add structured logging
# Issue 4: Create health check endpoints
# Issue 5: Set up Docker Compose environment
```

---

## Your Week 1 Goals

By end of Week 1, you should have:

- ✅ Go project initialized with proper structure
- ✅ Docker Compose running (Redis, PostgreSQL, Prometheus, Grafana, Jaeger)
- ✅ Basic HTTP server with health checks
- ✅ Configuration management working
- ✅ Structured logging in place
- ✅ First successful request to `/health` endpoint

---

## Development Workflow

### Daily Routine

```bash
# Start infrastructure
docker-compose up -d

# Check logs
docker-compose logs -f redis

# Run with hot reload (optional)
air

# Or run directly
go run cmd/gateway/main.go

# Run tests
go test ./...

# Check code quality
golangci-lint run
```

### Testing Your Changes

```bash
# Unit tests
go test ./internal/...

# Integration tests (later)
go test -tags=integration ./tests/...

# Benchmarks
go test -bench=. ./...
```

---

## Common Commands Reference

```bash
# Add new dependency
go get github.com/some/package

# Update dependencies
go mod tidy

# Verify dependencies
go mod verify

# View dependency graph
go mod graph

# Build binary
go build -o bin/gateway cmd/gateway/main.go

# Run binary
./bin/gateway

# Cross-compile (for deployment)
GOOS=linux GOARCH=amd64 go build -o bin/gateway-linux cmd/gateway/main.go
```

---

## Recommended VS Code Extensions

- Go (by Go Team at Google)
- Go Test Explorer
- REST Client (for API testing)
- Docker
- YAML
- GitLens

---

## Learning Resources for Week 1

### Go Basics (if needed)
- [Tour of Go](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

### HTTP Servers in Go
- [Building Web Apps with Go](https://www.usegolang.com/)
- Chi Router documentation
- Go net/http package docs

### Architecture Patterns
- Clean Architecture in Go
- [Go Project Layout](https://github.com/golang-standards/project-layout)

---

## Troubleshooting

### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Redis Connection Failed
```bash
# Check Redis is running
docker-compose ps redis

# Test Redis connection
redis-cli ping
```

### Go Module Issues
```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
```

---

## Next Steps After Week 1

Once you have the foundation working:

1. **Read Phase 2 in ROADMAP.md**
2. **Implement OpenAI provider** (your first LLM integration)
3. **Create provider abstraction** (foundation for multi-provider)
4. **Add request/response models**
5. **Test with real OpenAI API calls**

---

## Questions to Ask Yourself

Before moving to Phase 2:

- [ ] Can I start/stop the server cleanly?
- [ ] Are my health checks responding?
- [ ] Is configuration loading from files and env vars?
- [ ] Are logs structured and readable?
- [ ] Can I run tests?
- [ ] Is Docker Compose working with all services?

---

## Get Help

If stuck:
- Check Go documentation
- Review similar projects (LiteLLM, Kong Gateway)
- Read the ROADMAP.md for context
- Search GitHub issues in related projects

---

## You're Ready! 🚀

Start with Phase 1.1 in the ROADMAP and build incrementally. Remember:

- **Start small:** Get something working first
- **Test frequently:** Don't build too much without testing
- **Commit often:** Small, atomic commits
- **Document as you go:** Future you will thank present you

Good luck building your LLM gateway!
