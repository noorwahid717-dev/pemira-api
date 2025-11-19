# Authentication System Implementation

Implementasi lengkap sistem autentikasi untuk Pemira API dengan fitur Login dan Me.

## 📁 Struktur File

```
internal/auth/
├── model.go              # Legacy model (deprecated)
├── entity.go             # Entity User & Session (core models)
├── dto_auth.go           # DTO untuk Login, Refresh, Me
├── repository.go         # Interface Repository
├── repository_pgx.go     # Implementasi Repository dengan pgx
├── service_auth.go       # Business Logic (Login, Refresh, GetMe)
├── handler_auth.go       # HTTP Handlers
├── jwt.go                # JWT Manager (Generate & Validate)
├── password.go           # Password Hashing & Token Generation
└── dto.go                # Legacy DTO (deprecated)

internal/http/
├── response/
│   └── response.go       # HTTP Response Helpers
└── middleware/
    ├── auth.go           # JWT Authentication Middleware
    ├── rbac.go           # Role-Based Access Control
    └── ...

internal/shared/
├── constants/
│   └── constants.go      # Role constants (STUDENT, ADMIN, TPS_OPERATOR)
└── ctxkeys/
    └── keys.go           # Context Keys untuk user info
```

## 1. Model & Entity

### Entity: UserAccount
**File:** `internal/auth/entity.go`

```go
type UserAccount struct {
    ID           int64          `json:"id"`
    Username     string         `json:"username"`
    PasswordHash string         `json:"-"`
    Role         constants.Role `json:"role"`
    VoterID      *int64         `json:"voter_id,omitempty"`
    TPSID        *int64         `json:"tps_id,omitempty"`
    IsActive     bool           `json:"is_active"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
}
```

### Entity: UserSession
```go
type UserSession struct {
    ID                int64      `json:"id"`
    UserID            int64      `json:"user_id"`
    RefreshTokenHash  string     `json:"-"`
    UserAgent         *string    `json:"user_agent,omitempty"`
    IPAddress         *string    `json:"ip_address,omitempty"`
    CreatedAt         time.Time  `json:"created_at"`
    ExpiresAt         time.Time  `json:"expires_at"`
    RevokedAt         *time.Time `json:"revoked_at,omitempty"`
}
```

### DTO: UserProfile
```go
type UserProfile struct {
    Name              string `json:"name,omitempty"`              // untuk STUDENT
    FacultyName       string `json:"faculty_name,omitempty"`      // untuk STUDENT
    StudyProgramName  string `json:"study_program_name,omitempty"` // untuk STUDENT
    CohortYear        *int   `json:"cohort_year,omitempty"`       // untuk STUDENT
    TPSCode           string `json:"tps_code,omitempty"`          // untuk TPS_OPERATOR
    TPSName           string `json:"tps_name,omitempty"`          // untuk TPS_OPERATOR
}
```

### DTO: AuthUser (Response User)
```go
type AuthUser struct {
    ID       int64          `json:"id"`
    Username string         `json:"username"`
    Role     constants.Role `json:"role"`
    VoterID  *int64         `json:"voter_id,omitempty"`
    TPSID    *int64         `json:"tps_id,omitempty"`
    Profile  *UserProfile   `json:"profile,omitempty"`
}
```

## 2. DTO Request & Response

### Login
**File:** `internal/auth/dto_auth.go`

```go
// Request
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

