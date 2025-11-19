# Authentication Quick Reference

Panduan singkat penggunaan sistem autentikasi Pemira API.

## 📦 Import Packages

```go
import (
    "pemira-api/internal/auth"
    "pemira-api/internal/http/middleware"
    "pemira-api/internal/http/response"
    "pemira-api/internal/shared/constants"
    "pemira-api/internal/shared/ctxkeys"
)
```

## 🚀 Quick Setup

### 1. Initialize Dependencies

```go
// JWT Config
jwtConfig := auth.JWTConfig{
    Secret:           os.Getenv("JWT_SECRET"),
    AccessTokenTTL:   15 * time.Minute,
    RefreshTokenTTL:  7 * 24 * time.Hour,
}

// JWT Manager
jwtManager := auth.NewJWTManager(jwtConfig)

// Repository
authRepo := auth.NewPgRepository(dbPool)

// Service
authService := auth.NewAuthService(authRepo, jwtManager, jwtConfig)

// Handler
authHandler := auth.NewAuthHandler(authService)
```

### 2. Setup Routes

```go
func setupRoutes(r chi.Router, authHandler *auth.AuthHandler, jwtManager *auth.JWTManager) {
    // Public routes
    r.Post("/auth/login", authHandler.Login)
    r.Post("/auth/refresh", authHandler.RefreshToken)
    
    // Protected routes (require JWT)
    r.Group(func(r chi.Router) {
        r.Use(middleware.JWTAuth(jwtManager))
        r.Get("/auth/me", authHandler.Me)
        r.Post("/auth/logout", authHandler.Logout)
    })
    
    // Student only routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthStudentOnly(jwtManager))
        r.Post("/votes", voteHandler.Vote)
    })
    
    // Admin only routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthAdminOnly(jwtManager))
        r.Get("/admin/dashboard", adminHandler.Dashboard)
    })
    
    // TPS Operator only routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthTPSOperatorOnly(jwtManager))
        r.Post("/tps/votes", tpsHandler.RecordVote)
    })
}
```

## 🔐 API Endpoints

### POST /auth/login

**Request:**
```json
{
    "username": "student123",
    "password": "password123"
}
```

**Response 200:**
```json
{
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "random-base64-string",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
        "id": 1,
        "username": "student123",
        "role": "STUDENT",
        "voter_id": 123,
        "profile": {
            "name": "John Doe",
            "faculty_name": "Fakultas Teknik",
            "study_program_name": "Teknik Informatika",
            "cohort_year": 2020
        }
    }
}
```

### GET /auth/me

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response 200:**
```json
{
    "id": 1,
    "username": "student123",
    "role": "STUDENT",
    "voter_id": 123,
    "profile": {
        "name": "John Doe",
        "faculty_name": "Fakultas Teknik",
        "study_program_name": "Teknik Informatika",
        "cohort_year": 2020
    }
}
```

### POST /auth/refresh

**Request:**
```json
{
    "refresh_token": "random-base64-string"
}
```

**Response 200:**
```json
{
    "access_token": "new-jwt-token",
    "refresh_token": "new-refresh-token",
    "token_type": "Bearer",
    "expires_in": 900
}
```

### POST /auth/logout

**Request:**
```json
{
    "refresh_token": "random-base64-string"
}
```

**Response 200:**
```json
{
    "message": "Logged out successfully."
}
```

## 🔨 Usage in Handlers

### Get Current User from Context

```go
func (h *MyHandler) SomeProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
    // Get user ID (required)
    userID, ok := ctxkeys.GetUserID(r.Context())
    if !ok {
        response.Unauthorized(w, "UNAUTHORIZED", "User not authenticated")
        return
    }
    
    // Get user role (optional)
    role, ok := ctxkeys.GetUserRole(r.Context())
    if ok {
        if role == string(constants.RoleAdmin) {
            // Admin-specific logic
        }
    }
    
    // Get voter ID (optional, only for STUDENT role)
    voterID, ok := ctxkeys.GetVoterID(r.Context())
    if ok {
        // Use voter ID
    }
    
    // Use userID for business logic
    // ...
}
```

### Call Auth Service Directly

```go
// In your handler or service
func (h *MyHandler) CustomLogin(w http.ResponseWriter, r *http.Request) {
    var req auth.LoginRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    userAgent := r.Header.Get("User-Agent")
    ipAddress := r.RemoteAddr
    
    result, err := h.authService.Login(r.Context(), req, userAgent, ipAddress)
    if err != nil {
        // Handle error
        return
    }
    
    response.JSON(w, http.StatusOK, result)
}
```

### Get User Profile

```go
// Get current user with profile
authUser, err := authService.GetCurrentUser(ctx, userID)
if err != nil {
    // Handle error
}

// Access profile
if authUser.Profile != nil {
    name := authUser.Profile.Name
    faculty := authUser.Profile.FacultyName
}
```

## 📝 Response Helper Usage

```go
// Success response
response.JSON(w, http.StatusOK, data)

// Error responses
response.BadRequest(w, "VALIDATION_ERROR", "Invalid input")
response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid")
response.Forbidden(w, "FORBIDDEN", "Akses ditolak")
response.NotFound(w, "NOT_FOUND", "Resource tidak ditemukan")
response.UnprocessableEntity(w, "VALIDATION_ERROR", "Data tidak valid")
response.InternalServerError(w, "INTERNAL_ERROR", "Terjadi kesalahan")
response.Conflict(w, "CONFLICT", "Data sudah ada")
```

