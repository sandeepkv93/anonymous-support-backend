# Anonymous Support Backend

[![CI](https://github.com/sandeepkv93/anonymous-support-backend/workflows/CI/badge.svg)](https://github.com/sandeepkv93/anonymous-support-backend/actions)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A production-ready backend system for an anonymous peer support mobile application where users can post real-time struggles with addictions/cravings and receive immediate community support.

**🎯 Production-Ready Features:**
- ✅ Connect-RPC API with full service implementation
- ✅ Comprehensive configuration validation
- ✅ Structured logging with request correlation IDs
- ✅ Advanced health checks (liveness, readiness, dependency status)
- ✅ Prometheus metrics endpoint
- ✅ Panic recovery with stack traces
- ✅ Database connection pooling (Postgres, MongoDB, Redis)
- ✅ JWT hardening (issuer, audience, not-before validation)
- ✅ RBAC authorization system
- ✅ CI/CD pipeline (GitHub Actions)
- ✅ Kubernetes deployment manifests
- ✅ Production-grade middleware stack

## 🚀 Features

- **Anonymous & Authenticated Users**: Support for both anonymous users and email-based accounts
- **Real-time Communication**: WebSocket-based real-time updates for posts and responses
- **Support Circles**: Create and join topic-specific support communities
- **Post Types**: SOS alerts, check-ins, victories, and questions
- **Quick Support**: One-tap support buttons for immediate encouragement
- **Strength Points**: Gamification to encourage helpful participation
- **Content Moderation**: Automatic flagging of harmful content
- **User Analytics**: Track streaks, cravings, and recovery progress
- **Rate Limiting**: Prevent spam and abuse
- **Secure**: JWT authentication, encrypted sensitive data, HTTPS ready

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     API Gateway Layer                   │
│  (Connect-RPC Handlers + WebSocket Manager)             │
└─────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────┐
│                    Service Layer                        │
│  (Business Logic - User, Post, Support, Moderation)     │
└─────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────┐
│                  Repository Layer                       │
│  (Data Access - Postgres, Redis, MongoDB)               │
└─────────────────────────────────────────────────────────┘
```

## 🛠️ Tech Stack

- **Language**: Go 1.24+
- **API Framework**: Connect-RPC (type-safe gRPC alternative)
- **Real-time**: WebSockets (gorilla/websocket)
- **Databases**:
  - PostgreSQL (users, circles, moderation)
  - MongoDB (posts, responses, analytics)
  - Redis (sessions, real-time, caching)
- **Authentication**: JWT tokens
- **Additional**: Zap (logging), Viper (config), golang-migrate (migrations)

## 📋 Prerequisites

- Go 1.24 or higher
- Docker & Docker Compose
- PostgreSQL 16
- MongoDB 7
- Redis 7

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/sandeepkv93/anonymous-support-backend.git
cd anonymous-support-backend
```

### 2. Run Setup Script

```bash
./scripts/setup.sh
```

This will:

- Create `.env` file from template
- Install required tools (buf, migrate)
- Start infrastructure services (Postgres, MongoDB, Redis)
- Run database migrations
- Initialize MongoDB collections

### 3. Start the Application

**Option A: Run locally**

```bash
make run
```

**Option B: Run with Docker**

```bash
make docker-up
```

The server will start on `http://localhost:8080`

## 📚 Available Commands

```bash
make help              # Show all available commands
make install-tools     # Install buf, migrate, and other tools
make proto             # Generate code from proto files
make migrate-up        # Run database migrations
make migrate-down      # Rollback migrations
make mongo-init        # Initialize MongoDB collections
make run               # Run the application
make build             # Build the application binary
make test              # Run tests
make docker-up         # Start all services with Docker
make docker-down       # Stop all services
make clean             # Clean build artifacts
```

## 🗄️ Database Migrations

Migrations are in `migrations/postgres/`:

- `001_create_users` - User accounts
- `002_create_circles` - Support circles
- `003_create_memberships` - Circle memberships
- `004_create_blocks` - User blocking
- `005_create_reports` - Content reporting

MongoDB collections are initialized via `migrations/mongodb/init.js`

## 🔌 API Endpoints

### REST Endpoints

- `GET /health` - Health check endpoint
- `WS /ws` - WebSocket connection (requires authentication)

### Connect-RPC Services (to be implemented with proto generation)

- **AuthService**: Registration, login, token refresh
- **UserService**: Profile management, streak tracking
- **PostService**: Create/read posts, feed generation
- **SupportService**: Responses, quick support, stats
- **CircleService**: Create/join circles, circle feeds
- **ModerationService**: Report content, moderation queue

## 🔐 Authentication

The API uses JWT bearer tokens:

```
Authorization: Bearer <access_token>
```

- Access tokens expire in 15 minutes
- Refresh tokens expire in 7 days
- Sessions are stored in Redis

## ⚙️ Configuration

Configuration is managed via environment variables. Copy `.env.example` to `.env` and customize:

```bash
# Server
SERVER_PORT=8080
SERVER_ENV=development

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=support_user
POSTGRES_PASSWORD=support_pass
POSTGRES_DB=support_db

# MongoDB
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=support_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-super-secret-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Encryption (must be exactly 32 bytes for AES-256)
ENCRYPTION_KEY=32-byte-encryption-key-for-aes-256

# Rate Limiting
RATE_LIMIT_POSTS_PER_HOUR=10
RATE_LIMIT_RESPONSES_PER_HOUR=100

# Moderation
ENABLE_AUTO_MODERATION=true
PROFANITY_FILTER_LEVEL=strict
```

## 📁 Project Structure

```
anonymous-support-backend/
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── domain/           # Domain models
│   ├── repository/       # Data access layer
│   │   ├── postgres/
│   │   ├── mongodb/
│   │   └── redis/
│   ├── service/          # Business logic
│   ├── handler/          # HTTP/WebSocket handlers
│   │   ├── connectrpc/
│   │   └── websocket/
│   ├── middleware/       # Auth, logging, CORS
│   ├── pkg/              # Shared utilities
│   │   ├── jwt/
│   │   ├── encryption/
│   │   ├── validator/
│   │   └── moderator/
│   └── proto/            # Protocol buffer definitions
├── migrations/           # Database migrations
│   ├── postgres/
│   └── mongodb/
├── scripts/              # Setup and utility scripts
├── docker-compose.yml    # Local development stack
├── Dockerfile
├── Makefile
└── README.md
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./internal/service/...

# Run with coverage
go test -cover ./...
```

## 🔒 Security Features

- **Password Hashing**: bcrypt with cost 12
- **Email Encryption**: AES-256 encryption for stored emails
- **JWT Validation**: Signature verification and expiry checks
- **Rate Limiting**: Per-user, per-endpoint rate limits
- **Content Moderation**: Automatic profanity and harmful content detection
- **User Blocking**: Users can block others
- **CORS**: Configurable cross-origin policies

## 📊 Monitoring & Observability

- **Structured Logging**: Zap with request correlation IDs
- **Health Checks**:
  - `/health` - Full dependency health check (Postgres, MongoDB, Redis)
  - `/health/ready` - Readiness probe for Kubernetes
  - `/health/live` - Liveness probe for Kubernetes
- **Metrics**: Prometheus metrics at `/metrics`
  - HTTP request metrics (total, duration, by endpoint)
  - Database query metrics (operations, duration)
  - WebSocket connection metrics
  - Cache hit/miss ratios
- **Error Tracking**: Panic recovery middleware with stack traces
- **Request Tracking**: Request ID propagation for distributed tracing

## 🚢 Deployment

### Docker Deployment

Build and run with Docker Compose:

```bash
docker-compose up -d
```

### Kubernetes Deployment

Production-ready Kubernetes manifests are available in the `k8s/` directory:

```bash
kubectl apply -f k8s/deployment.yaml
```

**Features:**
- HorizontalPodAutoscaler for auto-scaling (3-10 replicas)
- Resource limits and requests configured
- Liveness and readiness probes
- Secret management for sensitive configuration
- Service and ingress configuration

### Production Considerations

1. **Environment Variables**: Set secure values for:

   - `JWT_SECRET`
   - `ENCRYPTION_KEY`
   - Database passwords

2. **HTTPS**: Use a reverse proxy (nginx, Caddy) for TLS termination

3. **Database Backups**: Set up automated backups for PostgreSQL and MongoDB

4. **Monitoring**: Integrate with monitoring solutions (Datadog, New Relic, etc.)

5. **Scaling**:
   - Run multiple app instances behind a load balancer
   - Use Redis for cross-instance WebSocket message routing
   - Consider read replicas for databases

## 🐛 Troubleshooting

### Connection Issues

```bash
# Check if services are running
docker-compose ps

# View logs
docker-compose logs -f app
docker-compose logs -f postgres
docker-compose logs -f mongodb
docker-compose logs -f redis

# Restart services
docker-compose restart
```

### Migration Issues

```bash
# Check migration status
migrate -path migrations/postgres -database "<connection-string>" version

# Force migration version
migrate -path migrations/postgres -database "<connection-string>" force <version>
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License.

## 👥 Support

For questions and support:

- Open an issue on [GitHub](https://github.com/sandeepkv93/anonymous-support-backend/issues)
- Check the [Discussions](https://github.com/sandeepkv93/anonymous-support-backend/discussions)

## 🔗 Links

- **Repository**: [github.com/sandeepkv93/anonymous-support-backend](https://github.com/sandeepkv93/anonymous-support-backend)
- **Issues**: [Report bugs or request features](https://github.com/sandeepkv93/anonymous-support-backend/issues)
- **CI/CD**: [GitHub Actions](https://github.com/sandeepkv93/anonymous-support-backend/actions)

---

Built with ❤️ for recovery and peer support | [View on GitHub](https://github.com/sandeepkv93/anonymous-support-backend)