// Response
type LoginResponse struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    TokenType    string    `json:"token_type"`     // "Bearer"
    ExpiresIn    int64     `json:"expires_in"`     // seconds
    User         *AuthUser `json:"user"`
}
```

### Me (GET /auth/me)
```go
// Response langsung menggunakan *AuthUser
```

## 3. Repository

### Interface
**File:** `internal/auth/repository.go`

```go
type Repository interface {
    // User operations
    GetUserByUsername(ctx context.Context, username string) (*UserAccount, error)
    GetUserByID(ctx context.Context, userID int64) (*UserAccount, error)
    
    // Session operations
    CreateSession(ctx context.Context, session *UserSession) (*UserSession, error)
    GetSessionByTokenHash(ctx context.Context, tokenHash string) (*UserSession, error)
    RevokeSession(ctx context.Context, sessionID int64) error
    
    // Profile operations (custom method)
    GetUserProfile(ctx context.Context, user *UserAccount) (*UserProfile, error)
}
```

### Implementasi pgx
**File:** `internal/auth/repository_pgx.go`

#### GetUserByUsername
```go
func (r *PgRepository) GetUserByUsername(ctx context.Context, username string) (*UserAccount, error) {
    query := `
        SELECT id, username, password_hash, role, voter_id, tps_id, 
               is_active, created_at, updated_at
        FROM user_accounts
        WHERE username = $1
    `
    // ... scan & return
}
```

#### GetUserProfile
Mengambil profil berdasarkan role:
- **STUDENT**: Query ke tabel `voters` untuk nama, fakultas, prodi, angkatan
- **TPS_OPERATOR**: Query ke tabel `tps` untuk kode & nama TPS
- **ADMIN/SUPER_ADMIN**: Tidak ada profil tambahan

```go
func (r *PgRepository) GetUserProfile(ctx context.Context, user *UserAccount) (*UserProfile, error) {
    profile := &UserProfile{}
    
    switch user.Role {
    case constants.RoleStudent:
        if user.VoterID != nil {
            // Query voters table
        }
    case constants.RoleTPSOperator:
        if user.TPSID != nil {
            // Query tps table
        }
    }
    
    return profile, nil
}
```

#### CreateSession
```go
func (r *PgRepository) CreateSession(ctx context.Context, session *UserSession) (*UserSession, error) {
    query := `
        INSERT INTO user_sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, user_id, refresh_token_hash, user_agent, ip_address, 
                  created_at, expires_at, revoked_at
    `
    // ... insert & return
}
```

## 4. Service

### Configuration
**File:** `internal/auth/service_auth.go`

```go
type AuthService struct {
    repo       Repository
    jwtManager *JWTManager
    config     JWTConfig
}

