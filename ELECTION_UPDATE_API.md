# Election Settings Update API - Complete Documentation

## 📋 Overview

Dokumentasi lengkap untuk **update semua pengaturan pemilu**. Terdapat beberapa endpoint untuk update berbagai aspek pengaturan.

---

## 🔗 Update Endpoints

### 1. Update General Election Info

**PUT** `/api/v1/admin/elections/{electionID}`

Update informasi umum pemilu (name, description, year, academic_year).

#### Request Body
```json
{
  "name": "Pemilihan Raya BEM 2024",
  "description": "Pemilihan Raya Badan Eksekutif Mahasiswa periode 2024-2025",
  "year": 2024,
  "slug": "PEMIRA-2024",
  "academic_year": "2023/2024"
}
```

#### Response (200 OK)
```json
{
  "id": 1,
  "year": 2024,
  "slug": "PEMIRA-2024",
  "name": "Pemilihan Raya BEM 2024",
  "description": "Pemilihan Raya Badan Eksekutif Mahasiswa periode 2024-2025",
  "academic_year": "2023/2024",
  "status": "DRAFT",
  "current_phase": "REGISTRATION",
  "online_enabled": true,
  "tps_enabled": true,
  "voting_window": {
    "start_at": "2025-12-15T00:00:00+07:00",
    "end_at": "2025-12-17T23:59:59+07:00"
  },
  "created_at": "2025-11-25T19:53:35.677183+07:00",
  "updated_at": "2025-11-25T21:00:00.000000+07:00"
}
```

#### cURL Example
```bash
curl -X PUT "http://localhost:8080/api/v1/admin/elections/1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pemilihan Raya BEM 2024",
    "description": "Pemilihan Raya BEM periode 2024-2025",
    "year": 2024,
    "slug": "PEMIRA-2024",
    "academic_year": "2023/2024"
  }'
```

---

### 2. Patch General Info (Partial Update)

**PATCH** `/api/v1/admin/elections/{electionID}`

Update sebagian informasi umum pemilu (hanya field yang dikirim).

#### Request Body
```json
{
  "name": "Pemilihan Raya BEM 2025 - Updated",
  "description": "Deskripsi baru"
}
```

#### Response (200 OK)
Same as PUT response

#### cURL Example
```bash
curl -X PATCH "http://localhost:8080/api/v1/admin/elections/1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pemilihan Raya BEM 2025 - Updated"
  }'
```

---

### 3. Update Election Phases

**PUT** `/api/v1/admin/elections/{electionID}/phases`

Update jadwal semua tahapan pemilu.

#### Request Body
```json
{
  "phases": [
    {
      "key": "REGISTRATION",
      "label": "Pendaftaran",
      "start_at": "2025-11-01T00:00:00+07:00",
      "end_at": "2025-11-30T23:59:59+07:00"
    },
    {
      "key": "VERIFICATION",
      "label": "Verifikasi Berkas",
      "start_at": "2025-12-01T00:00:00+07:00",
      "end_at": "2025-12-07T23:59:59+07:00"
    },
    {
      "key": "CAMPAIGN",
      "label": "Masa Kampanye",
      "start_at": "2025-12-08T00:00:00+07:00",
      "end_at": "2025-12-10T23:59:59+07:00"
    },
    {
      "key": "QUIET_PERIOD",
      "label": "Masa Tenang",
      "start_at": "2025-12-11T00:00:00+07:00",
      "end_at": "2025-12-14T23:59:59+07:00"
    },
    {
      "key": "VOTING",
      "label": "Voting",
      "start_at": "2025-12-15T00:00:00+07:00",
      "end_at": "2025-12-17T23:59:59+07:00"
    },
    {
      "key": "RECAP",
      "label": "Rekapitulasi",
      "start_at": "2025-12-21T00:00:00+07:00",
      "end_at": "2025-12-22T23:59:59+07:00"
    }
  ]
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "election_id": 1,
    "phases": [
      {
        "key": "REGISTRATION",
        "label": "Pendaftaran",
        "start_at": "2025-11-01T00:00:00+07:00",
        "end_at": "2025-11-30T23:59:59+07:00"
      },
      ...
    ]
  }
}
```

#### cURL Example
```bash
curl -X PUT "http://localhost:8080/api/v1/admin/elections/1/phases" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phases": [
      {
        "key": "REGISTRATION",
        "label": "Pendaftaran",
        "start_at": "2025-11-01T00:00:00+07:00",
        "end_at": "2025-11-30T23:59:59+07:00"
      }
    ]
  }'
```

#### Notes
- Semua 6 phases harus dikirim
- Phase keys yang valid: `REGISTRATION`, `VERIFICATION`, `CAMPAIGN`, `QUIET_PERIOD`, `VOTING`, `RECAP`
- Format waktu: ISO 8601 dengan timezone (e.g., `2025-11-01T00:00:00+07:00`)

---

### 4. Update Mode Settings

**PUT** `/api/v1/admin/elections/{electionID}/settings/mode`

Update pengaturan mode voting (Online/TPS).

#### Request Body
```json
{
  "online_enabled": true,
  "tps_enabled": true,
  "online_settings": {},
  "tps_settings": {
    "require_checkin": true,
    "require_ballot_qr": true
  }
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "election_id": 1,
    "online_enabled": true,
    "tps_enabled": true,
    "online_settings": {},
    "tps_settings": {
      "require_checkin": true,
      "require_ballot_qr": true
    },
    "updated_at": "2025-11-25T21:00:00+07:00"
  }
}
```

#### cURL Example
```bash
curl -X PUT "http://localhost:8080/api/v1/admin/elections/1/settings/mode" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "online_enabled": true,
    "tps_enabled": false
  }'
```

