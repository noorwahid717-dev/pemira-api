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
│   ├── api/          # Main API server
│   └── worker/       # Background worker
├── internal/
│   ├── auth/         # Authentication & JWT
│   ├── config/       # Configuration
│   ├── domain/       # Business logic
│   ├── http/         # HTTP handlers
│   │   ├── middleware/
│   │   └── response/
│   └── ws/           # WebSocket handlers
├── pkg/
│   └── database/     # Database utilities
├── migrations/       # SQL migrations
├── docker-compose.yml
├── Dockerfile
└── Makefile
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