type JWTConfig struct {
    Secret           string
    AccessTokenTTL   time.Duration  // e.g., 15 minutes
    RefreshTokenTTL  time.Duration  // e.g., 7 days
}
```

### Login Flow
```go
func (s *AuthService) Login(ctx context.Context, req LoginRequest, userAgent, ipAddress string) (*LoginResponse, error)
```

**Steps:**
1. Get user by username → `repo.GetUserByUsername()`
2. Check if user is active
3. Verify password → `bcrypt.CompareHashAndPassword()`
4. Generate access token → `jwtManager.GenerateAccessToken()`
5. Generate refresh token → `GenerateRandomToken()` + hash
6. Create session → `repo.CreateSession()`
7. Get user profile → `repo.GetUserProfile()`
8. Return `LoginResponse` with tokens + user data

**Error Handling:**
- User not found → `ErrInvalidCredentials` (blur error untuk security)
- User inactive → `ErrInactiveUser`
- Wrong password → `ErrInvalidCredentials`

### GetMe (GetCurrentUser)
```go
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*AuthUser, error)
```

**Steps:**
1. Get user by ID → `repo.GetUserByID()`
2. Get user profile → `repo.GetUserProfile()`
3. Return `AuthUser` with profile

## 5. HTTP Handlers

### Handler Setup
**File:** `internal/auth/handler_auth.go`

```go
type AuthHandler struct {
    service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler
```

### POST /auth/login
```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request)
```

**Request Body:**
```json
{
    "username": "student123",
    "password": "password"
}
```

**Response 200 OK:**
```json
{
    "access_token": "eyJhbGc...",
    "refresh_token": "random-token-base64",
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

**Error Responses:**
- `400 VALIDATION_ERROR`: Body tidak valid
- `422 VALIDATION_ERROR`: Username/password kosong
- `401 INVALID_CREDENTIALS`: Username atau password salah
- `403 USER_INACTIVE`: Akun tidak aktif
- `500 INTERNAL_ERROR`: Error sistem

### GET /auth/me
```go
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request)
```

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response 200 OK:**
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

**Error Responses:**
- `401 UNAUTHORIZED`: Token tidak valid atau tidak ditemukan
- `404 USER_NOT_FOUND`: Pengguna tidak ditemukan
- `500 INTERNAL_ERROR`: Error sistem

## 6. JWT & Token Management

### JWT Manager
**File:** `internal/auth/jwt.go`

```go
type JWTManager struct {
    config JWTConfig
}

func (j *JWTManager) GenerateAccessToken(user *UserAccount) (string, error)
func (j *JWTManager) ValidateAccessToken(tokenString string) (*JWTClaims, error)
```

### JWT Claims Structure
```go
type JWTClaims struct {
    UserID  int64          `json:"sub"`
    Role    constants.Role `json:"role"`
    VoterID *int64         `json:"voter_id,omitempty"`
    TPSID   *int64         `json:"tps_id,omitempty"`
    Exp     int64          `json:"exp"`
    Iat     int64          `json:"iat"`
}
```

### Access Token (JWT)
- Algorithm: HS256
- TTL: Configurable (recommended 15 minutes)
- Claims: user_id, role, voter_id, tps_id, exp, iat

### Refresh Token
- Format: Random base64-encoded string (32 bytes)
- Storage: Hashed dengan bcrypt di database
- TTL: Configurable (recommended 7 days)
- Associated with: User session (user_agent, ip_address)

## 7. Middleware

### JWT Authentication
**File:** `internal/http/middleware/auth.go`

```go
func JWTAuth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler
```

**Fungsi:**
1. Extract token dari header `Authorization: Bearer <token>`
2. Validate & parse JWT
3. Extract claims (user_id, role, voter_id, tps_id)
4. Store di context menggunakan `ctxkeys`

**Context Keys:**
```go
// internal/shared/ctxkeys/keys.go
const (
    UserIDKey     contextKey = "user_id"
    UserRoleKey   contextKey = "user_role"
    VoterIDKey    contextKey = "voter_id"
)

func GetUserID(ctx context.Context) (int64, bool)
func GetUserRole(ctx context.Context) (string, bool)
func GetVoterID(ctx context.Context) (int64, bool)
```

### Role-Based Middleware
```go
func AuthStudentOnly(jwtManager *auth.JWTManager) func(http.Handler) http.Handler
func AuthAdminOnly(jwtManager *auth.JWTManager) func(http.Handler) http.Handler
func AuthTPSOperatorOnly(jwtManager *auth.JWTManager) func(http.Handler) http.Handler
```

## 8. Password & Token Utilities

**File:** `internal/auth/password.go`

```go
// Password hashing (bcrypt cost 12)
func HashPassword(password string) (string, error)
func VerifyPassword(hashedPassword, password string) error

// Random token generation (crypto/rand)
func GenerateRandomToken(length int) (string, error)

// Refresh token hashing
func HashRefreshToken(token string) (string, error)
func VerifyRefreshToken(hashedToken, token string) error
```

## 9. Response Helpers

**File:** `internal/http/response/response.go`

```go
// Success response
func JSON(w http.ResponseWriter, statusCode int, data interface{})
func Success(w http.ResponseWriter, statusCode int, data interface{})

// Error responses
func Error(w http.ResponseWriter, statusCode int, code, message string, details interface{})
func BadRequest(w http.ResponseWriter, code, message string)
func Unauthorized(w http.ResponseWriter, code, message string)
func Forbidden(w http.ResponseWriter, code, message string)
func NotFound(w http.ResponseWriter, code, message string)
func UnprocessableEntity(w http.ResponseWriter, code, message string)
func InternalServerError(w http.ResponseWriter, code, message string)
func Conflict(w http.ResponseWriter, code, message string)
```

**Response Format:**
```json
// Success
{
    "data": { ... }
}

// Error
{
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": null
}
```

## 10. Database Schema

### Table: user_accounts
```sql
CREATE TABLE user_accounts (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) NOT NULL,  -- STUDENT, ADMIN, TPS_OPERATOR, SUPER_ADMIN
    voter_id BIGINT REFERENCES voters(id),
    tps_id BIGINT REFERENCES tps(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Table: user_sessions
```sql
CREATE TABLE user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES user_accounts(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);

CREATE INDEX idx_user_sessions_token ON user_sessions(refresh_token_hash);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
```

### Table: voters (for STUDENT profile)
```sql
CREATE TABLE voters (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    faculty_name VARCHAR(255),
    study_program_name VARCHAR(255),
    cohort_year INTEGER,
    -- ... other fields
);
```

### Table: tps (for TPS_OPERATOR profile)
```sql
CREATE TABLE tps (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    -- ... other fields
);
```

## 11. Routing Setup (Example)

```go
// cmd/api/main.go atau router setup
func setupAuthRoutes(r chi.Router, deps Dependencies) {
    authHandler := auth.NewAuthHandler(deps.AuthService)
    jwtMiddleware := middleware.JWTAuth(deps.JWTManager)
    
    // Public routes
    r.Post("/auth/login", authHandler.Login)
    r.Post("/auth/refresh", authHandler.RefreshToken)
    
    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(jwtMiddleware)
        r.Get("/auth/me", authHandler.Me)
        r.Post("/auth/logout", authHandler.Logout)
    })
}
```

## 12. Configuration Example

```go
// Config structure
type Config struct {
    JWT struct {
        Secret           string
        AccessTokenTTL   time.Duration
        RefreshTokenTTL  time.Duration
    }
}

