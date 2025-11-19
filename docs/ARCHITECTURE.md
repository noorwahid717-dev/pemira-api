# PEMIRA API Architecture

## 1. Gambaran Besar

```
[ Frontend (Web Mahasiswa, Admin, TPS Panel) ]
                ↓ (HTTPS REST/WebSocket)
[ Backend API (modular monolith) ]
   ├─ Auth & Session
   ├─ Election Management
   ├─ Voters & DPT
   ├─ Candidates & Campaign Content
   ├─ TPS & QR & Checkin
   ├─ Voting Engine (MISSION CRITICAL)
   ├─ Monitoring & Live Count (read-model)
   └─ Audit & Logs
                ↓
         [ PostgreSQL ]
                ↓
        [ (Optional) Redis ]
```

## 2. Module Boundaries

### Auth & Identity
- Login Mahasiswa (NIM + DOB / SSO)
- Login Admin & Panitia
- JWT/Session management
- Role: STUDENT, ADMIN, TPS_OPERATOR, SUPER_ADMIN

### Election Management
- CRUD pemilu
- Phase management (pendaftaran, kampanye, voting, rekap)
- Voting mode (Online / TPS / Hybrid)

### Voters & DPT
- DPT management
- Import/Export CSV/XLSX
- Voter status tracking

### Candidates & Campaign
- CRUD kandidat
- Media management (foto, poster, PDF, video)
- Public & admin endpoints

### TPS & QR & Check-in
- TPS management
- QR code generation/rotation
- Check-in workflow: scan → queue → approve/reject
- Real-time WebSocket updates

### Voting Engine (CRITICAL)
- Online voting
- TPS voting (post check-in approval)
- Transaction-based with row-level locks
- Prevents double voting
- Generates anonymous vote tokens

### Monitoring & Live Count
- Real-time vote aggregation
- Participation statistics
- WebSocket broadcasts
- Read-model optimization

### Audit & Logs
- All sensitive operations logged
- Super admin access only

## 3. Security Principles

1. **Voting Path**: Transaction + row-level lock (FOR UPDATE)
2. **Audit**: All sensitive actions logged
3. **JWT**: Short timeout + refresh token
4. **Rate Limiting**: Login & voting endpoints
5. **CSRF**: Prevention for cookie-based auth
6. **Double Vote Prevention**: Database constraints + transaction isolation

## 4. API Surface

### Mahasiswa Endpoints
```
POST   /auth/login/student
GET    /me
GET    /me/voter-status
GET    /elections/current
GET    /elections/:id/candidates
POST   /voting/online/cast
POST   /tps/checkin/scan
POST   /voting/tps/cast
```

### TPS Panel Endpoints
```
POST   /auth/login/tps
GET    /tps/:id/checkins?status=PENDING
POST   /tps/:id/checkins/:checkinId/approve
POST   /tps/:id/checkins/:checkinId/reject
WS     /ws/tps/:id
```

### Admin Endpoints
```
POST   /auth/login/admin
GET    /admin/dashboard/summary
GET    /admin/monitoring/live
GET    /admin/candidates
POST   /admin/candidates
GET    /admin/dpt
POST   /admin/dpt/import
GET    /admin/tps
POST   /admin/tps
PUT    /admin/elections/:id/config
```

## 5. Data Flow: Voting Process

### Online Voting
```
1. User requests POST /voting/online/cast
2. Middleware: Verify JWT, extract user_id
3. Service layer:
   - BEGIN TRANSACTION
   - SELECT * FROM voter_status WHERE voter_id = ? FOR UPDATE
   - Check: has_voted = false
   - Check: election phase = VOTING
   - Generate vote_token
   - INSERT INTO votes (candidate_id, token_hash, voted_via)
   - UPDATE voter_status SET has_voted = true
   - INSERT INTO audit_logs
   - COMMIT
4. Response: { success: true, token: "xxx" }
```

### TPS Voting
```
1. Mahasiswa scan QR → POST /tps/checkin/scan
2. Backend: Create tps_checkins (status: PENDING)
3. WebSocket: Broadcast to TPS panel
4. TPS Operator: Approve via POST /tps/:id/checkins/:checkinId/approve
5. Mahasiswa: POST /voting/tps/cast
6. Same transaction flow as online voting
```

## 6. Observability

### Metrics
- `votes_total` - Total votes cast
- `votes_per_minute` - Vote rate
- `voting_errors_total` - Error count
- `tps_queue_length` - Pending check-ins

### Logs
- Structured JSON logging
- Vote events (without candidate data)
- All audit trail

### Tracing (Optional)
- OpenTelemetry for distributed tracing
- Critical path: voting transaction

## 7. Non-Functional Requirements

### Performance
- Vote endpoint: < 500ms p95
- Live count: Update every 5-10s
- WebSocket: < 100ms latency

### Reliability
- Database: PITR backup
- Transaction isolation: READ COMMITTED
- Retry logic: Idempotent operations

### Scalability
- Horizontal: Multiple API instances
- Read replicas: For reporting
- Redis: Cache + pub/sub for WebSocket

## 8. Technology Stack

- **Language**: Go 1.22+
- **HTTP**: chi router
- **Database**: PostgreSQL 16
- **WebSocket**: nhooyr.io/websocket
- **Auth**: JWT (golang-jwt)
- **Logging**: slog
- **Metrics**: Prometheus
- **Cache**: Redis (optional)
