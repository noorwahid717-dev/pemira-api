# Election Settings API - Quick Reference

## 🚀 Endpoint Utama

### Get All Settings (Recommended)
```
GET /api/v1/admin/elections/{electionID}/settings
```
**Mengembalikan**: Election info + Phases + Mode Settings + Branding dalam 1 response

---

## 📦 Response Structure

```json
{
  "election": { /* General info */ },
  "phases": { /* Jadwal tahapan */ },
  "mode_settings": { /* Online/TPS settings */ },
  "branding": { /* Logo & branding */ }
}
```

---

## 🔐 Authentication

**Required**: Admin role

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}' | jq -r '.access_token')
```

---

## 💡 Quick Examples

### Get Settings for Election ID 1
```bash
curl -s "http://localhost:8080/api/v1/admin/elections/1/settings" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### Get Only Election Info
```bash
curl -s "http://localhost:8080/api/v1/admin/elections/1/settings" \
  -H "Authorization: Bearer $TOKEN" | jq '.election'
```

### Get Only Phases
```bash
curl -s "http://localhost:8080/api/v1/admin/elections/1/settings" \
  -H "Authorization: Bearer $TOKEN" | jq '.phases'
```

### Get Mode Settings
```bash
curl -s "http://localhost:8080/api/v1/admin/elections/1/settings" \
  -H "Authorization: Bearer $TOKEN" | jq '.mode_settings'
```

---

## 🎨 Frontend Integration

### React/Next.js Example
```typescript
async function loadElectionSettings(electionId: number) {
  const token = localStorage.getItem('access_token');
  
  const response = await fetch(
    `http://localhost:8080/api/v1/admin/elections/${electionId}/settings`,
    {
      headers: { 
        'Authorization': `Bearer ${token}` 
      }
    }
  );
  
  if (!response.ok) throw new Error('Failed to load settings');
  
  return await response.json();
}

// Usage
const settings = await loadElectionSettings(1);
console.log('Election:', settings.election.name);
console.log('Status:', settings.election.status);
console.log('Phases:', settings.phases.phases);
```

---

## 📋 Data Fields Quick Access

### Election Status Values
- `DRAFT` - Masih draft
- `PUBLISHED` - Sudah dipublish
- `VOTING_OPEN` - Voting sedang berlangsung
- `VOTING_CLOSED` - Voting sudah ditutup
- `ARCHIVED` - Sudah diarsipkan

### Phase Keys
- `REGISTRATION` - Pendaftaran
- `VERIFICATION` - Verifikasi Berkas
- `CAMPAIGN` - Masa Kampanye
- `QUIET_PERIOD` - Masa Tenang
- `VOTING` - Voting
- `RECAP` - Rekapitulasi

---

## 🔗 Related Endpoints

| Endpoint | Purpose | Method |
|----------|---------|--------|
| `/admin/elections` | List all elections | GET |
| `/admin/elections/{id}` | Single election info | GET |
| `/admin/elections/{id}/phases` | Update phases | PUT |
| `/admin/elections/{id}/settings/mode` | Update mode | PUT |
| `/admin/elections/{id}/branding/logo/{slot}` | Upload logo | POST |

---

## ⚡ Performance Tips

1. **Cache the response** - Settings tidak sering berubah
2. **Use SWR/React Query** - Automatic caching & revalidation
3. **Load on mount** - Fetch sekali saat component mount
4. **Refetch on update** - Invalidate cache setelah update settings

---

## 🐛 Common Errors

| Status | Code | Solution |
|--------|------|----------|
| 401 | `INVALID_TOKEN` | Login ulang, token expired |
| 403 | `FORBIDDEN` | User bukan admin |
| 404 | `ELECTION_NOT_FOUND` | Check election ID valid |

---

## 📝 TypeScript Types

```typescript
type ElectionStatus = 
  | 'DRAFT' 
  | 'PUBLISHED' 
  | 'VOTING_OPEN' 
  | 'VOTING_CLOSED' 
  | 'ARCHIVED';

type PhaseKey = 
  | 'REGISTRATION'
  | 'VERIFICATION'
  | 'CAMPAIGN'
  | 'QUIET_PERIOD'
  | 'VOTING'
  | 'RECAP';

interface Phase {
  key: PhaseKey;
  label: string;
  start_at: string;
  end_at: string;
}
```

---

## 🎯 Common Use Cases

### 1. Settings Page Load
```typescript
useEffect(() => {
  const loadSettings = async () => {
    const data = await loadElectionSettings(electionId);
    setElection(data.election);
    setPhases(data.phases.phases);
    setModeSettings(data.mode_settings);
    setBranding(data.branding);
  };
  
  loadSettings();
}, [electionId]);
```

### 2. Check Current Phase
```typescript
const currentPhase = settings.election.current_phase;
const isVoting = currentPhase === 'VOTING';
```

### 3. Validate Voting Window
```typescript
const now = new Date();
const startAt = new Date(settings.election.voting_window.start_at);
const endAt = new Date(settings.election.voting_window.end_at);
const canVote = now >= startAt && now <= endAt;
```

---

**📚 Full Documentation**: See `ELECTION_SETTINGS_API.md`
