# TPS HTTP Handler Guide

## Overview

HTTP handlers untuk TPS (Tempat Pemungutan Suara) endpoints dengan chi router, mencakup student check-in dan panel approval.

## Endpoints

### Student Endpoints

#### 1. POST /tps/checkin/scan
Mahasiswa scan QR code di TPS untuk check-in.

#### 2. GET /tps/checkin/status
Cek status check-in mahasiswa.

### TPS Panel Endpoints (Operator)

#### 1. POST /tps/{tpsID}/checkins/{checkinID}/approve
Panitia TPS approve check-in mahasiswa.

#### 2. POST /tps/{tpsID}/checkins/{checkinID}/reject
Panitia TPS reject check-in mahasiswa.

#### 3. GET /tps/{tpsID}/checkins
List semua check-in di TPS (queue).

#### 4. GET /tps/{tpsID}/summary
Summary statistik TPS.

## Architecture Flow

```
Mahasiswa di TPS
    ↓ Scan QR Code
POST /tps/checkin/scan
    ↓
Handler: ScanQR
    ↓
Service: TPSService.ScanQR()
    ↓
- Parse & validate QR payload
- Check TPS status (ACTIVE)
- Check election phase (VOTING)
- Check voter eligibility
- Check already voted
- Create tps_checkins (STATUS: PENDING)
    ↓
Response: CheckinID, TPS info, Status PENDING
    ↓
Mahasiswa wait for approval

Panel TPS (via WebSocket or Polling)
    ↓ See pending check-ins
Panitia verify identity
    ↓ Approve
POST /tps/{tpsID}/checkins/{checkinID}/approve
    ↓
Handler: ApproveCheckin
    ↓
Service: TPSService.ApproveCheckin()
    ↓
- Check operator access to TPS
- Check checkin status (must be PENDING)
- Update tps_checkins (STATUS: APPROVED)
- Set expires_at (now + 15 minutes)
    ↓
Response: Approved, ExpiresAt
    ↓
Mahasiswa can now vote via TPS

Vote Flow
    ↓
POST /voting/tps/cast
    ↓
VotingService.CastTPSVote()
    ↓
- Get latest APPROVED check-in
- Validate not expired
- Cast vote
- Mark check-in as USED
```

## Request & Response Formats

### 1. POST /tps/checkin/scan

**Request:**
```json
{
  "qr_payload": "TPS001:abc123def456"
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "checkin_id": 123,
    "tps": {
      "id": 1,
      "code": "TPS001",
      "name": "TPS Gedung A"
    },
    "status": "PENDING",
    "scan_at": "2024-11-20T10:15:30Z",
    "message": "Check-in berhasil. Silakan menunggu verifikasi panitia TPS."
  }
}
```

**Error Responses:**

| Status | Code | Message |
|--------|------|---------|
| 400 | QR_INVALID | Kode QR tidak valid. |
| 400 | QR_REVOKED | Kode QR ini sudah tidak berlaku. |
| 400 | TPS_INACTIVE | TPS belum atau tidak aktif. |
| 400 | TPS_CLOSED | TPS sudah ditutup. |
| 400 | ELECTION_NOT_OPEN | Fase voting belum dibuka atau sudah ditutup. |
| 403 | NOT_ELIGIBLE | Anda tidak berhak memilih untuk pemilu ini. |
| 404 | TPS_NOT_FOUND | TPS tidak ditemukan. |
| 409 | ALREADY_VOTED | Anda sudah menggunakan hak suara. |

### 2. POST /tps/{tpsID}/checkins/{checkinID}/approve

**Request:**
```
POST /tps/1/checkins/123/approve
Authorization: Bearer <jwt_token>
```

No request body required.

**Success Response (200 OK):**
```json
{
  "data": {
    "checkin_id": 123,
    "status": "APPROVED",
    "approved_at": "2024-11-20T10:16:00Z",
    "expires_at": "2024-11-20T10:31:00Z",
    "voter": {
      "id": 12345,
      "nim": "1234567890",
      "name": "John Doe"
    },
    "tps": {
      "id": 1,
      "code": "TPS001",
      "name": "TPS Gedung A"
    }
  }
}
```

**Error Responses:**