## 🔑 Working with Roles

```go
// Role constants
constants.RoleStudent      // "STUDENT"
constants.RoleAdmin        // "ADMIN"
constants.RoleTPSOperator  // "TPS_OPERATOR"
constants.RoleSuperAdmin   // "SUPER_ADMIN"

// Check role in handler
role, _ := ctxkeys.GetUserRole(r.Context())
if role == string(constants.RoleStudent) {
    // Student-specific logic
}

// Use middleware for role-based access
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthStudentOnly(jwtManager))
    // Only students can access these routes
})
```

## 🛡️ Security Best Practices

### 1. Environment Variables
```bash
# .env
JWT_SECRET=your-secret-key-min-32-chars
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h  # 7 days
```

### 2. Password Requirements
```go
// In your registration/creation handler
if len(password) < 8 {
    response.UnprocessableEntity(w, "VALIDATION_ERROR", "Password minimal 8 karakter")
    return
}

hashedPassword, err := auth.HashPassword(password)
if err != nil {
    response.InternalServerError(w, "INTERNAL_ERROR", "Gagal memproses password")
    return
}
```

### 3. Rate Limiting (recommended)
```go
// Apply rate limiting to login endpoint
r.With(middleware.RateLimit(5, time.Minute)).Post("/auth/login", authHandler.Login)
```

### 4. IP Extraction (behind proxy)
```go
func extractIP(r *http.Request) string {
    ip := r.Header.Get("X-Real-IP")
    if ip == "" {
        ip = r.Header.Get("X-Forwarded-For")
        if ip != "" {
            ip = strings.Split(ip, ",")[0]
        }
    }
    if ip == "" {
        ip = r.RemoteAddr
    }
    return strings.TrimSpace(ip)
}
```

## 🧪 Testing Examples

### Using curl

```bash
# Login
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"student123","password":"password123"}')

# Extract access token
ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')

# Get current user
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Using httpie

```bash
# Login
http POST :8080/auth/login username=student123 password=password123

# Get me
http GET :8080/auth/me "Authorization: Bearer <token>"
```

## 📊 Database Queries

### Create User Account
```sql
INSERT INTO user_accounts (username, password_hash, role, voter_id, is_active)
VALUES ('student123', '$2a$12$...', 'STUDENT', 123, true);
```

### Check Active Sessions
```sql
SELECT id, user_id, user_agent, ip_address, created_at, expires_at
FROM user_sessions
WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW();
```

### Cleanup Expired Sessions
```sql
DELETE FROM user_sessions
WHERE expires_at < NOW() OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '30 days');
```

## 🐛 Common Issues & Solutions

### Issue: "invalid token"
**Solution:** Pastikan token format: `Bearer <token>`, tidak ada spasi extra

### Issue: "token expired"
**Solution:** Gunakan refresh token untuk mendapatkan access token baru

### Issue: "user not found in context"
**Solution:** Pastikan middleware JWTAuth sudah diapply ke route

### Issue: "password hash too long"
**Solution:** Column password_hash harus TEXT, bukan VARCHAR(255)

### Issue: "cannot scan NULL into *int64"
**Solution:** Gunakan pointer untuk nullable fields (voter_id, tps_id)

## 🎯 Complete Example: Protected Endpoint

```go
package mypackage

import (
    "net/http"
    "pemira-api/internal/http/response"
    "pemira-api/internal/shared/ctxkeys"
)

type MyHandler struct {
    service *MyService
}

// Protected endpoint - only accessible with valid JWT
func (h *MyHandler) GetMyData(w http.ResponseWriter, r *http.Request) {
    // 1. Get authenticated user from context (set by JWTAuth middleware)
    userID, ok := ctxkeys.GetUserID(r.Context())
    if !ok {
        response.Unauthorized(w, "UNAUTHORIZED", "User not authenticated")
        return
    }
    
    // 2. Get user role if needed
    role, ok := ctxkeys.GetUserRole(r.Context())
    if !ok {
        response.Unauthorized(w, "UNAUTHORIZED", "User role not found")
        return
    }
    
    // 3. Use userID and role in your business logic
    data, err := h.service.GetUserData(r.Context(), userID, role)
    if err != nil {
        response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil data")
        return
    }
    
    // 4. Return success response
    response.JSON(w, http.StatusOK, data)
}

// Router setup
func SetupRoutes(r chi.Router, handler *MyHandler, jwtManager *auth.JWTManager) {
    r.Group(func(r chi.Router) {
        // Apply JWT middleware to all routes in this group
        r.Use(middleware.JWTAuth(jwtManager))
        
        // All routes here require authentication
        r.Get("/my/data", handler.GetMyData)
    })
}
```

## 📚 Additional Resources

- **Full Documentation**: `AUTH_IMPLEMENTATION.md`
- **JWT Library**: github.com/golang-jwt/jwt/v5
- **Password Hashing**: golang.org/x/crypto/bcrypt
- **Database Driver**: github.com/jackc/pgx/v5

---

**Catatan:** Semua endpoint di atas sudah diimplementasikan dan siap digunakan di `internal/auth/`.