#### Notes
- `online_enabled`: Enable/disable online voting
- `tps_enabled`: Enable/disable TPS voting
- `tps_settings.require_checkin`: Require check-in sebelum vote di TPS
- `tps_settings.require_ballot_qr`: Require scan QR ballot di TPS

---

### 5. Upload Branding Logo

**POST** `/api/v1/admin/elections/{electionID}/branding/logo/{slot}`

Upload logo branding (primary atau secondary).

#### Path Parameters
- `electionID`: ID pemilu
- `slot`: `primary` atau `secondary`

#### Request
```
Content-Type: multipart/form-data

file: <binary data>
```

#### Response (200 OK)
```json
{
  "id": "uuid-logo-id",
  "content_type": "image/png",
  "size": 12345
}
```

#### cURL Example
```bash
curl -X POST "http://localhost:8080/api/v1/admin/elections/1/branding/logo/primary" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/logo.png"
```

#### Notes
- Max file size: 2 MB
- Supported formats: PNG, JPG, JPEG, SVG
- Slot `primary`: Logo utama
- Slot `secondary`: Logo sekunder

---

### 6. Delete Branding Logo

**DELETE** `/api/v1/admin/elections/{electionID}/branding/logo/{slot}`

Hapus logo branding.

#### Response (200 OK)
```json
{
  "primary_logo_id": null,
  "secondary_logo_id": "remaining-logo-id",
  "updated_at": "2025-11-25T21:00:00+07:00"
}
```

#### cURL Example
```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/elections/1/branding/logo/primary" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📊 Complete Update Flow Example

### Frontend Complete Update
```typescript
async function updateElectionSettings(electionId: number, settings: any) {
  const token = localStorage.getItem('access_token');
  
  // 1. Update general info
  await fetch(`/api/v1/admin/elections/${electionId}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      name: settings.name,
      description: settings.description,
      year: settings.year,
      slug: settings.slug,
      academic_year: settings.academicYear
    })
  });
  
  // 2. Update phases
  await fetch(`/api/v1/admin/elections/${electionId}/phases`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      phases: settings.phases
    })
  });
  
  // 3. Update mode settings
  await fetch(`/api/v1/admin/elections/${electionId}/settings/mode`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      online_enabled: settings.onlineEnabled,
      tps_enabled: settings.tpsEnabled,
      tps_settings: settings.tpsSettings
    })
  });
  
  // 4. Upload logo if changed
  if (settings.newPrimaryLogo) {
    const formData = new FormData();
    formData.append('file', settings.newPrimaryLogo);
    
    await fetch(`/api/v1/admin/elections/${electionId}/branding/logo/primary`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      },
      body: formData
    });
  }
}
```

---

## 🔒 Authorization

**Required**: Admin role

Semua endpoint memerlukan:
```
Authorization: Bearer {access_token}
```

---

## ⚠️ Error Responses

### 400 Bad Request
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Data tidak valid."
}
```

### 401 Unauthorized
```json
{
  "code": "UNAUTHORIZED",
  "message": "Token tidak valid."
}
```

### 403 Forbidden
```json
{
  "code": "FORBIDDEN",
  "message": "Akses ditolak. Hanya admin yang dapat mengakses."
}
```

### 404 Not Found
```json
{
  "code": "ELECTION_NOT_FOUND",
  "message": "Pemilu tidak ditemukan."
}
```

### 422 Unprocessable Entity
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Phase key tidak valid atau tidak lengkap."
}
```

---

## 📝 Validation Rules

### General Info
- `name`: Required, min 3 characters
- `slug`: Required, unique, alphanumeric + dash
- `year`: Required, positive integer
- `description`: Optional

### Phases
- All 6 phases must be provided
- Valid keys: `REGISTRATION`, `VERIFICATION`, `CAMPAIGN`, `QUIET_PERIOD`, `VOTING`, `RECAP`
- `start_at` must be before `end_at`
- Phases should not overlap (recommended but not enforced)

### Mode Settings
- `online_enabled`: boolean
- `tps_enabled`: boolean
- At least one mode must be enabled

### Branding
- File size: max 2 MB
- Formats: PNG, JPG, JPEG, SVG
- Slot: `primary` or `secondary`

---

## 🎯 Common Update Scenarios

### Scenario 1: Change Election Name Only
```bash
curl -X PATCH "http://localhost:8080/api/v1/admin/elections/1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "New Name"}'
```

### Scenario 2: Extend Voting Period
```bash
curl -X PUT "http://localhost:8080/api/v1/admin/elections/1/phases" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phases": [
      ...all phases with updated VOTING dates...
    ]
  }'
```

### Scenario 3: Disable TPS Voting
```bash
curl -X PUT "http://localhost:8080/api/v1/admin/elections/1/settings/mode" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "online_enabled": true,
    "tps_enabled": false
  }'
```

### Scenario 4: Change Primary Logo
```bash
# Delete old logo
curl -X DELETE "http://localhost:8080/api/v1/admin/elections/1/branding/logo/primary" \
  -H "Authorization: Bearer $TOKEN"

# Upload new logo
curl -X POST "http://localhost:8080/api/v1/admin/elections/1/branding/logo/primary" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@new-logo.png"
```

---

## 📚 Related Documentation

- **GET Settings**: `ELECTION_SETTINGS_API.md`
- **Quick Reference**: `ELECTION_SETTINGS_QUICK_REFERENCE.md`
- **Schedule Setup**: `ELECTION_SCHEDULE_SETUP.md`

---

**Created**: 2025-11-25  
**Last Updated**: 2025-11-25  
**API Version**: v1
