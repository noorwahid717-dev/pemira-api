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

### Public
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/` - API version info
- `GET /api/v1/elections/current` - Get current active election

### Authentication
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/auth/me` - Get current user (protected)
- `POST /api/v1/auth/logout` - Logout (protected)

### Elections (Protected)
- `GET /api/v1/elections/{id}/me/status` - Get voter status for election

### Voting (Student Only)
- `POST /api/v1/voting/online/cast` - Cast online vote
- `POST /api/v1/voting/tps/cast` - Cast TPS vote
- `GET /api/v1/voting/tps/status` - Get TPS voting status
- `GET /api/v1/voting/receipt` - Get voting receipt

### Admin - Election Management
- `GET /api/v1/admin/elections` - List elections
- `POST /api/v1/admin/elections` - Create election
- `GET /api/v1/admin/elections/{id}` - Get election detail
- `PUT /api/v1/admin/elections/{id}` - Update election
- `POST /api/v1/admin/elections/{id}/open-voting` - Open voting
- `POST /api/v1/admin/elections/{id}/close-voting` - Close voting

### Admin - DPT Management
- `POST /api/v1/admin/elections/{id}/voters/import` - Import voters
- `GET /api/v1/admin/elections/{id}/voters` - List voters
- `GET /api/v1/admin/elections/{id}/voters/export` - Export voters

### Admin - TPS Management
- `GET /api/v1/admin/tps` - List TPS
- `POST /api/v1/admin/tps` - Create TPS
- `GET /api/v1/admin/tps/{id}` - Get TPS detail
- `PUT /api/v1/admin/tps/{id}` - Update TPS
- `DELETE /api/v1/admin/tps/{id}` - Delete TPS
- `GET /api/v1/admin/tps/{id}/operators` - List operators
- `POST /api/v1/admin/tps/{id}/operators` - Create operator
- `DELETE /api/v1/admin/tps/{id}/operators/{userID}` - Remove operator
- `GET /api/v1/admin/elections/{id}/tps/monitor` - Monitor TPS

## Documentation

- [Admin Election API](./ADMIN_ELECTION_API.md) - Election management endpoints
- [Admin TPS API](./ADMIN_TPS_API.md) - TPS management endpoints
- [DPT API](./DPT_API_DOCUMENTATION.md) - DPT management endpoints
- [Voting API](./VOTING_API_IMPLEMENTATION.md) - Voting system implementation
- [Auth Implementation](./AUTH_IMPLEMENTATION.md) - Authentication & authorization

## License

MIT
