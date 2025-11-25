# Election Settings API - Complete Documentation

## 📋 Overview

Endpoint untuk memuat semua pengaturan pemilu (election settings) dalam satu request. Menggabungkan informasi umum, jadwal tahapan, mode settings, dan branding.

## 🔗 Endpoint

### Get All Election Settings

**GET** `/api/v1/admin/elections/{electionID}/settings`

Mendapatkan semua pengaturan pemilu dalam satu response.

#### Headers
```
Authorization: Bearer {token}
```

#### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `electionID` | integer | ID pemilu yang akan diambil settingsnya |

#### Success Response (200 OK)

```json
{
  "election": {
    "id": 1,
    "year": 2024,
    "slug": "PEMIRA-2024",
    "name": "Pemilihan Raya BEM 2024",
    "description": "Pemilihan Raya Badan Eksekutif Mahasiswa periode 2024-2025",
    "academic_year": "2023/2024",
    "status": "VOTING_CLOSED",
    "current_phase": "REGISTRATION",
    "online_enabled": true,
    "tps_enabled": true,
    "voting_window": {
      "start_at": "2025-12-15T00:00:00+07:00",
      "end_at": "2025-12-17T23:59:59+07:00"
    },
    "created_at": "2025-11-25T19:53:35.677183+07:00",
    "updated_at": "2025-11-25T20:26:42.281096+07:00"
  },
  "phases": {
    "election_id": 1,
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
  },
  "mode_settings": {
    "election_id": 1,
    "online_enabled": true,
    "tps_enabled": true,
    "online_settings": {},
    "tps_settings": {
      "require_checkin": true,
      "require_ballot_qr": true
    },
    "updated_at": "2025-11-25T20:26:42.281096+07:00"
  },
  "branding": {
    "primary_logo_id": null,
    "secondary_logo_id": null,
    "updated_at": "2025-11-25T20:20:34.627465+07:00"
  }
}
```

#### Error Responses

**400 Bad Request**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "electionID tidak valid."
}
```

**401 Unauthorized**
```json
{
  "code": "UNAUTHORIZED",
  "message": "Token tidak valid."
}
```

**403 Forbidden**
```json
{
  "code": "FORBIDDEN",
  "message": "Akses ditolak. Hanya admin yang dapat mengakses endpoint ini."
}
```

**404 Not Found**
```json
{
  "code": "ELECTION_NOT_FOUND",
  "message": "Pemilu tidak ditemukan."
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Gagal mengambil pengaturan pemilu."
}
```

## 📊 Response Structure

### Election Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | ID pemilu |
| `year` | integer | Tahun pemilu |
| `slug` | string | Slug unik pemilu |
| `name` | string | Nama pemilu |
| `description` | string | Deskripsi pemilu |
| `academic_year` | string | Tahun akademik |
| `status` | string | Status pemilu (`DRAFT`, `PUBLISHED`, `VOTING_OPEN`, `VOTING_CLOSED`, `ARCHIVED`) |
| `current_phase` | string | Phase saat ini |
| `online_enabled` | boolean | Apakah voting online diaktifkan |
| `tps_enabled` | boolean | Apakah voting TPS diaktifkan |
| `voting_window` | object | Window waktu voting |
| `created_at` | timestamp | Waktu dibuat |
| `updated_at` | timestamp | Waktu terakhir diupdate |

### Phases Object

| Field | Type | Description |
|-------|------|-------------|
| `election_id` | integer | ID pemilu |
| `phases` | array | Array berisi tahapan pemilu |

**Phase Item:**
- `key`: Kode tahapan (`REGISTRATION`, `VERIFICATION`, `CAMPAIGN`, `QUIET_PERIOD`, `VOTING`, `RECAP`)
- `label`: Label tahapan dalam bahasa Indonesia
- `start_at`: Waktu mulai tahapan
- `end_at`: Waktu selesai tahapan

### Mode Settings Object

| Field | Type | Description |
|-------|------|-------------|
| `election_id` | integer | ID pemilu |
| `online_enabled` | boolean | Status voting online |
| `tps_enabled` | boolean | Status voting TPS |
| `online_settings` | object | Pengaturan khusus online voting |
| `tps_settings` | object | Pengaturan khusus TPS voting |
| `updated_at` | timestamp | Waktu terakhir diupdate |

**TPS Settings:**
- `require_checkin`: Apakah check-in diperlukan
- `require_ballot_qr`: Apakah QR ballot diperlukan

### Branding Object

| Field | Type | Description |
|-------|------|-------------|
| `primary_logo_id` | string/null | ID file logo primer |
| `secondary_logo_id` | string/null | ID file logo sekunder |
| `updated_at` | timestamp | Waktu terakhir diupdate |

## 💻 Usage Examples

### cURL

```bash
# Login first
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}' | jq -r '.access_token')

# Get all settings
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/settings" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

### JavaScript (Fetch)

```javascript
const token = localStorage.getItem('access_token');

const response = await fetch('http://localhost:8080/api/v1/admin/elections/1/settings', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});

const settings = await response.json();
console.log('Election:', settings.election);
console.log('Phases:', settings.phases);
console.log('Mode Settings:', settings.mode_settings);
console.log('Branding:', settings.branding);
```

### TypeScript Interface

```typescript
interface ElectionSettings {
  election: {
    id: number;
    year: number;
    slug: string;
    name: string;
    description: string;
    academic_year: string;
    status: 'DRAFT' | 'PUBLISHED' | 'VOTING_OPEN' | 'VOTING_CLOSED' | 'ARCHIVED';
    current_phase: string;
    online_enabled: boolean;
    tps_enabled: boolean;
    voting_window: {
      start_at: string;
      end_at: string;
    };
    created_at: string;
    updated_at: string;
  };
  phases: {
    election_id: number;
    phases: Array<{
      key: string;
      label: string;
      start_at: string;
      end_at: string;
    }>;
  };
  mode_settings: {
    election_id: number;
    online_enabled: boolean;
    tps_enabled: boolean;
    online_settings: Record<string, any>;
    tps_settings: {
      require_checkin: boolean;
      require_ballot_qr: boolean;
    };
    updated_at: string;
  };
  branding: {
    primary_logo_id: string | null;
    secondary_logo_id: string | null;
    updated_at: string;
  };
}
```

## 🔄 Related Endpoints

Endpoint individual yang digabungkan dalam `/settings`:

- `GET /admin/elections/{electionID}` - Info umum pemilu
- `GET /admin/elections/{electionID}/phases` - Jadwal tahapan
- `GET /admin/elections/{electionID}/settings/mode` - Mode settings (online/TPS)
- `GET /admin/elections/{electionID}/branding` - Branding info

## ⚠️ Notes

1. **Authorization Required**: Endpoint ini memerlukan autentikasi dengan role `ADMIN`
2. **Single Request**: Mengambil semua data dalam satu request, efisien untuk loading settings page
3. **Real-time Data**: Data selalu up-to-date dari database
4. **Timezone**: Semua timestamp dalam format ISO 8601 dengan timezone WIB (UTC+7)
5. **Performance**: Response time < 100ms untuk single election

## 🎯 Use Cases

### Frontend Settings Page

Gunakan endpoint ini untuk:
- Load halaman pengaturan pemilu
- Display election configuration dashboard
- Initialize form edit settings
- Preview election info sebelum publish

### Admin Dashboard

Gunakan untuk:
- Quick overview pengaturan pemilu
- Monitoring phase saat ini
- Verifikasi konfigurasi sebelum voting dimulai

---

**Created**: 2025-11-25  
**Last Updated**: 2025-11-25  
**API Version**: v1
