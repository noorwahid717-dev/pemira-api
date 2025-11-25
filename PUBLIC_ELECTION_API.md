# Public Election Endpoints - Documentation

## 📋 Overview

Dokumentasi untuk **public endpoints** (tanpa autentikasi) untuk mendapatkan informasi pemilu aktif dan timeline.

---

## 🔗 Endpoints

### 1. Get Current Active Election

**GET** `/api/v1/elections/current`

Mengembalikan pemilu yang sedang aktif (status `VOTING_OPEN`).

#### Authentication
❌ **No authentication required** - Public endpoint

#### Response (200 OK)
```json
{
  "id": 2,
  "year": 2025,
  "name": "Pemilihan Raya BEM 2025",
  "slug": "PEMIRA-2025",
  "status": "VOTING_OPEN",
  "voting_start_at": "2025-11-25T12:00:00+07:00",
  "voting_end_at": "2025-12-01T23:59:59+07:00",
  "online_enabled": true,
  "tps_enabled": true
}
```

#### Error Responses

**404 Not Found** - Tidak ada pemilu aktif
```json
{
  "code": "ELECTION_NOT_FOUND",
  "message": "Tidak ada pemilu yang sedang berlangsung."
}
```

#### cURL Example
```bash
curl "http://localhost:8080/api/v1/elections/current"
```

---

### 2. List All Elections

**GET** `/api/v1/elections`

Mengembalikan daftar pemilu yang aktif/tersedia (excludes ARCHIVED).

#### Authentication
❌ **No authentication required** - Public endpoint

#### Response (200 OK)
```json
[
  {
    "id": 2,
    "year": 2025,
    "name": "Pemilihan Raya BEM 2025",
    "slug": "PEMIRA-2025",
    "status": "VOTING_OPEN",
    "voting_start_at": "2025-11-25T12:00:00+07:00",
    "voting_end_at": "2025-12-01T23:59:59+07:00",
    "online_enabled": true,
    "tps_enabled": true
  },
  {
    "id": 1,
    "year": 2024,
    "name": "Pemilihan Raya BEM 2024",
    "slug": "PEMIRA-2024",
    "status": "VOTING_CLOSED",
    "voting_start_at": "2024-12-15T00:00:00+07:00",
    "voting_end_at": "2024-12-17T23:59:59+07:00",
    "online_enabled": true,
    "tps_enabled": true
  }
]
```

#### cURL Example
```bash
curl "http://localhost:8080/api/v1/elections"
```

---

### 3. Get Election Timeline/Phases

**GET** `/api/v1/elections/{electionID}/phases`  
**GET** `/api/v1/elections/{electionID}/timeline` (alias)

Mengembalikan jadwal tahapan pemilu (timeline) untuk ditampilkan di landing page.

#### Authentication
❌ **No authentication required** - Public endpoint

#### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `electionID` | integer | ID pemilu yang ingin ditampilkan timeline-nya |

#### Response (200 OK)
```json
{
  "election_id": 2,
  "phases": [
    {
      "key": "REGISTRATION",
      "label": "Pendaftaran",
      "start_at": "2025-11-01T00:00:00+07:00",
      "end_at": "2025-11-20T23:59:59+07:00"
    },
    {
      "key": "VERIFICATION",
      "label": "Verifikasi Berkas",
      "start_at": "2025-11-21T00:00:00+07:00",
      "end_at": "2025-11-22T23:59:59+07:00"
    },
    {
      "key": "CAMPAIGN",
      "label": "Masa Kampanye",
      "start_at": "2025-11-23T00:00:00+07:00",
      "end_at": "2025-11-24T23:59:59+07:00"
    },
    {
      "key": "QUIET_PERIOD",
      "label": "Masa Tenang",
      "start_at": "2025-11-25T00:00:00+07:00",
      "end_at": "2025-11-25T12:00:00+07:00"
    },
    {
      "key": "VOTING",
      "label": "Voting",
      "start_at": "2025-11-25T12:00:00+07:00",
      "end_at": "2025-12-01T23:59:59+07:00"
    },
    {
      "key": "RECAP",
      "label": "Rekapitulasi",
      "start_at": "2025-12-02T00:00:00+07:00",
      "end_at": "2025-12-03T23:59:59+07:00"
    }
  ]
}
```

#### Error Responses

**404 Not Found** - Election tidak ditemukan
```json
{
  "code": "ELECTION_NOT_FOUND",
  "message": "Pemilu tidak ditemukan."
}
```

#### cURL Example
```bash
# Get phases
curl "http://localhost:8080/api/v1/elections/2/phases"

# Or use timeline alias
curl "http://localhost:8080/api/v1/elections/2/timeline"
```

---

## 💻 Frontend Integration

### React Example

