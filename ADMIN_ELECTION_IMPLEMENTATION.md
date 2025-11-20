# Admin Election Management - Implementation Summary

## Overview

Implementasi sistem manajemen pemilu untuk admin/panitia yang mencakup:
- ✅ CRUD pemilu (Create, Read, Update, List)
- ✅ Kontrol status voting (Open/Close)
- ✅ Toggle mode voting (Online/TPS/Hybrid)
- ✅ Integration dengan voting system yang sudah ada

## Architecture

### Layer Structure

```
HTTP Handler (admin_http_handler.go)
       ↓
Service Layer (admin_service.go)
       ↓
Repository Interface (admin_repository.go)
       ↓
PostgreSQL Implementation (admin_repository_pgx.go)
       ↓
Database (elections table)
```

## Files Created

### 1. `internal/election/admin_model.go`
- **AdminElectionDTO**: Response model untuk admin
- **AdminElectionListFilter**: Filter untuk list elections
- **AdminElectionCreateRequest**: Request body untuk create
- **AdminElectionUpdateRequest**: Request body untuk update (partial)
- **Pagination**: Metadata pagination

### 2. `internal/election/admin_repository.go`
Interface repository dengan method:
- `ListElections()`: List dengan filter, search, pagination
- `GetElectionByID()`: Get detail by ID
- `CreateElection()`: Create election baru
- `UpdateElection()`: Update election (partial)
- `SetVotingStatus()`: Update status + timestamp voting

### 3. `internal/election/admin_repository_pgx.go`
PostgreSQL implementation dengan fitur:
- Dynamic query building untuk filter & search
- Partial update (hanya field non-nil yang di-update)
- COALESCE untuk optional timestamp update
- Proper error handling (ErrElectionNotFound)

### 4. `internal/election/admin_service.go`
Business logic layer dengan:
- **List()**: Pagination logic
- **Create()**: Create dengan status DRAFT
- **Get()**: Get by ID
- **Update()**: Partial update
- **OpenVoting()**: Business rules untuk open voting
- **CloseVoting()**: Business rules untuk close voting

#### Business Rules - OpenVoting
```go
✅ Allowed:
- Status: DRAFT → VOTING_OPEN
- Status: CAMPAIGN → VOTING_OPEN
- Status: REGISTRATION → VOTING_OPEN

❌ Not Allowed:
- Already VOTING_OPEN → Error: ErrElectionAlreadyOpen
- Already ARCHIVED → Error: ErrInvalidStatusChange

Auto-set:
- voting_start_at = NOW() (if still null)
```

#### Business Rules - CloseVoting
```go
✅ Allowed:
- Status: VOTING_OPEN → VOTING_CLOSED

❌ Not Allowed:
- Not VOTING_OPEN → Error: ErrElectionNotInOpenState

Auto-set:
- voting_end_at = NOW()
```

### 5. `internal/election/admin_http_handler.go`
HTTP handlers dengan:
- Proper validation & error handling
- Response helper usage (response package)
- URL parameter parsing
- JSON request/response

## Integration Points

### 1. Main Application (`cmd/api/main.go`)

```go
// Repository initialization
electionAdminRepo := election.NewPgAdminRepository(pool)

// Service initialization
electionAdminService := election.NewAdminService(electionAdminRepo)

// Handler initialization
electionAdminHandler := election.NewAdminHandler(electionAdminService)

// Routing
r.Group(func(r chi.Router) {
    r.Use(httpMiddleware.AuthAdminOnly(jwtManager))
    
    r.Route("/admin/elections", func(r chi.Router) {
        r.Get("/", electionAdminHandler.List)
        r.Post("/", electionAdminHandler.Create)
        r.Get("/{electionID}", electionAdminHandler.Get)
        r.Put("/{electionID}", electionAdminHandler.Update)
        r.Post("/{electionID}/open-voting", electionAdminHandler.OpenVoting)
        r.Post("/{electionID}/close-voting", electionAdminHandler.CloseVoting)
    })
})
```

### 2. Repository Helper (`internal/election/repository_pgx.go`)

Added `NewRepository()` function untuk compatibility:
```go
func NewRepository(db *pgxpool.Pool) Repository {
    return NewPgRepository(db)
}
```

### 3. Public Election Handler
Already implemented in `internal/election/http_handler.go`:
- `GET /elections/current` - Get active election (VOTING_OPEN)
- `GET /elections/{id}/me/status` - Get voter status

## Database Schema

