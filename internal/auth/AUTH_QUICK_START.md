# Auth System Quick Start

## 🎯 Overview

Complete JWT-based authentication system dengan 3 roles:
- **STUDENT** (Mahasiswa/Pemilih)
- **ADMIN** (Admin Universitas/Panitia)
- **TPS_OPERATOR** (Petugas TPS)

## 🚀 Setup

### 1. Run Migration

```bash
# Apply migration
migrate -path migrations -database "postgresql://user:pass@localhost/pemira?sslmode=disable" up

# Or using make
make migrate-up
```

### 2. Initialize Auth Module

```go
package main

import (
    "time"
    "pemira-api/internal/auth"
)

func setupAuth(db *pgxpool.Pool) (*auth.AuthHandler, *auth.JWTManager) {
    // JWT Config
    jwtConfig := auth.JWTConfig{
        Secret:           os.Getenv("JWT_SECRET"), // "your-secret-key"
        AccessTokenTTL:   30 * time.Minute,
        RefreshTokenTTL:  7 * 24 * time.Hour,
    }
    
    // Initialize components
    jwtManager := auth.NewJWTManager(jwtConfig)
    repo := auth.NewPgRepository(db)
    authService := auth.NewAuthService(repo, jwtManager, jwtConfig)
    authHandler := auth.NewAuthHandler(authService)
    
    return authHandler, jwtManager
}
```

### 3. Mount Routes

```go
func main() {
    authHandler, jwtManager := setupAuth(pool)
    
    r := chi.NewRouter()
    
    // Public auth endpoints (no middleware)
    r.Post("/auth/login", authHandler.Login)
    r.Post("/auth/refresh", authHandler.RefreshToken)
    r.Post("/auth/logout", authHandler.Logout)
    
    // Protected endpoint
    r.Group(func(g chi.Router) {
        g.Use(middleware.JWTAuth(jwtManager))
        g.Get("/auth/me", authHandler.Me)
    })
    
    // Student routes
    r.Group(func(g chi.Router) {
        g.Use(middleware.AuthStudentOnly(jwtManager))
        g.Get("/elections/{id}/candidates", candidateHandler.ListPublic)
        // ... other student endpoints
    })
    
    // Admin routes
    r.Route("/admin", func(ad chi.Router) {
        ad.Use(middleware.AuthAdminOnly(jwtManager))
        ad.Post("/elections", electionHandler.Create)
        // ... other admin endpoints
    })
    
    // TPS routes
    r.Route("/tps", func(tr chi.Router) {
        tr.Use(middleware.AuthTPSOperatorOnly(jwtManager))
        tr.Post("/checkin/scan", tpsHandler.ScanCheckin)
        // ... other TPS endpoints
    })
}
```

## 📚 API Endpoints

### 1. Login (All Roles)

```bash
POST /auth/login
Content-Type: application/json

{
  "username": "22012345",  # NIM for students, username for others
  "password": "password123"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "random-token-here",
  "token_type": "Bearer",
  "expires_in": 1800,
  "user": {
    "id": 10,
    "username": "22012345",
    "role": "STUDENT",
    "voter_id": 123,
    "tps_id": null,
    "profile": {
      "name": "Budi Setiawan",
      "faculty_name": "Fakultas Teknik",
      "study_program_name": "Informatika",
      "cohort_year": 2021
    }
  }
}
```

### 2. Refresh Token

```bash
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "your-refresh-token"
}
```

**Response:**
```json
{
  "access_token": "new-jwt-token",
  "refresh_token": "new-refresh-token",
  "token_type": "Bearer",
  "expires_in": 1800
}
```

### 3. Logout

```bash
POST /auth/logout
Content-Type: application/json

{
  "refresh_token": "your-refresh-token"
}
```

**Response:**
```json
{
  "message": "Logged out successfully."
}
```

### 4. Get Current User

