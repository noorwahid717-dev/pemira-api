# Candidate API Specification

API endpoints untuk manajemen kandidat PEMIRA UNIWA (publik & admin).

## Response Format

**Success:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": null
  }
}
```

---

## 📱 Public API (Mahasiswa)

### 1. List Kandidat

**Endpoint:**
```
GET /elections/{election_id}/candidates
```

**Auth:** Required (JWT - Role: STUDENT)

**Headers:**
```
Authorization: Bearer <JWT>
Content-Type: application/json
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | string | `PUBLISHED` | Filter by status (PUBLISHED/HIDDEN) |
| `search` | string | - | Search by name/number/tagline |
| `page` | integer | 1 | Page number |
| `limit` | integer | 10 | Items per page |
| `order_by` | string | `number` | Sort by (number/name) |

**Example Request:**
```bash
curl -X GET "https://api.pemira.uniwa.ac.id/elections/1/candidates?page=1&limit=10" \
  -H "Authorization: Bearer eyJhbGc..."
```

**Response 200 (Success):**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "election_id": 1,
        "number": 1,
        "name": "Pasangan Calon A",
        "photo_url": "https://cdn.pemira.uniwa.ac.id/candidates/1.jpg",
        "short_bio": "Mahasiswa Fakultas Teknik, aktif di BEM, Mapala, dan komunitas riset AI.",
        "tagline": "Bersama Membangun BEM yang Responsif & Transparan",
        "faculty_name": "Fakultas Teknik",
        "study_program_name": "Informatika",
        "status": "PUBLISHED",
        "stats": {
          "total_votes": 1234,
          "percentage": 45.67
        }
      },
      {
        "id": 2,
        "election_id": 1,
        "number": 2,
        "name": "Pasangan Calon B",
        "photo_url": "https://cdn.pemira.uniwa.ac.id/candidates/2.jpg",
        "short_bio": "Aktivis sosial dan ketua himpunan mahasiswa FEB.",
        "tagline": "Kolaborasi, Inovasi, Aksi Nyata",
        "faculty_name": "Fakultas Ekonomi dan Bisnis",
        "study_program_name": "Manajemen",
        "status": "PUBLISHED",
        "stats": {
          "total_votes": 1467,
          "percentage": 54.33
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total_items": 2,
      "total_pages": 1
    }
  }
}
```

**Notes:**
- `stats` will be `null` or `{total_votes: 0, percentage: 0}` if voting not started/counted
- Only `PUBLISHED` candidates are visible to students
- Sort by `number` ascending by default

---

### 2. Detail Kandidat

**Endpoint:**
```
GET /elections/{election_id}/candidates/{candidate_id}
```

**Auth:** Required (JWT - Role: STUDENT)

**Example Request:**
```bash
curl -X GET "https://api.pemira.uniwa.ac.id/elections/1/candidates/1" \
  -H "Authorization: Bearer eyJhbGc..."
```

**Response 200 (Success):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "election_id": 1,
    "number": 1,
    "name": "Pasangan Calon A",
    "photo_url": "https://cdn.pemira.uniwa.ac.id/candidates/1.jpg",
    
    "short_bio": "Mahasiswa Fakultas Teknik, aktif di BEM, Mapala, dan komunitas riset AI.",
    "long_bio": "Penjabaran lebih panjang tentang riwayat organisasi, prestasi, dan pengalaman kepemimpinan ...",
    
    "tagline": "Bersama Membangun BEM yang Responsif & Transparan",
    
    "faculty_name": "Fakultas Teknik",
    "study_program_name": "Informatika",
    "cohort_year": 2021,
    
    "vision": "Mewujudkan BEM UNIWA sebagai rumah kolaborasi yang inklusif dan berdampak bagi seluruh mahasiswa.",
    "missions": [
      "Meningkatkan transparansi anggaran dan program kerja BEM.",
      "Membangun ekosistem kolaborasi antar UKM dan komunitas.",
      "Menghadirkan kanal aspirasi yang responsif dan mudah diakses."
    ],
    
    "main_programs": [
      {
        "title": "UNIWA Aspiration Hub",
        "description": "Platform digital untuk aspirasi mahasiswa dengan SLA respon maksimal 3x24 jam.",
        "category": "Transparansi & Aspirasi"
      },
      {
        "title": "Festival Inovasi Mahasiswa",
        "description": "Event tahunan lintas fakultas untuk memamerkan karya dan proyek mahasiswa.",
        "category": "Pengembangan Mahasiswa"
      }
    ],
    
    "media": {
      "video_url": "https://www.youtube.com/watch?v=abc123",
      "gallery_photos": [
        "https://cdn.pemira.uniwa.ac.id/candidates/1/gallery/1.jpg",
        "https://cdn.pemira.uniwa.ac.id/candidates/1/gallery/2.jpg"
      ],
      "document_manifesto_url": "https://cdn.pemira.uniwa.ac.id/candidates/1/visi-misi.pdf"
    },
    
    "social_links": [
      {
        "platform": "instagram",
        "url": "https://instagram.com/paslon_a"
      },
      {
        "platform": "tiktok",
        "url": "https://tiktok.com/@paslon_a"
      }
    ],
    
    "status": "PUBLISHED",
    
    "stats": {
      "total_votes": 1234,
      "percentage": 45.67
    }
  }
}
```

