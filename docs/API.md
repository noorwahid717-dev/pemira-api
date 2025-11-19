# PEMIRA API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
All protected endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

---

## 1. Authentication Endpoints

### POST /auth/login/student
Login untuk mahasiswa
```json
Request:
{
  "username": "12345678",  // NIM
  "password": "2000-01-01" // Date of Birth atau SSO
}

Response:
{
  "data": {
    "token": "eyJhbGc...",
    "user": {
      "id": 1,
      "username": "12345678",
      "role": "STUDENT",
      "full_name": "John Doe"
    }
  }
}
```

### POST /auth/login/admin
Login untuk admin/panitia
```json
Request:
{
  "username": "admin",
  "password": "password"
}
```

### GET /auth/me
Get current user info (Protected)

---

## 2. Election Endpoints

### GET /elections/current
Get election yang sedang aktif
```json
Response:
{
  "data": {
    "id": 1,
    "name": "PEMIRA 2024",
    "voting_mode": "HYBRID",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-12-31T23:59:59Z"
  }
}
```

### GET /elections/{id}
Get election detail

### GET /elections/{id}/candidates
Get daftar kandidat
```json
Response:
{
  "data": [
    {
      "id": 1,
      "order_number": 1,
      "name": "Candidate 1",
      "vision_mission": "...",
      "photo_url": "https://..."
    }
  ]
}
```

---

## 3. Voter Endpoints

### GET /me/voter-status (Protected - Student)
Cek status pemilih
```json
Response:
{
  "data": {
    "voter": {
      "nim": "12345678",
      "full_name": "John Doe",
      "faculty": "Teknik"
    },
    "status": {
      "status": "ELIGIBLE",
      "has_voted": false
    }
  }
}
```

### GET /admin/dpt (Protected - Admin)
Get DPT list dengan pagination
```
Query params:
- page: int (default: 1)
- per_page: int (default: 10)
- search: string
- faculty: string
```

### POST /admin/dpt/import (Protected - Admin)
Import DPT dari file CSV/XLSX

---

## 4. Voting Endpoints

### POST /voting/online/cast (Protected - Student)
Cast vote untuk online voting
```json
Request:
{
  "election_id": 1,
  "candidate_id": 1
}

Response:
{
  "data": {
    "success": true,
    "token": "a1b2c3d4..." // Vote receipt token
  }
}
```

### POST /voting/tps/cast (Protected - Student)
Cast vote setelah TPS check-in approved

---

## 5. TPS Endpoints

### POST /tps/checkin/scan (Protected - Student)
Scan QR TPS
```json
Request:
{
  "tps_id": 1,
  "voter_id": 123
}

Response:
{
  "data": {
    "id": 1,
    "status": "PENDING",
    "checked_in_at": "2024-01-01T10:00:00Z"
  }
}
```

### GET /tps/{id}/checkins (Protected - TPS Operator)
Get pending check-ins
```
Query params:
- status: PENDING | APPROVED | REJECTED
```

### POST /tps/{id}/checkins/{checkinId}/approve (Protected - TPS Operator)
Approve check-in
```json
Response:
{
  "data": {
    "success": true
  }
}
```

### GET /admin/tps (Protected - Admin)
Get all TPS

### POST /admin/tps (Protected - Admin)
Create TPS
```json
Request:
{
  "election_id": 1,
  "name": "TPS 1",
  "location": "Gedung A"
}
```

---

## 6. Monitoring Endpoints

### GET /admin/monitoring/summary (Protected - Admin)
Get dashboard summary
```json
Response:
{
  "data": {
    "total_votes": 1000,
    "total_eligible": 5000,
    "participation_pct": 20.0,
    "candidate_votes": {
      "1": 500,
      "2": 300,
      "3": 200
    }
  }
}
```

### GET /admin/monitoring/live-count/{electionID} (Protected - Admin)
Get live count snapshot dengan detail lengkap

---

## 7. Announcement Endpoints

### GET /announcements
Get published announcements
```
Query params:
- election_id: int
- page: int
- per_page: int
```

### POST /admin/announcements (Protected - Admin)
Create announcement
```json
Request:
{
  "election_id": 1,
  "title": "Pengumuman Penting",
  "content": "...",
  "type": "INFO",
  "priority": 3,
  "is_published": true
}
```

---

## 8. Audit Endpoints

### GET /admin/audit-logs (Protected - Super Admin)
Get audit logs
```
Query params:
- entity_type: VOTE | ELECTION | CANDIDATE | TPS
- action: VOTE_CAST | ELECTION_UPDATED | ...
- page: int
- per_page: int
```

---

## 9. WebSocket Endpoints

### WS /ws/tps/{tpsId}
Real-time TPS check-in updates
```json
Message format:
{
  "type": "CHECKIN_PENDING",
  "channel": "tps/1",
  "data": {
    "checkin_id": 123,
    "voter_name": "John Doe",
    "checked_in_at": "2024-01-01T10:00:00Z"
  }
}
```

### WS /ws/live-count/{electionId}
Real-time vote count updates
```json
Message format:
{
  "type": "VOTE_COUNT_UPDATE",
  "channel": "live-count/1",
  "data": {
    "candidate_id": 1,
    "total_votes": 501
  }
}
```

---

## Error Response Format

All errors follow this format:
```json
{
  "code": "ERROR_CODE",
  "message": "Human readable error message",
  "details": {} // Optional additional details
}
```

Common error codes:
- `BAD_REQUEST` - Invalid request data
- `UNAUTHORIZED` - Missing or invalid token
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `INTERNAL_ERROR` - Server error
- `ALREADY_VOTED` - User already voted
- `INVALID_PHASE` - Election not in voting phase
- `VOTER_NOT_ELIGIBLE` - Voter not eligible to vote

---

## Rate Limiting

- Login endpoints: 5 requests per minute
- Voting endpoints: 3 requests per minute
- Other endpoints: 60 requests per minute

When rate limit exceeded:
```json
{
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "Too many requests"
}
```
