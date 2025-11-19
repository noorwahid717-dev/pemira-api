# Admin Candidate Handler Integration Guide

## Endpoints

### List Candidates (Admin)
```
GET /admin/elections/{electionID}/candidates
```

**Query Parameters:**
- `search` (optional) - Search by name
- `status` (optional) - Filter by status: DRAFT, PUBLISHED, HIDDEN, ARCHIVED
- `page` (optional, default: 1)
- `limit` (optional, default: 20)

**Response:**
```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "election_id": 1,
        "number": 1,
        "name": "Pasangan Calon A",
        "status": "DRAFT",
        ...
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total_items": 5,
      "total_pages": 1
    }
  }
}
```

---

### Create Candidate
```
POST /admin/elections/{electionID}/candidates
```

**Request Body:**
```json
{
  "number": 1,
  "name": "Pasangan Calon A",
  "photo_url": "https://...",
  "short_bio": "...",
  "long_bio": "...",
  "tagline": "...",
  "faculty_name": "Fakultas Teknik",
  "study_program_name": "Informatika",
  "cohort_year": 2021,
  "vision": "...",
  "missions": ["...", "..."],
  "main_programs": [
    {
      "title": "Program 1",
      "description": "...",
      "category": "Pendidikan"
    }
  ],
  "media": {
    "video_url": "https://youtube.com/...",
    "gallery_photos": ["https://..."],
    "document_manifesto_url": "https://..."
  },
  "social_links": [
    {
      "platform": "instagram",
      "url": "https://instagram.com/..."
    }
  ],
  "status": "DRAFT"
}
```

**Response:** `201 Created` with candidate detail

---

### Get Candidate Detail
```
GET /admin/candidates/{candidateID}?election_id={electionID}
```

**Response:** Candidate detail (all fields)

---

### Update Candidate
```
PUT /admin/candidates/{candidateID}?election_id={electionID}
```

**Request Body:** Partial update (all fields optional)
```json
{
  "name": "Pasangan Calon A (Updated)",
  "status": "PUBLISHED"
}
```

**Response:** Updated candidate detail

---

### Delete Candidate
```
DELETE /admin/candidates/{candidateID}?election_id={electionID}
```

**Response:** `204 No Content`

---

### Publish Candidate
```
POST /admin/candidates/{candidateID}/publish?election_id={electionID}
```

Changes status to `PUBLISHED`.

**Response:** Updated candidate detail

---

### Unpublish Candidate
```
POST /admin/candidates/{candidateID}/unpublish?election_id={electionID}
```

Changes status to `HIDDEN`.

**Response:** Updated candidate detail

---

## Router Integration Example

```go
package api

import (
    "github.com/go-chi/chi/v5"
    "pemira-api/internal/candidate"
    "pemira-api/internal/http/middleware"
)

type Dependencies struct {
    AdminCandidateService candidate.AdminCandidateService
    // ... other services
}

func NewRouter(dep Dependencies) http.Handler {
    r := chi.NewRouter()

    // Global middleware
    r.Use(middleware.Logger)
    r.Use(middleware.CORS)

    // Admin routes
    r.Route("/admin", func(ad chi.Router) {
        ad.Use(middleware.AuthAdminOnly) // JWT + role = ADMIN

        adminCandHandler := candidate.NewAdminHandler(dep.AdminCandidateService)

        // List & Create
        ad.Route("/elections/{electionID}", func(er chi.Router) {
            er.Get("/candidates", adminCandHandler.List)
            er.Post("/candidates", adminCandHandler.Create)
        })

        // Detail, Update, Delete, Publish, Unpublish
        ad.Route("/candidates/{candidateID}", func(cr chi.Router) {
            cr.Get("/", adminCandHandler.Detail)
            cr.Put("/", adminCandHandler.Update)
            cr.Delete("/", adminCandHandler.Delete)
            cr.Post("/publish", adminCandHandler.Publish)
            cr.Post("/unpublish", adminCandHandler.Unpublish)
        })
    })

    return r
}
```

## Error Responses

### 400 Bad Request
```json
{
  "code": "BAD_REQUEST",
  "message": "electionID tidak valid."
}
```

### 404 Not Found
```json
{
  "code": "NOT_FOUND",
  "message": "Kandidat tidak ditemukan."
}
```

### 409 Conflict
```json
{
  "code": "CANDIDATE_NUMBER_TAKEN",
  "message": "Nomor kandidat sudah digunakan di pemilu ini."
}
```

### 422 Unprocessable Entity
```json
{
  "code": "VALIDATION_ERROR",
  "message": "number dan name wajib diisi."
}
```

### 500 Internal Server Error
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Terjadi kesalahan pada sistem."
}
```

## Implementation Checklist

- [x] Admin HTTP Handler (`admin_http_handler.go`)
- [ ] Admin Service Implementation
- [ ] Repository methods for CRUD
- [ ] Integration tests
- [ ] Router mounting
- [ ] Authentication middleware
- [ ] Authorization (admin-only)

## Next Steps

1. Implement `AdminCandidateService` interface in `service.go`
2. Add repository methods:
   - `Create(ctx, candidate) (*Candidate, error)`
   - `Update(ctx, electionID, candidateID, updates) (*Candidate, error)`
   - `Delete(ctx, electionID, candidateID) error`
   - `UpdateStatus(ctx, electionID, candidateID, status) error`
3. Mount routes in main router
4. Add validation logic in service layer
5. Test all endpoints
