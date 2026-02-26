# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**New API** is an AI Gateway & API Management Platform that provides a unified interface to 30+ LLM and AI service providers. Based on the [One API](https://github.com/songquanpeng/one-api) project with significant enhancements, it serves as a multi-tenant SaaS platform with user management, billing, quota tracking, and intelligent request routing.

**Tech Stack:**
- **Backend**: Go 1.25.1 (Gin framework, GORM ORM)
- **Frontend**: React 18 (Vite, Semi UI, TailwindCSS, Bun package manager)
- **Databases**: PostgreSQL/MySQL/SQLite with Redis caching
- **Deployment**: Docker + Docker Compose
- **Internationalization**: Chinese, English, French, Japanese

## Architecture

### Layered Architecture Pattern

```
Frontend (React SPA)
    ↓
Router Layer (gin routes)
    ↓
Controller Layer (HTTP handlers)
    ↓
Service Layer (business logic)
    ↓
Relay Layer (30+ AI provider integrations)
    ↓
Model Layer (GORM + Cache)
    ↓
Database & Redis
```

### Key Directory Structure

- **router/** - 5 router categories: API (admin), Relay (AI models), Dashboard, Video, Web (static files)
- **controller/** - HTTP request handlers and business logic
- **service/** - Channel selection, quota management, token counting, format conversion
- **relay/** - Provider-specific handlers for OpenAI, Claude, Gemini, Bedrock, Cohere, etc.
  - Implements protocol conversion between different AI API formats
  - Handles WebSocket and SSE streaming
  - Supports image, audio, video, embedding, and rerank endpoints
- **model/** - GORM entities: User, Token, Channel, Ability, Quota, Log, Task, Redemption
- **middleware/** - Authentication, rate limiting, request logging
- **constant/** - Dependency-free constants layer (see Critical Conventions below)
- **common/** - Shared utilities: database, caching, crypto, email, HTTP client
- **dto/** - Data transfer objects
- **types/** - Type definitions
- **setting/** - Configuration management
- **web/** - React frontend with Vite build system

### Entry Point

[main.go](main.go) initializes the application:
1. `InitResources()` - Load env vars, initialize database, Redis, logging
2. Spawn background goroutines:
   - Channel caching synchronization
   - Options/settings hot-reload
   - Quota data updates
   - Channel auto-testing
   - Task bulk updates (Midjourney, Suno, etc.)
   - Batch update system
3. Setup Gin server with middleware
4. Configure routers
5. Start on PORT (default 3000)

## Development Commands

### Build & Run

```bash
# Full build and start
make all

# Build frontend only (uses Bun)
make build-frontend

# Start backend only
make start-backend

# Compile backend binary
go build -o new-api

# Full stack with Docker
docker-compose up -d
```

### Frontend Development

```bash
cd web

# Install dependencies
bun install

# Start Vite dev server (with proxy to backend)
bun run dev

# Production build
bun run build

# Code formatting
bun run lint:fix

# Extract translation strings
bun run i18n:extract
```

### Testing

```bash
# Run all Go tests
go test ./...

# Test specific package with verbose output
go test -v ./controller

# API endpoint performance test
./bin/time_test.sh <domain> <api_key> <count> [model]
```

### Docker

```bash
# Build container
docker build -t new-api:latest .

# Start services (PostgreSQL + Redis + New-API)
docker-compose up -d

# Default ports:
# - 3000: Main application
# - 5432: PostgreSQL
# - 6379: Redis
```

## Critical Code Conventions

### constant/ Package Rule (CRITICAL)

**From [constant/README.md](constant/README.md):**

The `constant/` package is **dependency-free** and MUST NOT import any other project packages (only Go standard library). This is a critical architectural constraint.

- Used for global constants only (no business logic)
- Violating this creates circular dependencies and breaks the build
- Files include: `azure.go`, `cache_key.go`, `channel_setting.go`, `context_key.go`, `env.go`, `finish_reason.go`, `midjourney.go`, `task.go`, `user_setting.go`

When adding code to `constant/`, ensure zero cross-package imports from this project.

### Environment Configuration

Configuration is environment-driven via `.env` file. See [.env.example](.env.example) for all options.

**Key Variables:**
- `SQL_DSN` - Database connection string (PostgreSQL/MySQL/SQLite)
- `REDIS_CONN_STRING` - Redis connection (e.g., `redis://redis:6379`)
- `SESSION_SECRET` - **Required for multi-machine deployment** (must be identical across instances)
- `CRYPTO_SECRET` - **Required when using shared Redis**
- `STREAMING_TIMEOUT` - Default 300s, increase if experiencing empty completions
- `BATCH_UPDATE_ENABLED` - Enable batch updates for performance
- `SYNC_FREQUENCY` - Channel cache sync interval (seconds)
- `CHANNEL_UPDATE_FREQUENCY` - How often to update channel status

### Multi-Machine Deployment

For distributed deployments:
1. Set `SESSION_SECRET` to the same value across all instances
2. Set `CRYPTO_SECRET` when using shared Redis for encryption
3. Configure `SYNC_FREQUENCY` for cache synchronization
4. Use PostgreSQL or MySQL (not SQLite) for shared database

## Database Support

- **PostgreSQL** (default): `postgresql://user:pass@host:5432/new-api`
- **MySQL** 5.7.8+: `root:password@tcp(host:3306)/new-api`
- **SQLite**: Set `SQLITE_PATH` environment variable (local only)
- Migrations handled automatically by GORM on startup

Default configuration in [docker-compose.yml](docker-compose.yml) uses PostgreSQL. To switch to MySQL, uncomment the mysql service and update `SQL_DSN`.

## Key Features

### Multi-Provider AI Gateway
- 30+ AI provider integrations: OpenAI, Claude (Anthropic), Gemini (Google), AWS Bedrock, Azure OpenAI, Cohere, Mistral, DeepSeek, Qwen, Moonshot, Perplexity, and more
- Protocol conversion: OpenAI ↔ Claude Messages ↔ Gemini Chat
- Image generation: Midjourney-Proxy, Stable Diffusion
- Audio: Suno API
- Video: Multiple providers
- Embeddings and reranking support

### User & Access Management
- Multi-tenant user system with groups
- API token generation and management
- Role-based access control (RBAC)
- OAuth integration: Discord, LinuxDO, Telegram, OIDC
- WebAuthn/Passkey support

### Billing & Quota
- Token-based usage tracking
- Model pricing configuration (per-token, per-call, group ratios)
- Quota pre-consumption with free model handling
- Payment integration: Stripe, 易支付
- Redemption/promotional codes
- Usage analytics dashboard

### Intelligent Routing
- Weighted random channel selection
- Automatic failover on errors
- User-level rate limiting
- Channel health monitoring and auto-testing
- Load balancing across providers

### Asynchronous Processing
- Task system for long-running operations
- Midjourney image generation tracking
- Suno audio generation
- Video generation workflows
- Batch update system for performance

## Important Files

- [main.go](main.go) - Application entry point with initialization flow
- [docker-compose.yml](docker-compose.yml) - Production deployment configuration
- [.env.example](.env.example) - Complete list of environment variables
- [makefile](makefile) - Build orchestration
- [constant/README.md](constant/README.md) - Critical architectural rule
- `web/package.json` - Frontend dependencies and NPM scripts
- `go.mod` - Backend dependencies (Go 1.25.1)

## Internationalization

4 languages supported: Chinese (default), English, French, Japanese

Translation files: `web/src/i18n/`

Extract new translation strings:
```bash
cd web && bun run i18n:extract
```

## Licensing

Dual licensing model:
- **AGPLv3** (open source) - Requires preserving brand/logo/copyright and source disclosure for network services
- **Commercial License** - Required to remove branding

See [LICENSE](LICENSE) for full details. When using AGPLv3, you must disclose source code if providing this as a network service (SaaS).
