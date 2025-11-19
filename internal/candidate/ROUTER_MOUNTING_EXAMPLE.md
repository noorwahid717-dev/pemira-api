# Router Mounting Example

## Cara Mount Candidate Handlers ke Router

Karena project ini belum punya struktur router terpusat, berikut contoh cara mount handler kandidat (public & admin) ke router utama di `cmd/api/main.go`.

### Option 1: Mount Langsung di main.go

```go
package main

import (
    // ... imports lain
    "pemira-api/internal/candidate"
    "pemira-api/internal/http/middleware" // jika sudah ada auth middleware
)

func main() {
    // ... setup config, database, logger

    // Initialize repositories
    candidateRepo := candidate.NewPgCandidateRepository(pool)
    candidateStatsProvider := candidate.NewPgCandidateStatsAdapter(pool)
    
    // Initialize service
    candidateService := candidate.NewService(candidateRepo, candidateStatsProvider)
    
    // Initialize handlers
    publicCandidateHandler := candidate.NewHandler(candidateService)
    adminCandidateHandler := candidate.NewAdminHandler(candidateService)

    r := chi.NewRouter()
    
    // ... global middleware

    r.Route("/api/v1", func(r chi.Router) {
        // Public routes (mahasiswa)
        r.Group(func(g chi.Router) {
            // g.Use(middleware.AuthStudentOnly) // jika sudah ada
            
            g.Route("/elections/{electionID}", func(er chi.Router) {
                er.Get("/candidates", publicCandidateHandler.ListPublic)
                er.Get("/candidates/{candidateID}", publicCandidateHandler.DetailPublic)
            })
        })

        // Admin routes
        r.Route("/admin", func(ad chi.Router) {
            // ad.Use(middleware.AuthAdminOnly) // jika sudah ada
            
            // List & Create
            ad.Route("/elections/{electionID}", func(er chi.Router) {
                er.Get("/candidates", adminCandidateHandler.List)
                er.Post("/candidates", adminCandidateHandler.Create)
            })
            
            // Detail, Update, Delete, Publish, Unpublish
            ad.Route("/candidates/{candidateID}", func(cr chi.Router) {
                cr.Get("/", adminCandidateHandler.Detail)
                cr.Put("/", adminCandidateHandler.Update)
                cr.Delete("/", adminCandidateHandler.Delete)
                cr.Post("/publish", adminCandidateHandler.Publish)
                cr.Post("/unpublish", adminCandidateHandler.Unpublish)
            })
        })
    })

    // ... start server
}
```

### Option 2: Buat Router Package (Recommended)

Buat file `internal/api/router.go`:

```go
package api

import (
    "github.com/go-chi/chi/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    
    "pemira-api/internal/candidate"
    "pemira-api/internal/http/middleware"
)

type Dependencies struct {
    DB *pgxpool.Pool
    // tambahkan dependencies lain jika perlu
}

func NewRouter(deps Dependencies) chi.Router {
    r := chi.NewRouter()
    
    // Initialize candidate module
    candidateRepo := candidate.NewPgCandidateRepository(deps.DB)
    candidateStatsProvider := candidate.NewPgCandidateStatsAdapter(deps.DB)
    candidateService := candidate.NewService(candidateRepo, candidateStatsProvider)
    
    publicCandidateHandler := candidate.NewHandler(candidateService)
    adminCandidateHandler := candidate.NewAdminHandler(candidateService)

    r.Route("/api/v1", func(r chi.Router) {
        // Public routes
        r.Group(func(g chi.Router) {
            // g.Use(middleware.AuthStudentOnly)
            
            g.Route("/elections/{electionID}", func(er chi.Router) {
                er.Get("/candidates", publicCandidateHandler.ListPublic)
                er.Get("/candidates/{candidateID}", publicCandidateHandler.DetailPublic)
            })
        })

        // Admin routes
        r.Route("/admin", func(ad chi.Router) {
            // ad.Use(middleware.AuthAdminOnly)
            
            ad.Route("/elections/{electionID}", func(er chi.Router) {
                er.Get("/candidates", adminCandidateHandler.List)
                er.Post("/candidates", adminCandidateHandler.Create)
            })
            
            ad.Route("/candidates/{candidateID}", func(cr chi.Router) {
                cr.Get("/", adminCandidateHandler.Detail)
                cr.Put("/", adminCandidateHandler.Update)
                cr.Delete("/", adminCandidateHandler.Delete)
                cr.Post("/publish", adminCandidateHandler.Publish)
                cr.Post("/unpublish", adminCandidateHandler.Unpublish)
            })
        })
    })
    
    return r
}
```

Lalu di `cmd/api/main.go`:

```go
import "pemira-api/internal/api"

func main() {
    // ... setup

    deps := api.Dependencies{
        DB: pool,
    }
    
    apiRouter := api.NewRouter(deps)
    
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    // ... middleware lain
    
    r.Mount("/", apiRouter)
    
    // ... start server
}
```

## Endpoint Summary

### Public (Mahasiswa)
- `GET /api/v1/elections/{electionID}/candidates` - List kandidat published
- `GET /api/v1/elections/{electionID}/candidates/{candidateID}` - Detail kandidat

### Admin
- `GET /api/v1/admin/elections/{electionID}/candidates` - List semua kandidat
- `POST /api/v1/admin/elections/{electionID}/candidates` - Create kandidat
- `GET /api/v1/admin/candidates/{candidateID}?election_id=X` - Detail kandidat
- `PUT /api/v1/admin/candidates/{candidateID}?election_id=X` - Update kandidat
- `DELETE /api/v1/admin/candidates/{candidateID}?election_id=X` - Delete kandidat
- `POST /api/v1/admin/candidates/{candidateID}/publish?election_id=X` - Publish kandidat
- `POST /api/v1/admin/candidates/{candidateID}/unpublish?election_id=X` - Unpublish kandidat

## Middleware Required

Pastikan middleware auth sudah ada:

```go
// internal/http/middleware/auth.go

func AuthStudentOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validasi JWT token
        // Check role = STUDENT
        next.ServeHTTP(w, r)
    })
}

func AuthAdminOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validasi JWT token
        // Check role = ADMIN
        next.ServeHTTP(w, r)
    })
}
```