```typescript
import { useState, useEffect } from 'react';

interface Election {
  id: number;
  year: number;
  name: string;
  slug: string;
  status: string;
  voting_start_at: string;
  voting_end_at: string;
  online_enabled: boolean;
  tps_enabled: boolean;
}

interface Phase {
  key: string;
  label: string;
  start_at: string;
  end_at: string;
}

function LandingPage() {
  const [election, setElection] = useState<Election | null>(null);
  const [phases, setPhases] = useState<Phase[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadElectionData() {
      try {
        // Try to get current active election
        let response = await fetch('http://localhost:8080/api/v1/elections/current');
        
        if (response.status === 404) {
          // Fallback: Get list and use first one
          const elections = await fetch('http://localhost:8080/api/v1/elections')
            .then(r => r.json());
          
          if (elections.length > 0) {
            setElection(elections[0]);
            
            // Get timeline for this election
            const timeline = await fetch(
              `http://localhost:8080/api/v1/elections/${elections[0].id}/phases`
            ).then(r => r.json());
            
            setPhases(timeline.phases);
          } else {
            setError('Belum ada pemilu aktif');
          }
        } else {
          const currentElection = await response.json();
          setElection(currentElection);
          
          // Get timeline
          const timeline = await fetch(
            `http://localhost:8080/api/v1/elections/${currentElection.id}/timeline`
          ).then(r => r.json());
          
          setPhases(timeline.phases);
        }
      } catch (err) {
        setError('Gagal memuat data pemilu');
        console.error(err);
      } finally {
        setLoading(false);
      }
    }

    loadElectionData();
  }, []);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  if (!election) return <div>Belum ada pemilu aktif</div>;

  return (
    <div>
      <h1>{election.name}</h1>
      <p>Status: {election.status}</p>
      
      {election.status === 'VOTING_OPEN' && (
        <div>
          <p>Voting berlangsung:</p>
          <p>{new Date(election.voting_start_at).toLocaleString()}</p>
          <p>sampai</p>
          <p>{new Date(election.voting_end_at).toLocaleString()}</p>
        </div>
      )}

      <h2>Timeline Pemilu</h2>
      <ul>
        {phases.map(phase => (
          <li key={phase.key}>
            <strong>{phase.label}</strong>
            <br />
            {new Date(phase.start_at).toLocaleDateString()} - 
            {new Date(phase.end_at).toLocaleDateString()}
          </li>
        ))}
      </ul>
    </div>
  );
}
```

### JavaScript Fetch
```javascript
// Get current election
const currentElection = await fetch('http://localhost:8080/api/v1/elections/current')
  .then(r => r.json())
  .catch(() => null);

if (!currentElection) {
  // Fallback to list
  const elections = await fetch('http://localhost:8080/api/v1/elections')
    .then(r => r.json());
  console.log('Available elections:', elections);
}

// Get timeline
const timeline = await fetch('http://localhost:8080/api/v1/elections/2/timeline')
  .then(r => r.json());
console.log('Phases:', timeline.phases);
```

---

## 🎯 Use Cases

### 1. Landing Page
```typescript
// Load current election info
GET /elections/current

// If 404, fallback to:
GET /elections  // Get list and use first one

// Then load timeline:
GET /elections/{id}/timeline
```

### 2. Election Selector
```typescript
// Load all available elections
GET /elections

// User selects one, then load details:
GET /elections/{selected_id}/timeline
GET /elections/{selected_id}/candidates
```

### 3. Quick Status Check
```typescript
// Just check if voting is open
const current = await fetch('/elections/current');
if (current.status === 200) {
  const data = await current.json();
  if (data.status === 'VOTING_OPEN') {
    // Show voting button
  }
}
```

---

## 📊 Response Fields

### Election Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | ID pemilu |
| `year` | integer | Tahun pemilu |
| `name` | string | Nama pemilu |
| `slug` | string | Slug unik |
| `status` | string | Status (`VOTING_OPEN`, `DRAFT`, etc) |
| `voting_start_at` | timestamp | Waktu mulai voting |
| `voting_end_at` | timestamp | Waktu selesai voting |
| `online_enabled` | boolean | Apakah voting online tersedia |
| `tps_enabled` | boolean | Apakah voting TPS tersedia |

### Phase Object

| Field | Type | Description |
|-------|------|-------------|
| `key` | string | Kode tahapan |
| `label` | string | Label tampilan |
| `start_at` | timestamp | Waktu mulai |
| `end_at` | timestamp | Waktu selesai |

### Phase Keys
- `REGISTRATION` - Pendaftaran
- `VERIFICATION` - Verifikasi Berkas
- `CAMPAIGN` - Masa Kampanye
- `QUIET_PERIOD` - Masa Tenang
- `VOTING` - Voting
- `RECAP` - Rekapitulasi

---

## ✅ Testing

```bash
# Test current election
curl "http://localhost:8080/api/v1/elections/current" | jq '.'

# Test list
curl "http://localhost:8080/api/v1/elections" | jq 'length'

# Test timeline
curl "http://localhost:8080/api/v1/elections/2/phases" | jq '.phases | length'
```

---

## 🚀 Status

✅ **All endpoints working and tested**  
✅ **No authentication required**  
✅ **Ready for frontend integration**  
✅ **Production ready**

---

**Created**: 2025-11-25  
**API Version**: v1  
**Status**: Production Ready ✅