| Status | Code | Message |
|--------|------|---------|
| 400 | CHECKIN_NOT_PENDING | Check-in bukan dalam status menunggu (PENDING). |
| 400 | CHECKIN_EXPIRED | Waktu check-in sudah kadaluarsa, silakan scan ulang di TPS. |
| 403 | TPS_ACCESS_DENIED | Anda tidak memiliki akses ke TPS ini. |
| 404 | CHECKIN_NOT_FOUND | Data check-in tidak ditemukan. |
| 404 | TPS_NOT_FOUND | TPS tidak ditemukan. |

### 3. POST /tps/{tpsID}/checkins/{checkinID}/reject

**Request:**
```json
{
  "reason": "KTP tidak sesuai dengan data"
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "checkin_id": 123,
    "status": "REJECTED",
    "reason": "KTP tidak sesuai dengan data"
  }
}
```

## Handler Implementation

### File Structure

```
internal/tps/
├── http_handler.go              # Main handler & routing
├── http_handler_checkin_v2.go   # Check-in handlers (new)
├── service.go                   # Business logic
├── repository.go                # Repository interfaces
├── dto.go                       # DTOs
├── errors.go                    # Domain errors
└── entity.go                    # Entity models
```

### Handler Code

```go
type TPSHandler struct {
    svc *TPSService
}

func NewTPSHandler(svc *TPSService) *TPSHandler {
    return &TPSHandler{svc: svc}
}

// MountPublic registers public routes (for students)
func (h *TPSHandler) MountPublic(r chi.Router) {
    r.Post("/tps/checkin/scan", h.ScanQR)
    r.Get("/tps/checkin/status", h.StudentCheckinStatus)
}

// MountPanel registers panel routes (for TPS operators)
func (h *TPSHandler) MountPanel(r chi.Router) {
    r.Post("/tps/{tpsID}/checkins/{checkinID}/approve", h.ApproveCheckin)
    r.Post("/tps/{tpsID}/checkins/{checkinID}/reject", h.PanelRejectCheckin)
    r.Get("/tps/{tpsID}/checkins", h.PanelListCheckins)
}
```

### Context Usage

```go
// For students (voterID)
voterID, ok := ctxkeys.GetVoterID(ctx)
if !ok {
    response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", 
        "Token tidak valid atau tidak memiliki akses.", nil)
    return
}

// For TPS operators (userID)
operatorUserID, ok := ctxkeys.GetUserID(ctx)
if !ok {
    response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", 
        "Token tidak valid atau tidak memiliki akses.", nil)
    return
}
```

### Request Validation

```go
// Parse body
var req ScanQRRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", 
        "Format body tidak valid.", nil)
    return
}

// Validate struct
if err := h.validate.Struct(req); err != nil {
    response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", 
        "qr_payload wajib diisi.", map[string]string{
            "field":      "qr_payload",
            "constraint": "required",
        })
    return
}
```

### Path Parameters

```go
tpsIDStr := chi.URLParam(r, "tpsID")
checkinIDStr := chi.URLParam(r, "checkinID")

tpsID, err := strconv.ParseInt(tpsIDStr, 10, 64)
if err != nil || tpsID <= 0 {
    response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", 
        "tps_id tidak valid.", nil)
    return
}
```

## Router Setup

### Option 1: Separate Groups by Role

```go
func NewRouter(deps Dependencies) http.Handler {
    r := chi.NewRouter()
    
    // Student routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthRequired)
        r.Use(middleware.StudentOnly)
        
        tpsHandler := tps.NewTPSHandler(deps.TPSService)
        tpsHandler.MountPublic(r)
    })
    
    // TPS Operator routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthRequired)
        r.Use(middleware.TPSOperatorOnly)
        
        tpsHandler := tps.NewTPSHandler(deps.TPSService)
        tpsHandler.MountPanel(r)
    })
    
    return r
}
```

### Option 2: Nested Routes

```go
r.Route("/tps", func(r chi.Router) {
    // Public endpoints (students)
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthStudentOnly)
        r.Post("/checkin/scan", tpsHandler.ScanQR)
        r.Get("/checkin/status", tpsHandler.StudentCheckinStatus)
    })
    
    // Panel endpoints (operators)
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthTPSOperatorOnly)
        r.Post("/{tpsID}/checkins/{checkinID}/approve", tpsHandler.ApproveCheckin)
        r.Post("/{tpsID}/checkins/{checkinID}/reject", tpsHandler.PanelRejectCheckin)
    })
})
```

## Middleware Requirements

### TPSOperatorOnly Middleware