```bash
GET /auth/me
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "id": 10,
  "username": "22012345",
  "role": "STUDENT",
  "voter_id": 123,
  "profile": {
    "name": "Budi Setiawan",
    "faculty_name": "Fakultas Teknik",
    "study_program_name": "Informatika",
    "cohort_year": 2021
  }
}
```

## 🔐 JWT Token Structure

### Access Token Claims

```json
{
  "sub": 10,              # user_id
  "role": "STUDENT",
  "voter_id": 123,        # optional, for STUDENT
  "tps_id": 3,            # optional, for TPS_OPERATOR
  "exp": 1717578000,
  "iat": 1717576200
}
```

## 🧪 Testing

### 1. Create Test User (Student)

```sql
-- Insert test voter
INSERT INTO voters (nim, name, email, faculty_name, study_program_name, cohort_year, gender)
VALUES ('22012345', 'Budi Setiawan', 'budi@univ.ac.id', 'Fakultas Teknik', 'Informatika', 2021, 'M')
RETURNING id; -- assume returns 123

-- Create user account
INSERT INTO user_accounts (username, password_hash, role, voter_id, is_active)
VALUES (
    '22012345',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyKdprsKrEQC', -- password: "password123"
    'STUDENT',
    123,
    true
);
```

### 2. Test Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "22012345",
    "password": "password123"
  }'
```

### 3. Test Protected Endpoint

```bash
# Get access token from login response
TOKEN="eyJhbGc..."

curl http://localhost:8080/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

## 🛡️ Security Features

1. **Password Hashing**: bcrypt with cost 12
2. **Refresh Token**: Hashed in database
3. **Session Management**: Track user agents & IPs
4. **Token Expiry**: Separate TTL for access (30m) and refresh (7d)
5. **Role-Based Access**: Dedicated middleware for each role
6. **Inactive Users**: Blocked from login

## 📝 Create Users Programmatically

```go
func createStudentUser(ctx context.Context, repo auth.Repository, nim, password string, voterID int64) error {
    // Hash password
    hashedPassword, err := auth.HashPassword(password)
    if err != nil {
        return err
    }
    
    // Create user account
    user := &auth.UserAccount{
        Username:     nim,
        PasswordHash: hashedPassword,
        Role:         constants.RoleStudent,
        VoterID:      &voterID,
        IsActive:     true,
    }
    
    _, err = repo.CreateUserAccount(ctx, user)
    return err
}

func createAdminUser(ctx context.Context, repo auth.Repository, username, password string) error {
    hashedPassword, err := auth.HashPassword(password)
    if err != nil {
        return err
    }
    
    user := &auth.UserAccount{
        Username:     username,
        PasswordHash: hashedPassword,
        Role:         constants.RoleAdmin,
        IsActive:     true,
    }
    
    _, err = repo.CreateUserAccount(ctx, user)
    return err
}
```

## 🔧 Environment Variables

```env
JWT_SECRET=your-super-secret-key-here
ACCESS_TOKEN_TTL=30m
REFRESH_TOKEN_TTL=168h  # 7 days
```

## ⚠️ Important Notes

1. **JWT Secret**: Must be strong and kept secret
2. **HTTPS**: Always use HTTPS in production
3. **Token Storage**: 
   - Frontend: Store access token in memory
   - Frontend: Store refresh token in httpOnly cookie or secure storage
4. **Refresh Strategy**: Auto-refresh before access token expires
5. **Logout**: Always revoke refresh token on logout

## 🐛 Troubleshooting

### "Invalid token"
- Check if token is expired
- Verify JWT_SECRET matches
- Ensure Bearer token format

### "User not found"
- Check if user exists in user_accounts table
- Verify username is correct

### "Access denied"
- Check user role matches endpoint requirement
- Verify token has correct role claim

## 🎯 Next Steps

- [ ] Add forgot password flow
- [ ] Add email verification
- [ ] Add 2FA support
- [ ] Add password change endpoint
- [ ] Add session list & revoke all
- [ ] Add login history/audit log
