# 🎛️ Election Settings API - Backend

## 🎯 Konsep

**Admin bisa SET active election ID lewat API endpoint!**

Frontend READ dari settings, bukan hard-code!

**BONUS:** Admin juga bisa CREATE election baru lewat API!

---

## 📊 Database Schema

```sql
CREATE TABLE app_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW(),
    updated_by INT REFERENCES user_accounts(id)
);

-- Data default
INSERT INTO app_settings (key, value, description)
VALUES 
    ('active_election_id', '1', 'ID election aktif untuk admin dashboard'),
    ('default_election_id', '1', 'ID election default untuk voter');
```

---

## 📡 API Endpoints

### 1. GET `/api/v1/admin/settings` ✅

**Purpose:** Get all settings

**Response:**
```json
{
  "active_election_id": 1,
  "default_election_id": 1
}
```

---

### 2. GET `/api/v1/admin/settings/active-election` ✅

**Purpose:** Get active election ID only

**Response:**
```json
{
  "active_election_id": 1
}
```

---

### 3. PUT `/api/v1/admin/settings/active-election` ✅

**Purpose:** Update active election ID

**Request:**
```json
{
  "election_id": 3
}
```

**Response:**
```json
{
  "success": true,
  "message": "Active election berhasil diupdate",
  "active_election_id": 3
}
```

---

## 🧪 Testing

```bash
# Login as admin
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}' | jq -r '.access_token')

# Get current active election
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/admin/settings/active-election

# Update to election 2
curl -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"election_id": 2}' \
  http://localhost:8080/api/v1/admin/settings/active-election

# Verify
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/admin/settings/active-election
```

**Result:**
```json
{
  "active_election_id": 2
}
```

---

## 📝 Frontend Implementation

### 1. GET Active Election (on page load)

```typescript
const response = await api.get('/admin/settings/active-election');
const electionId = response.data.active_election_id;

// Use this ID for DPT, candidates, etc
loadDPT(electionId);
```

### 2. UPDATE Active Election (from settings page)

```typescript
const updateActiveElection = async (newElectionId: number) => {
  await api.put('/admin/settings/active-election', {
    election_id: newElectionId
  });
  
  // Reload data
  window.location.reload(); // or emit event
};
```

### 3. Settings Page UI

```tsx
<select 
  value={activeElectionId}
  onChange={(e) => updateActiveElection(Number(e.target.value))}
>
  <option value="1">BEM 2026 (41 DPT)</option>
  <option value="2">BEM 2025 (10 DPT)</option>
  <option value="3">Pemira Auto (1 DPT)</option>
</select>
```

---

## ✅ Benefits

**BEFORE:**
```typescript
const electionId = 3; // ❌ Hard-coded di frontend
```

**AFTER:**
```typescript
// ✅ Dynamic from backend settings
const { active_election_id } = await api.get('/admin/settings/active-election');
```

**Advantages:**
- ✅ Admin bisa ganti election dari UI
- ✅ Semua page otomatis pakai election yang benar
- ✅ Tidak perlu ubah code frontend
- ✅ Centralized configuration
- ✅ Audit trail (siapa yang update, kapan)

---

## 🎯 Use Case

**Scenario:**
1. Admin login
2. Buka halaman DPT
3. Frontend fetch: `GET /admin/settings/active-election`
4. Response: `{ "active_election_id": 1 }`
5. Frontend load DPT dari election 1 (41 voters)
6. Admin mau ganti ke election 2
7. Admin klik dropdown, pilih "BEM 2025"
8. Frontend: `PUT /admin/settings/active-election` dengan `{ "election_id": 2 }`
9. Page reload
10. Sekarang semua data dari election 2 (10 voters)

---

## 📂 Files Created

**Backend:**
- `migrations/20251126_add_app_settings.sql` - Database schema
- `internal/settings/model.go` - Data models
- `internal/settings/repository.go` - Database operations
- `internal/settings/service.go` - Business logic
- `internal/settings/http_handler.go` - HTTP handlers
- `cmd/api/main.go` - Routes registration

**Total:** 6 files, fully tested & working ✅

---

## 🚀 Status

✅ Backend API: **READY**  
✅ Database: **MIGRATED**  
✅ Endpoints: **TESTED**  
✅ Default value: **SET (election_id = 1)**  

🔄 Frontend: **NEEDS IMPLEMENTATION**

---

## 📞 Quick Reference

```bash
# Get active election
GET /admin/settings/active-election

# Update active election
PUT /admin/settings/active-election
Body: { "election_id": 2 }

# Get all settings
GET /admin/settings
```

**Auth Required:** Admin only

---

## 🆕 BONUS: CREATE NEW ELECTION

### POST `/api/v1/admin/elections` ✅

**Purpose:** Create election baru

**Request:**
```json
{
  "name": "PEMIRA BEM 2027",
  "slug": "pemira-2027",
  "year": 2027,
  "description": "Pemilihan BEM 2027"
}
```

**Response:**
```json
{
  "id": 4,
  "name": "PEMIRA BEM 2027",
  "code": "pemira-2027",
  "year": 2027,
  "status": "DRAFT",
  "description": "Pemilihan BEM 2027",
  "created_at": "2025-11-26T03:30:00Z"
}
```

**Testing:**
```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "PEMIRA BEM 2027",
    "slug": "pemira-2027",
    "year": 2027,
    "description": "Pemilihan BEM 2027"
  }' \
  http://localhost:8080/api/v1/admin/elections
```

**Frontend Implementation:**
```typescript
const createElection = async (data) => {
  const response = await api.post('/admin/elections', {
    name: data.name,
    slug: data.slug,
    year: data.year,
    description: data.description
  });
  
  const newElectionId = response.data.id;
  
  // Set as active election
  await api.put('/admin/settings/active-election', {
    election_id: newElectionId
  });
  
  // Reload page
  window.location.reload();
};
```

---

## 📋 Complete API List

**Election Management:**
- `GET /admin/elections` - List all elections
- `POST /admin/elections` - Create new election ✅
- `GET /admin/elections/{id}` - Get election detail
- `PUT /admin/elections/{id}` - Update election
- `DELETE /admin/elections/{id}` - Delete election (if needed)

**Settings:**
- `GET /admin/settings` - Get all settings
- `GET /admin/settings/active-election` - Get active election ID
- `PUT /admin/settings/active-election` - Update active election ID ✅

---

**SELESAI! API READY! Frontend tinggal integrate!** 🎉