**Notes:**
- Only `PUBLISHED` candidates are accessible
- `stats` may be hidden until voting ends (configurable)
- All media URLs are optional (can be null)

---

## 🔒 Admin API

### 3. List Kandidat (Admin)

**Endpoint:**
```
GET /admin/elections/{election_id}/candidates
```

**Auth:** Required (JWT - Role: ADMIN/PANITIA)

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | string | - | Filter by status (all visible if omitted) |
| `search` | string | - | Search by name/number |
| `page` | integer | 1 | Page number |
| `limit` | integer | 20 | Items per page |
| `order_by` | string | `number` | Sort by (number/name/created_at) |
| `order` | string | `asc` | Sort direction (asc/desc) |

**Response 200 (Success):**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "election_id": 1,
        "number": 1,
        "name": "Pasangan Calon A",
        "photo_url": "https://cdn.pemira.uniwa.ac.id/candidates/1.jpg",
        "short_bio": "...",
        "tagline": "...",
        "status": "PUBLISHED",
        "stats": {
          "total_votes": 1234,
          "percentage": 45.67
        },
        "created_at": "2025-01-15T08:00:00Z",
        "updated_at": "2025-01-20T10:30:00Z",
        "created_by": "admin@uniwa.ac.id",
        "updated_by": "panitia1@uniwa.ac.id"
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

**Notes:**
- Admin can see all statuses (DRAFT, PUBLISHED, HIDDEN, ARCHIVED)
- Includes audit fields (created_at, updated_at, created_by, updated_by)

---

### 4. Create Kandidat

**Endpoint:**
```
POST /admin/elections/{election_id}/candidates
```

**Auth:** Required (JWT - Role: ADMIN/PANITIA)

**Request Body:**
```json
{
  "number": 1,
  "name": "Pasangan Calon A",
  "photo_url": "https://cdn.pemira.uniwa.ac.id/candidates/1.jpg",
  "short_bio": "Mahasiswa Fakultas Teknik, aktif di BEM...",
  "long_bio": "Penjabaran lengkap...",
  "tagline": "Bersama Membangun BEM yang Responsif",
  "faculty_name": "Fakultas Teknik",
  "study_program_name": "Informatika",
  "cohort_year": 2021,
  "vision": "Mewujudkan BEM UNIWA...",
  "missions": [
    "Meningkatkan transparansi...",
    "Membangun ekosistem kolaborasi..."
  ],
  "main_programs": [
    {
      "title": "UNIWA Aspiration Hub",
      "description": "Platform digital...",
      "category": "Transparansi & Aspirasi"
    }
  ],
  "media": {
    "video_url": "https://www.youtube.com/watch?v=abc123",
    "gallery_photos": [],
    "document_manifesto_url": null
  },
  "social_links": [
    {
      "platform": "instagram",
      "url": "https://instagram.com/paslon_a"
    }
  ],
  "status": "DRAFT"
}
```

**Response 201 (Created):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "election_id": 1,
    "number": 1,
    "name": "Pasangan Calon A",
    "status": "DRAFT",
    "created_at": "2025-01-15T08:00:00Z"
  }
}
```

**Validation Rules:**
- `number`: required, unique per election, positive integer
- `name`: required, min 3 chars, max 255 chars
- `photo_url`: required, valid URL
- `short_bio`: required, max 500 chars
- `tagline`: required, max 200 chars
- `status`: enum (DRAFT, PUBLISHED, HIDDEN, ARCHIVED)
- `missions`: array, min 3 items, max 10 items
- `main_programs`: array, min 2 items, max 10 items

---

### 5. Update Kandidat

**Endpoint:**
```
PUT /admin/candidates/{candidate_id}
```

**Auth:** Required (JWT - Role: ADMIN/PANITIA)

**Request Body:** (all fields optional, partial update)
```json
{
  "name": "Pasangan Calon A (Updated)",
  "tagline": "New tagline...",
  "status": "PUBLISHED"
}
```

**Response 200 (Success):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "election_id": 1,
    "number": 1,
    "name": "Pasangan Calon A (Updated)",
    "status": "PUBLISHED",
    "updated_at": "2025-01-20T10:30:00Z"
  }
}
```

---

### 6. Delete Kandidat

**Endpoint:**
```
DELETE /admin/candidates/{candidate_id}
```

**Auth:** Required (JWT - Role: ADMIN)

