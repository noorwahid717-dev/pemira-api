# Quick Start: Candidate Module

## 🚀 Setup & Usage

### 1. Initialize Service

```go
package main

import (
    "pemira-api/internal/candidate"
    "github.com/jackc/pgx/v5/pgxpool"
)

func setupCandidateModule(db *pgxpool.Pool) (*candidate.Handler, *candidate.AdminHandler) {
    // Repository
    repo := candidate.NewPgCandidateRepository(db)
    
    // Stats provider (voting statistics)
    stats := candidate.NewPgStatsProvider(db)
    
    // Service (implements both public & admin interfaces)
    service := candidate.NewService(repo, stats)
    
    // Handlers
    publicHandler := candidate.NewHandler(service)
    adminHandler := candidate.NewAdminHandler(service)
    
    return publicHandler, adminHandler
}
```

### 2. Mount to Router

```go
func main() {
    // ... setup database, logger, etc.
    
    publicHandler, adminHandler := setupCandidateModule(pool)
    
    r := chi.NewRouter()
    
    // Public routes (for students/voters)
    r.Route("/api/v1/elections/{electionID}", func(er chi.Router) {
        er.Get("/candidates", publicHandler.ListPublic)
        er.Get("/candidates/{candidateID}", publicHandler.DetailPublic)
    })
    
    // Admin routes (for admin panel)
    r.Route("/api/v1/admin", func(ad chi.Router) {
        ad.Use(middleware.AuthAdminOnly) // Your auth middleware
        
        // List & Create
        ad.Route("/elections/{electionID}", func(er chi.Router) {
            er.Get("/candidates", adminHandler.List)
            er.Post("/candidates", adminHandler.Create)
        })
        
        // Single candidate operations
        ad.Route("/candidates/{candidateID}", func(cr chi.Router) {
            cr.Get("/", adminHandler.Detail)
            cr.Put("/", adminHandler.Update)
            cr.Delete("/", adminHandler.Delete)
            cr.Post("/publish", adminHandler.Publish)
            cr.Post("/unpublish", adminHandler.Unpublish)
        })
    })
    
    http.ListenAndServe(":8080", r)
}
```

### 3. Test Endpoints

#### List Public Candidates (Students)
```bash
curl http://localhost:8080/api/v1/elections/1/candidates
```

Response:
```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "election_id": 1,
        "number": 1,
        "name": "Pasangan Calon A",
        "photo_url": "https://...",
        "short_bio": "Mahasiswa Fakultas Teknik...",
        "tagline": "Bersama Membangun BEM",
        "faculty_name": "Fakultas Teknik",
        "study_program_name": "Informatika",
        "status": "PUBLISHED",
        "stats": {
          "total_votes": 1234,
          "percentage": 45.67
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total_items": 3,
      "total_pages": 1
    }
  }
}
```

#### Create Candidate (Admin)
```bash
curl -X POST http://localhost:8080/api/v1/admin/elections/1/candidates \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "number": 1,
    "name": "Pasangan Calon A",
    "photo_url": "https://example.com/photo.jpg",
    "short_bio": "Mahasiswa Fakultas Teknik",
    "long_bio": "Mahasiswa Informatika angkatan 2021...",
    "tagline": "Bersama Membangun BEM yang Responsif",
    "faculty_name": "Fakultas Teknik",
    "study_program_name": "Informatika",
    "cohort_year": 2021,
    "vision": "Mewujudkan BEM yang inklusif...",
    "missions": [
      "Meningkatkan komunikasi mahasiswa",
      "Mengoptimalkan layanan BEM"
    ],
    "main_programs": [
      {
        "title": "Program Beasiswa",
        "description": "Menambah akses beasiswa mahasiswa",
        "category": "Pendidikan"
      }
    ],
    "media": {
      "video_url": "https://youtube.com/watch?v=...",
      "gallery_photos": ["https://example.com/photo1.jpg"],
      "document_manifesto_url": "https://example.com/manifesto.pdf"
    },
    "social_links": [
      {
        "platform": "instagram",
        "url": "https://instagram.com/paslon1"
      }
    ],
    "status": "DRAFT"
  }'
```

#### Publish Candidate (Admin)
```bash
curl -X POST http://localhost:8080/api/v1/admin/candidates/1/publish?election_id=1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Response: Full candidate detail with `status: "PUBLISHED"`

#### Update Candidate (Admin)
```bash
curl -X PUT http://localhost:8080/api/v1/admin/candidates/1?election_id=1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pasangan Calon A (Updated)",
    "tagline": "New tagline"
  }'
```

#### Delete Candidate (Admin)
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/candidates/1?election_id=1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Response: `204 No Content`

## 📚 Advanced Usage

### Custom Stats Provider

If you need custom voting statistics logic:

```go
type CustomStatsProvider struct {
    // your dependencies
}

func (c *CustomStatsProvider) GetCandidateStats(ctx context.Context, electionID int64) (candidate.CandidateStatsMap, error) {
    // Your custom logic
    return candidate.CandidateStatsMap{
        1: candidate.CandidateStats{TotalVotes: 100, Percentage: 50.0},
    }, nil
}

// Use it
stats := &CustomStatsProvider{}
service := candidate.NewService(repo, stats)
```

### Pagination & Search

```bash
# Search by name
curl "http://localhost:8080/api/v1/elections/1/candidates?search=teknik&page=1&limit=10"

# Admin: filter by status
curl "http://localhost:8080/api/v1/admin/elections/1/candidates?status=DRAFT&page=1&limit=20" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### Status Workflow

```
DRAFT (created) 
  ↓ (publish)
PUBLISHED (visible to students)
  ↓ (unpublish)
HIDDEN (not visible)
  ↓ (publish again)
PUBLISHED
  ↓ (manual update)
ARCHIVED (soft delete)
```

## 🔒 Security Notes

1. **Authentication Required**: 
   - Public endpoints: No auth (read-only)
   - Admin endpoints: JWT token + ADMIN role

2. **Authorization**:
   - Implement `middleware.AuthAdminOnly` to check JWT claims
   - Verify role = "ADMIN" before allowing admin operations

3. **Input Validation**:
   - Candidate number must be unique per election
   - Name is required
   - Status must be valid enum value

## 🐛 Troubleshooting

### "Candidate number already used"
- Check existing candidates in the election
- Each candidate must have unique number within an election

### "Candidate not found"
- Verify election_id and candidate_id combination
- Students can only see PUBLISHED candidates

### "Unauthorized"
- Check JWT token is valid
- Verify token has ADMIN role for admin endpoints

## 📊 Performance Tips

1. **Pagination**: Use reasonable limits (default: 10 for public, 20 for admin)
2. **Caching**: Cache published candidates list (updates are rare)
3. **Indexes**: Ensure DB indexes on `(election_id, status, number)`
4. **Stats**: Stats calculation can be expensive, consider caching

## 🎯 What's Next?

- [ ] Add file upload for photos
- [ ] Add batch import/export
- [ ] Add candidate approval workflow
- [ ] Add change history/audit log
- [ ] Add email notifications on publish

## 💡 Tips

- Keep candidates in DRAFT until all content is ready
- Use HIDDEN instead of DELETE to preserve voting history
- Test thoroughly before publishing to students
- Use pagination for large elections (>50 candidates)