Uses existing `elections` table:
```sql
CREATE TABLE elections (
    id              BIGSERIAL PRIMARY KEY,
    year            INT NOT NULL,
    name            TEXT NOT NULL,
    code            TEXT NOT NULL UNIQUE,  -- slug
    status          TEXT NOT NULL,         -- ElectionStatus enum
    voting_start_at TIMESTAMPTZ,
    voting_end_at   TIMESTAMPTZ,
    online_enabled  BOOLEAN NOT NULL DEFAULT true,
    tps_enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## API Endpoints

All endpoints require `Authorization: Bearer <token>` with **ADMIN** role.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/elections` | List elections (filter, search, pagination) |
| POST | `/admin/elections` | Create new election |
| GET | `/admin/elections/{id}` | Get election detail |
| PUT | `/admin/elections/{id}` | Update election (partial) |
| POST | `/admin/elections/{id}/open-voting` | Open voting |
| POST | `/admin/elections/{id}/close-voting` | Close voting |

## Features

### 1. Election CRUD
- Create election dengan mode voting settings
- Update election info (year, name, slug)
- Get detail election
- List dengan filter (year, status, search) dan pagination

### 2. Voting Status Control
- **Open Voting**: Set status ke VOTING_OPEN + timestamp
- **Close Voting**: Set status ke VOTING_CLOSED + timestamp
- Business rules validation

### 3. Toggle Voting Mode
- `online_enabled`: true/false
- `tps_enabled`: true/false
- Dapat diubah kapan saja via Update endpoint
- Mendukung 3 mode:
  - Online Only: `online_enabled=true, tps_enabled=false`
  - TPS Only: `online_enabled=false, tps_enabled=true`
  - Hybrid: `online_enabled=true, tps_enabled=true`

### 4. Search & Filter
- Filter by year
- Filter by status
- Search by name or slug
- Pagination support

## Integration with Existing Features

### Voting System
- `/voting/online/cast` checks `online_enabled`
- `/voting/tps/cast` checks `tps_enabled`
- Validation based on election settings

### Voter Status
- `/elections/{id}/me/status` returns `online_allowed` & `tps_allowed`
- Based on `online_enabled` & `tps_enabled` from election

### Current Election
- `/elections/current` returns active election (VOTING_OPEN)
- Includes `online_enabled` & `tps_enabled` fields

## Testing Scenarios

### Scenario 1: Create and Open Election
```bash
# 1. Create election
POST /admin/elections
{
  "year": 2024,
  "name": "Pemilu Raya 2024",
  "slug": "pemira-2024",
  "online_enabled": true,
  "tps_enabled": true
}
→ Status: DRAFT

# 2. Open voting
POST /admin/elections/1/open-voting
→ Status: VOTING_OPEN
→ voting_start_at: NOW()

# 3. Check current election (public)
GET /elections/current
→ Returns the election
```

### Scenario 2: Toggle Voting Mode
```bash
# During voting, disable online mode
PUT /admin/elections/1
{
  "online_enabled": false
}
→ Only TPS voting allowed now

# Re-enable online
PUT /admin/elections/1
{
  "online_enabled": true
}
→ Hybrid mode active
```

### Scenario 3: Close Voting
```bash
POST /admin/elections/1/close-voting
→ Status: VOTING_CLOSED
→ voting_end_at: NOW()

# Try to vote (should fail)
POST /voting/online/cast
→ Error: Election not open
```

## Error Handling

All endpoints return proper HTTP status codes:
- **200 OK**: Success
- **201 Created**: Resource created
- **400 Bad Request**: Invalid request or business rule violation
- **404 Not Found**: Resource not found
- **422 Unprocessable Entity**: Validation error
- **500 Internal Server Error**: Server error

Error response format:
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

## Security

- All endpoints protected by JWT authentication
- Role-based access control (ADMIN only)
- Middleware: `httpMiddleware.AuthAdminOnly(jwtManager)`

## Documentation

- [API Documentation](./ADMIN_ELECTION_API.md)
- [README](./README.md)

## Commits

1. `feat: add admin election management (CRUD, open/close voting)`
   - admin_model.go
   - admin_repository.go
   - admin_repository_pgx.go
   - admin_service.go
   - admin_http_handler.go

2. `feat: integrate admin election routes and services in main app`
   - repository_pgx.go (added NewRepository helper)
   - cmd/api/main.go (wiring)

3. `docs: add admin election management API documentation`
   - ADMIN_ELECTION_API.md

4. `docs: update README with admin election endpoints`
   - README.md

## Next Steps

Possible enhancements:
- [ ] Schedule voting (future start/end time)
- [ ] Archive election
- [ ] Clone election from previous year
- [ ] Validation rule configurasi
- [ ] Election statistics dashboard
- [ ] Audit log untuk election changes
- [ ] Webhook/notification pada status change

## Dependencies

No new dependencies added. Uses existing:
- github.com/go-chi/chi/v5
- github.com/jackc/pgx/v5
- pemira-api/internal/http/response
- pemira-api/internal/http/middleware