```go
func TPSOperatorOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        role, ok := ctxkeys.GetUserRole(r.Context())
        if !ok {
            response.Forbidden(w, "Role tidak ditemukan")
            return
        }
        
        if role != string(constants.RoleTPSOperator) {
            response.Forbidden(w, "Akses ditolak. Endpoint ini hanya untuk operator TPS.")
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## Service Layer

### CheckinScan (example structure)

```go
func (s *TPSService) ScanQR(ctx context.Context, voterID int64, req *ScanQRRequest) (*CheckinResult, error) {
    // 1. Parse QR payload
    tpsID, qrSecret, err := parseQRPayload(req.QRPayload)
    if err != nil {
        return nil, ErrQRInvalid
    }
    
    // 2. Get TPS & validate
    tps, err := s.repo.GetByID(ctx, tpsID)
    if err != nil {
        return nil, ErrTPSNotFound
    }
    
    if tps.Status != StatusActive {
        return nil, ErrTPSInactive
    }
    
    // 3. Validate QR secret
    qr, err := s.repo.GetActiveQR(ctx, tpsID)
    if err != nil || qr.QRSecretSuffix != qrSecret {
        return nil, ErrQRInvalid
    }
    
    // 4. Check election phase
    // 5. Check voter eligibility
    // 6. Check already voted
    
    // 7. Create check-in
    checkin := &TPSCheckin{
        TPSID:      tpsID,
        VoterID:    voterID,
        ElectionID: tps.ElectionID,
        Status:     CheckinStatusPending,
        ScanAt:     time.Now().UTC(),
    }
    
    if err := s.repo.CreateCheckin(ctx, checkin); err != nil {
        return nil, err
    }
    
    return &CheckinResult{
        CheckinID: checkin.ID,
        TPS: CheckinTPSInfo{
            ID:   tps.ID,
            Code: tps.Code,
            Name: tps.Name,
        },
        Status: string(CheckinStatusPending),
        ScanAt: checkin.ScanAt,
    }, nil
}
```

### ApproveCheckin (example structure)

```go
func (s *TPSService) ApproveCheckin(ctx context.Context, tpsID, checkinID, operatorUserID int64) (*ApproveResult, error) {
    // 1. Check operator access to TPS
    hasAccess, err := s.repo.CheckPanitiaAccess(ctx, operatorUserID, tpsID)
    if err != nil || !hasAccess {
        return nil, ErrTPSAccessDenied
    }
    
    // 2. Get check-in
    checkin, err := s.repo.GetCheckinByID(ctx, checkinID)
    if err != nil {
        return nil, ErrCheckinNotFound
    }
    
    if checkin.Status != CheckinStatusPending {
        return nil, ErrCheckinNotPending
    }
    
    // 3. Get voter info
    voter, err := s.repo.GetVoterByID(ctx, checkin.VoterID)
    if err != nil {
        return nil, err
    }
    
    // 4. Get TPS info
    tps, err := s.repo.GetByID(ctx, tpsID)
    if err != nil {
        return nil, ErrTPSNotFound
    }
    
    // 5. Update check-in to APPROVED with TTL
    now := time.Now().UTC()
    expiresAt := now.Add(15 * time.Minute) // 15 min TTL
    
    checkin.Status = CheckinStatusApproved
    checkin.ApprovedAt = &now
    checkin.ApprovedByID = &operatorUserID
    checkin.ExpiresAt = &expiresAt
    
    if err := s.repo.UpdateCheckin(ctx, checkin); err != nil {
        return nil, err
    }
    
    return &ApproveResult{
        CheckinID:  checkin.ID,
        Status:     string(CheckinStatusApproved),
        ApprovedAt: now,
        ExpiresAt:  expiresAt,
        Voter: ApproveVoterInfo{
            ID:   voter.ID,
            NIM:  voter.NIM,
            Name: voter.Name,
        },
        TPS: ApproveTPSInfo{
            ID:   tps.ID,
            Code: tps.Code,
            Name: tps.Name,
        },
    }, nil
}
```

## Testing Examples

### Unit Test: ScanQR Handler

```go
func TestScanQR_Success(t *testing.T) {
    mockService := &MockTPSService{}
    handler := tps.NewTPSHandler(mockService)
    
    mockService.On("ScanQR", mock.Anything, int64(123), mock.Anything).
        Return(&tps.CheckinResult{
            CheckinID: 456,
            TPS: tps.CheckinTPSInfo{
                ID:   1,
                Code: "TPS001",
                Name: "TPS Gedung A",
            },
            Status: "PENDING",
            ScanAt: time.Now(),
        }, nil)
    
    body := `{"qr_payload":"TPS001:abc123"}`
    req := httptest.NewRequest("POST", "/tps/checkin/scan", strings.NewReader(body))
    ctx := context.WithValue(req.Context(), ctxkeys.VoterIDKey, int64(123))
    
    w := httptest.NewRecorder()
    handler.ScanQR(w, req.WithContext(ctx))
    
    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
}
```

### Integration Test

```go
func TestApproveCheckin_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create test data
    tpsID := setupTestTPS(t, db)
    voterID := setupTestVoter(t, db)
    operatorID := setupTestOperator(t, db, tpsID)
    checkinID := createPendingCheckin(t, db, tpsID, voterID)
    
    // Initialize service
    service := tps.NewService(tps.NewPostgresRepository(db))
    handler := tps.NewTPSHandler(service)
    
    // Execute request
    url := fmt.Sprintf("/tps/%d/checkins/%d/approve", tpsID, checkinID)
    req := httptest.NewRequest("POST", url, nil)
    ctx := context.WithValue(req.Context(), ctxkeys.UserIDKey, operatorID)
    
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("tpsID", strconv.FormatInt(tpsID, 10))
    rctx.URLParams.Add("checkinID", strconv.FormatInt(checkinID, 10))
    req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
    
    w := httptest.NewRecorder()
    handler.ApproveCheckin(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    
    var resp struct {
        Data struct {
            Status string `json:"status"`
        } `json:"data"`
    }
    json.NewDecoder(w.Body).Decode(&resp)
    assert.Equal(t, "APPROVED", resp.Data.Status)
}
```

## cURL Examples

### Scan QR Code

```bash
curl -X POST http://localhost:8080/tps/checkin/scan \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{"qr_payload":"TPS001:abc123def456"}' \
  | jq
