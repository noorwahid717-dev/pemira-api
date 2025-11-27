# Docker Setup

## Quick Start

### Using Docker Compose (Recommended)

1. **Start all services (PostgreSQL, Redis, API)**
   ```bash
   docker-compose up -d
   ```

2. **View logs**
   ```bash
   docker-compose logs -f api
   ```

3. **Stop all services**
   ```bash
   docker-compose down
   ```

### Build Docker Image Only

```bash
docker build -t pemira-api .
```

### Run Single Container

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://user:pass@host:5432/db \
  -e JWT_SECRET=your-secret \
  pemira-api
```

## Environment Variables

Configure these in `docker-compose.yml` or pass via `-e` flag:

- `APP_ENV` - Application environment (development/production)
- `HTTP_PORT` - HTTP server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `JWT_SECRET` - Secret key for JWT tokens
- `JWT_EXPIRATION` - Token expiration duration
- `LOG_LEVEL` - Logging level (info/debug/error)
- `CORS_ALLOWED_ORIGINS` - Allowed CORS origins

## Makefile Commands

```bash
make docker-up      # Start docker services
make docker-down    # Stop docker services
```

## Notes

- The Dockerfile uses multi-stage build for smaller image size
- Migrations folder is included in the image
- API runs as non-root user for security
- Health checks are configured for postgres and redis
