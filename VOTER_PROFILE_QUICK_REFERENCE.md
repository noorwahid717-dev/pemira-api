# Voter Profile API - Quick Reference

## 🚀 Quick Start

```bash
# Run migrations
make migrate-up

# Build and run
go build -o pemira-api cmd/api/main.go
./pemira-api

# Test endpoints
./test-voter-profile.sh <your_voter_token>
```

---

## 📡 Endpoints Summary

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/v1/voters/me/complete-profile` | Get complete voter profile | Voter |
| PUT | `/api/v1/voters/me/profile` | Update profile (email, phone, photo) | Voter |
| PUT | `/api/v1/voters/me/voting-method` | Change voting method preference | Voter |
| POST | `/api/v1/voters/me/change-password` | Change account password | Voter |
| GET | `/api/v1/voters/me/participation-stats` | Get participation statistics | Voter |
| DELETE | `/api/v1/voters/me/photo` | Delete profile photo | Voter |

---

## 📋 cURL Examples

### Get Complete Profile
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/voters/me/complete-profile
```

### Update Profile
```bash
curl -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"new@email.com","phone":"08123456789"}' \
  http://localhost:8080/api/v1/voters/me/profile
```

### Change Password
```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "oldpass",
    "new_password": "newpass123",
    "confirm_password": "newpass123"
  }' \
  http://localhost:8080/api/v1/voters/me/change-password
```

### Update Voting Method
```bash
curl -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"election_id":1,"preferred_method":"ONLINE"}' \
  http://localhost:8080/api/v1/voters/me/voting-method
```

### Get Participation Stats
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/voters/me/participation-stats
```

### Delete Photo
```bash
curl -X DELETE \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/voters/me/photo
```

---

## 🔑 Get Test Token

### Login as voter:
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"2021010001","password":"password123"}' \
  http://localhost:8080/api/v1/auth/login | jq -r '.access_token'
```

Save the token:
```bash
TOKEN=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"2021010001","password":"password123"}' \
  http://localhost:8080/api/v1/auth/login | jq -r '.access_token')
```

---

## ✅ Validation Rules

### Email
- Must be valid email format
- Example: `user@uniwa.ac.id`

### Phone
- Must be Indonesian format
- Formats: `08xxxxxxxxx` or `+62xxxxxxxxx`
- Example: `081234567890` or `+6281234567890`

### Password
- Minimum 8 characters
- Cannot be same as current password
- Must match confirmation

### Voting Method
- Must be either `ONLINE` or `TPS`
- Cannot change if already voted
- Cannot change to ONLINE if already checked in at TPS

---

## 🐛 Common Errors

### 401 Unauthorized
```json
{
  "error": "UNAUTHORIZED",
  "message": "Token tidak valid atau tidak ditemukan."
}
```
**Solution:** Check if token is valid and not expired

### 403 Forbidden
```json
{
  "error": "FORBIDDEN",
  "message": "Hanya voter yang dapat mengakses profil."
}
```
**Solution:** Login with voter account, not admin/tps

### 400 Invalid Email
```json
{
  "error": "INVALID_EMAIL",
  "message": "Format email tidak valid."
}
```
**Solution:** Use valid email format

### 400 Already Voted
```json
{
  "error": "ALREADY_VOTED",
  "message": "Tidak dapat mengubah metode voting karena sudah voting."
}
```
**Solution:** Cannot change voting method after voting

---

## 📦 Response Structures

### Complete Profile Response
```typescript
{
  personal_info: {
    voter_id: number
    name: string
    username: string
    email: string | null
    phone: string | null
    faculty_name: string
    study_program_name: string
    cohort_year: number
    semester: string
    photo_url: string | null
  }
  voting_info: {
    preferred_method: "ONLINE" | "TPS" | null
    has_voted: boolean
    voted_at: string | null
    tps_name: string | null
    tps_location: string | null
  }
  participation: {
    total_elections: number
    participated_elections: number
    participation_rate: number
    last_participation: string | null
  }
  account_info: {
    created_at: string
    last_login: string | null
    login_count: number
    account_status: "active" | "inactive"
  }
}
```

### Success Response
```typescript
{
  success: true
  message: string
  [additional_fields]: any
}
```

### Error Response
```typescript
{
  error: string
  message: string
}
```

---

## 🗄️ Database Tables Used

- `voters` - Voter personal information
- `user_accounts` - User authentication
- `voter_status` - Voting status per election
- `elections` - Election data
- `tps` - TPS information

---

## 📝 Files Overview

| File | Purpose |
|------|---------|
| `internal/voter/profile_handler.go` | HTTP handlers |
| `internal/voter/service.go` | Business logic |
| `internal/voter/repository_pgx.go` | Database queries |
| `internal/voter/dto.go` | Request/Response DTOs |
| `internal/voter/entity.go` | Domain entities |
| `cmd/api/main.go` | Routes registration |

---

## 🔧 Development

### Add New Endpoint
1. Add method to `repository.go` interface
2. Implement in `repository_pgx.go`
3. Add business logic to `service.go`
4. Create handler in `profile_handler.go`
5. Register route in `profile_handler.go` `RegisterRoutes()`

### Run Tests
```bash
# Unit tests (when available)
go test ./internal/voter/...

# Integration tests
./test-voter-profile.sh $TOKEN
```

---

## 📚 Related Documentation

- `VOTER_PROFILE_API_IMPLEMENTATION.md` - Full implementation details
- `AUTH_QUICK_REFERENCE.md` - Authentication guide
- Main spec document - Original requirements

---

**Quick Tips:**
- Use `jq` for pretty JSON output
- Set `TOKEN` env var for easier testing
- Check logs with `tail -f api.log`
- Database migrations in `migrations/023_*`

**Support:** Backend implementation complete and ready for frontend integration! ✅