```

### Approve Check-in

```bash
curl -X POST http://localhost:8080/tps/1/checkins/123/approve \
  -H "Authorization: Bearer eyJhbGc..." \
  | jq
```

### Get Check-in Queue

```bash
curl -X GET http://localhost:8080/tps/1/checkins?status=PENDING \
  -H "Authorization: Bearer eyJhbGc..." \
  | jq
```

## Common Issues & Solutions

### Issue: "QR_INVALID" response

**Cause:** QR payload format tidak valid atau tidak match

**Solution:**
1. Verify QR payload format: `TPS{code}:{secret}`
2. Check QR is not revoked
3. Regenerate QR if needed

### Issue: "TPS_ACCESS_DENIED" response

**Cause:** Operator tidak di-assign ke TPS tersebut

**Solution:**
1. Check `tps_panitia` table
2. Verify operator assigned to correct TPS
3. Admin must assign operator to TPS

### Issue: "CHECKIN_EXPIRED" on approve

**Cause:** Check-in TTL exceeded (> 15 minutes from scan)

**Solution:** Mahasiswa must scan QR again

### Issue: "CHECKIN_NOT_PENDING" on approve

**Cause:** Check-in already processed (APPROVED/REJECTED/USED)

**Solution:** Cannot re-approve, mahasiswa must scan again if needed

## Security Considerations

1. **QR Code Security:** Use unpredictable suffixes, rotate regularly
2. **TTL Enforcement:** Strict 15-minute check-in expiry
3. **Access Control:** Verify operator assigned to TPS
4. **Audit Trail:** Log all scan and approve actions
5. **Rate Limiting:** Prevent QR scanning abuse

## Performance Tips

1. **Index Optimization:** Index on `(tps_id, status, scan_at)`
2. **WebSocket for Real-time:** Use WS for panel updates instead of polling
3. **Cache TPS Data:** Cache TPS info to reduce DB queries
4. **Batch Operations:** Approve multiple check-ins in one request
5. **Database Pooling:** Proper connection pool configuration

## Next Steps

1. ✅ Implement handlers (DONE)
2. ✅ Add error mapping (DONE)
3. TODO: Wire to router in main.go
4. TODO: Implement WebSocket for real-time updates
5. TODO: Add integration tests
6. TODO: Add monitoring & metrics
7. TODO: Load testing