**Response 200 (Success):**
```json
{
  "success": true,
  "data": {
    "message": "Kandidat berhasil dihapus."
  }
}
```

**Notes:**
- Soft delete (set status to ARCHIVED) if votes exist
- Hard delete if no votes recorded

---

### 7. Publish Kandidat

**Endpoint:**
```
POST /admin/candidates/{candidate_id}/publish
```

**Auth:** Required (JWT - Role: ADMIN/PANITIA)

**Response 200 (Success):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "status": "PUBLISHED",
    "published_at": "2025-01-20T10:30:00Z"
  }
}
```

---

### 8. Unpublish Kandidat

**Endpoint:**
```
POST /admin/candidates/{candidate_id}/unpublish
```

**Auth:** Required (JWT - Role: ADMIN/PANITIA)

**Response 200 (Success):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "status": "HIDDEN",
    "unpublished_at": "2025-01-20T11:00:00Z"
  }
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid JWT token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `ELECTION_NOT_FOUND` | 404 | Election ID tidak ditemukan |
| `CANDIDATE_NOT_FOUND` | 404 | Candidate tidak ditemukan |
| `VALIDATION_ERROR` | 422 | Invalid request body/params |
| `DUPLICATE_NUMBER` | 409 | Nomor kandidat sudah digunakan |
| `INTERNAL_ERROR` | 500 | Server error |

**Error Response Example:**
```json
{
  "success": false,
  "error": {
    "code": "CANDIDATE_NOT_FOUND",
    "message": "Kandidat tidak ditemukan untuk pemilu ini.",
    "details": null
  }
}
```

---

## Frontend Integration

### Halaman Daftar Kandidat (Student)

```javascript
// Fetch candidates list
const response = await fetch('/elections/1/candidates?page=1&limit=10', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});

const { data } = await response.json();
// Render data.items as cards
```

### Halaman Detail Kandidat (Student)

```javascript
// Fetch candidate detail
const response = await fetch('/elections/1/candidates/1', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const { data: candidate } = await response.json();
// Render full profile with vision, mission, programs
```

### Admin CMS - Create Kandidat

```javascript
// Create new candidate
const response = await fetch('/admin/elections/1/candidates', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${adminToken}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    number: 1,
    name: "Paslon A",
    photo_url: "https://...",
    short_bio: "...",
    // ... other fields
    status: "DRAFT"
  })
});

const { data } = await response.json();
console.log('Created candidate:', data.id);
```

---

## Database Schema (Candidate Table)

```sql
CREATE TABLE candidates (
    id                      BIGSERIAL PRIMARY KEY,
    election_id             BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    
    -- Basic info
    number                  INTEGER NOT NULL,
    name                    TEXT NOT NULL,
    photo_url               TEXT NOT NULL,
    
    -- Bio
    short_bio               TEXT NOT NULL,
    long_bio                TEXT,
    tagline                 TEXT NOT NULL,
    
    -- Academic info
    faculty_name            TEXT,
    study_program_name      TEXT,
    cohort_year             INTEGER,
    
    -- Vision & Mission
    vision                  TEXT,
    missions                JSONB,  -- Array of strings
    
    -- Programs
    main_programs           JSONB,  -- Array of {title, description, category}
    
    -- Media
    media                   JSONB,  -- {video_url, gallery_photos[], document_manifesto_url}
    
    -- Social links
    social_links            JSONB,  -- Array of {platform, url}
    
    -- Status
    status                  candidate_status NOT NULL DEFAULT 'DRAFT',
    
    -- Audit
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_id           BIGINT REFERENCES user_accounts(id),
    updated_by_id           BIGINT REFERENCES user_accounts(id),
    
    CONSTRAINT ux_candidates_election_number UNIQUE (election_id, number)
);

CREATE INDEX idx_candidates_election ON candidates(election_id);
CREATE INDEX idx_candidates_status ON candidates(status);
```

---

## Implementation Checklist

### Backend (Go)

- [ ] Create `internal/candidate` package
- [ ] Define `Candidate` model with JSONB fields
- [ ] Implement repository layer (CRUD + filters)
- [ ] Implement service layer (business logic + permissions)
- [ ] Create public handler (2 endpoints)
- [ ] Create admin handler (6 endpoints)
- [ ] Add validation middleware
- [ ] Add auth middleware (role-based)
- [ ] Write unit tests
- [ ] Write integration tests

### Frontend

- [ ] Create candidate list page (student)
- [ ] Create candidate detail page (student)
- [ ] Create admin CMS pages (CRUD)
- [ ] Add image upload flow
- [ ] Add form validation
- [ ] Add loading states
- [ ] Add error handling

### DevOps

- [ ] Add CDN for media storage
- [ ] Setup image optimization
- [ ] Add API rate limiting
- [ ] Monitor endpoint performance
