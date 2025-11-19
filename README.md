# PEMIRA API

Electronic voting system API for university student elections (Pemilihan Raya Mahasiswa).

## Tech Stack

- **Language**: Go 1.22+
- **HTTP Router**: chi/v5
- **Database**: PostgreSQL + pgx/v5
- **Query Layer**: sqlc
- **Migrations**: goose
- **Auth**: JWT (golang-jwt/jwt/v5) + bcrypt
- **WebSocket**: nhooyr.io/websocket
- **Config**: envconfig
- **Logging**: log/slog
- **Metrics**: Prometheus
- **Validation**: go-playground/validator/v10
- **Cache**: Redis (optional)

## Project Structure

```
.
├── cmd/
│   ├── api/              # Main API server
│   └── worker/           # Background worker (optional)
├── internal/
│   ├── auth/             # Authentication, JWT, user management
│   ├── election/         # Election & phase management
│   ├── voter/            # DPT (Daftar Pemilih Tetap) management
│   ├── candidate/        # Candidate profiles & campaigns
│   ├── tps/              # TPS management, QR codes, checkins
│   ├── voting/           # Voting engine (online & TPS)
│   ├── monitoring/       # Live count, statistics (TODO)
│   ├── announcement/     # Announcements (TODO)
│   ├── audit/            # Audit logs (TODO)
│   ├── fileimport/       # CSV/XLSX import utilities (TODO)
│   ├── config/           # Configuration loader
│   ├── http/             # HTTP server setup
│   │   ├── middleware/   # Auth, RBAC, logging
│   │   └── response/     # JSON response helpers
│   ├── ws/               # WebSocket hub & handlers
│   └── shared/           # Common utilities
│       ├── constants/    # Enums, roles, statuses
│       ├── ctxkeys/      # Context keys
│       ├── errors.go     # Domain errors
│       └── pagination.go # Pagination helpers
├── pkg/
│   └── database/         # Database connection utilities
├── migrations/           # SQL migrations (goose)
├── docs/                 # Documentation (TODO: ERD, OpenAPI)
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## Module Structure

Each module follows this pattern:
```
internal/<module>/
├── entity.go         # Domain entities
├── repository.go     # Data access interface
├── service.go        # Business logic
├── http_handler.go   # HTTP handlers
├── dto.go            # Request/Response DTOs
└── ws_handler.go     # WebSocket handlers (if needed)
```

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make (optional)

### Installation

1. Clone the repository:
```bash
git clone git@github.com:noah-isme/pemira-api.git
cd pemira-api
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Start dependencies:
```bash
make docker-up
# or
docker-compose up -d
```

4. Run migrations:
```bash
make migrate-up
# or
goose -dir migrations postgres "$DATABASE_URL" up
```

5. Run the application:
```bash
make dev
# or
go run cmd/api/main.go
```

## Development

### Running Tests
```bash
make test
```

### Running Linter
```bash
make lint
```

### Creating Migrations
```bash
make migrate-create name=create_users_table
```

### Generate sqlc Code
```bash
make sqlc-generate
```

## API Endpoints

- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/` - API version info

## License

MIT