// Initialization
jwtConfig := auth.JWTConfig{
    Secret:           os.Getenv("JWT_SECRET"),
    AccessTokenTTL:   15 * time.Minute,
    RefreshTokenTTL:  7 * 24 * time.Hour,
}

jwtManager := auth.NewJWTManager(jwtConfig)
authRepo := auth.NewPgRepository(db)
authService := auth.NewAuthService(authRepo, jwtManager, jwtConfig)
authHandler := auth.NewAuthHandler(authService)
```

## 13. Testing Examples

### Test Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "student123",
    "password": "password123"
  }'
```

### Test Me
```bash
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer eyJhbGc..."
```

## 14. Security Considerations

1. **Password Storage**: bcrypt dengan cost 12
2. **Refresh Token**: Hashed di database, tidak disimpan plain text
3. **Error Messages**: Blur error untuk username/password (tidak bedakan user tidak ada vs password salah)
4. **Token Expiry**: Access token pendek (15 min), refresh token lebih lama (7 days)
5. **Session Tracking**: Store user_agent & ip_address untuk audit
6. **Session Revocation**: Mendukung revoke individual session atau semua sessions user
7. **HTTPS Only**: Production harus pakai HTTPS untuk token security

## 15. Additional Features Implemented

### Refresh Token
- **Endpoint**: POST /auth/refresh
- **Purpose**: Get new access token using refresh token
- **Response**: New access token + new refresh token (rotation)

### Logout
- **Endpoint**: POST /auth/logout
- **Purpose**: Revoke refresh token
- **Effect**: User harus login ulang

### Session Management
- **List Sessions**: GET /auth/sessions (could be added)
- **Revoke Session**: DELETE /auth/sessions/:id (could be added)
- **Revoke All**: POST /auth/sessions/revoke-all (could be added)

### Session Cleanup
```go
// Periodic cleanup of expired sessions
func (r *PgRepository) CleanupExpiredSessions(ctx context.Context) error
```

## 16. Error Codes Reference

| Code | HTTP Status | Description |
|------|-------------|-------------|
| VALIDATION_ERROR | 400/422 | Input validation error |
| INVALID_CREDENTIALS | 401 | Username atau password salah |
| UNAUTHORIZED | 401 | Token tidak valid/tidak ada |
| TOKEN_EXPIRED | 401 | Token sudah kadaluarsa |
| INVALID_TOKEN | 401 | Token format tidak valid |
| INVALID_REFRESH_TOKEN | 401 | Refresh token tidak valid |
| USER_INACTIVE | 403 | Akun tidak aktif |
| FORBIDDEN | 403 | Akses ditolak (role tidak sesuai) |
| USER_NOT_FOUND | 404 | User tidak ditemukan |
| INTERNAL_ERROR | 500 | Error sistem internal |

## Summary

Sistem autentikasi sudah lengkap dengan:
✅ Model & Entity (UserAccount, UserSession, UserProfile)
✅ DTO (LoginRequest, LoginResponse, AuthUser)
✅ Repository Pattern dengan pgx (GetByUsername, GetUserProfile, CreateSession)
✅ Service Layer (Login, GetMe, Refresh, Logout)
✅ HTTP Handlers (POST /auth/login, GET /auth/me)
✅ JWT Management (Generate, Validate)
✅ Password Hashing (bcrypt)
✅ Refresh Token Management (generation, hashing, rotation)
✅ Middleware (JWT Auth, Role-based)
✅ Context Keys untuk user info
✅ Response Helpers dengan error codes
✅ Session Tracking (user_agent, ip_address)

Semua file sudah ada di `internal/auth/` dan siap digunakan!
