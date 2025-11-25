# Election Settings CRUD API - Complete Reference

## 📋 Overview

Dokumentasi lengkap **CREATE, READ, UPDATE, DELETE** untuk semua pengaturan pemilu.

---

## 🔍 READ Operations (GET)

### 1. Get All Election Settings (Recommended)
```
GET /api/v1/admin/elections/{electionID}/settings
```
**Returns**: Election + Phases + Mode Settings + Branding dalam 1 response

### 2. Get Election Info Only
```
GET /api/v1/admin/elections/{electionID}
```

### 3. Get Phases Only
```
GET /api/v1/admin/elections/{electionID}/phases
```

### 4. Get Mode Settings Only
```
GET /api/v1/admin/elections/{electionID}/settings/mode
```

### 5. Get Branding Only
```
GET /api/v1/admin/elections/{electionID}/branding
```

### 6. Get Logo File
```
GET /api/v1/admin/elections/{electionID}/branding/logo/{slot}
```
Slot: `primary` atau `secondary`

### 7. List All Elections
```
GET /api/v1/admin/elections
```
Query params: `page`, `limit`, `status`, `year`, `search`

---

## ✏️ UPDATE Operations (PUT/PATCH)

### 1. Update General Info (Full)
```
PUT /api/v1/admin/elections/{electionID}
```
**Body**:
```json
{
  "name": "Election Name",
  "description": "Description",
  "year": 2024,
  "slug": "election-2024",
  "academic_year": "2023/2024"
}
```

### 2. Update General Info (Partial)
```
PATCH /api/v1/admin/elections/{electionID}
```
**Body**: Hanya field yang ingin diupdate
```json
{
  "name": "New Name"
}
```

### 3. Update Phases
```
PUT /api/v1/admin/elections/{electionID}/phases
```
**Body**:
```json
{
  "phases": [
    {
      "key": "REGISTRATION",
      "label": "Pendaftaran",
      "start_at": "2025-11-01T00:00:00+07:00",
      "end_at": "2025-11-30T23:59:59+07:00"
    },
    ...all 6 phases
  ]
}
```

### 4. Update Mode Settings
```
PUT /api/v1/admin/elections/{electionID}/settings/mode
```
**Body**:
```json
{
  "online_enabled": true,
  "tps_enabled": true,
  "tps_settings": {
    "require_checkin": true,
    "require_ballot_qr": true
  }
}
```

---

## ➕ CREATE Operations (POST)

### 1. Create New Election
```
POST /api/v1/admin/elections
```
**Body**:
```json
{
  "name": "Pemilihan Raya BEM 2025",
  "description": "Description",
  "year": 2025,
  "slug": "PEMIRA-2025",
  "academic_year": "2024/2025"
}
```

### 2. Upload Branding Logo
```
POST /api/v1/admin/elections/{electionID}/branding/logo/{slot}
```
**Content-Type**: `multipart/form-data`
**Body**: `file=<binary>`

Slot: `primary` atau `secondary`

---

## ❌ DELETE Operations (DELETE)

### 1. Delete Branding Logo
```
DELETE /api/v1/admin/elections/{electionID}/branding/logo/{slot}
```
Slot: `primary` atau `secondary`

---

## 🔄 Election Status Management

### Open Voting
```
POST /api/v1/admin/elections/{electionID}/actions/open-voting
```

### Close Voting
```
POST /api/v1/admin/elections/{electionID}/actions/close-voting
```

### Archive Election
```
POST /api/v1/admin/elections/{electionID}/actions/archive
```

---

## 📊 Complete CRUD Flow

### 1. CREATE New Election
```bash
# Step 1: Create election
curl -X POST "http://localhost:8080/api/v1/admin/elections" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pemilihan Raya BEM 2025",
    "description": "Election description",
    "year": 2025,
    "slug": "PEMIRA-2025",
    "academic_year": "2024/2025"
  }'

# Step 2: Set phases
curl -X PUT "http://localhost:8080/api/v1/admin/elections/3/phases" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "phases": [...] }'

# Step 3: Configure mode
curl -X PUT "http://localhost:8080/api/v1/admin/elections/3/settings/mode" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "online_enabled": true,
    "tps_enabled": true
  }'

# Step 4: Upload logo
curl -X POST "http://localhost:8080/api/v1/admin/elections/3/branding/logo/primary" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@logo.png"
```

### 2. READ Election Settings
```bash
# Get all settings in one request
curl "http://localhost:8080/api/v1/admin/elections/3/settings" \
  -H "Authorization: Bearer $TOKEN"
```

### 3. UPDATE Election Settings
```bash
# Update name only
curl -X PATCH "http://localhost:8080/api/v1/admin/elections/3" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Name"}'

# Update phases
curl -X PUT "http://localhost:8080/api/v1/admin/elections/3/phases" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"phases": [...]}'

# Update mode
curl -X PUT "http://localhost:8080/api/v1/admin/elections/3/settings/mode" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"online_enabled": false, "tps_enabled": true}'
```

### 4. DELETE Operations
```bash
# Delete logo
curl -X DELETE "http://localhost:8080/api/v1/admin/elections/3/branding/logo/primary" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🎯 Frontend Complete Example

### React Hook for Election CRUD
```typescript
import { useState, useEffect } from 'react';

function useElectionCRUD(electionId?: number) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const token = localStorage.getItem('access_token');
  const baseUrl = 'http://localhost:8080/api/v1/admin/elections';
  
  // CREATE
  async function createElection(data: ElectionCreateData) {
    setLoading(true);
    try {
      const response = await fetch(baseUrl, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      });
      
      if (!response.ok) throw new Error('Failed to create election');
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }
  
  // READ
  async function getSettings() {
    if (!electionId) return;
    
    setLoading(true);
    try {
      const response = await fetch(`${baseUrl}/${electionId}/settings`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (!response.ok) throw new Error('Failed to load settings');
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }
  
  // UPDATE
  async function updateElection(data: Partial<ElectionData>) {
    if (!electionId) return;
    
    setLoading(true);
    try {
      const response = await fetch(`${baseUrl}/${electionId}`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      });
      
      if (!response.ok) throw new Error('Failed to update election');
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }
  
  async function updatePhases(phases: Phase[]) {
    if (!electionId) return;
    
    setLoading(true);
    try {
      const response = await fetch(`${baseUrl}/${electionId}/phases`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ phases })
      });
      
      if (!response.ok) throw new Error('Failed to update phases');
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }
  
  async function updateModeSettings(settings: ModeSettings) {
    if (!electionId) return;
    
    setLoading(true);
    try {
      const response = await fetch(`${baseUrl}/${electionId}/settings/mode`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(settings)
      });
      
      if (!response.ok) throw new Error('Failed to update mode settings');
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }
  
  // UPLOAD
  async function uploadLogo(file: File, slot: 'primary' | 'secondary') {
    if (!electionId) return;
    
    setLoading(true);
    try {
      const formData = new FormData();
      formData.append('file', file);
      
      const response = await fetch(`${baseUrl}/${electionId}/branding/logo/${slot}`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formData
      });
      
      if (!response.ok) throw new Error('Failed to upload logo');
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }
  
  // DELETE
  async function deleteLogo(slot: 'primary' | 'secondary') {
    if (!electionId) return;
    
    setLoading(true);
    try {
      const response = await fetch(`${baseUrl}/${electionId}/branding/logo/${slot}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (!response.ok) throw new Error('Failed to delete logo');
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }
  
  return {
    loading,
    error,
    createElection,
    getSettings,
    updateElection,
    updatePhases,
    updateModeSettings,
    uploadLogo,
    deleteLogo
  };
}

// Usage in component
function ElectionSettingsPage({ electionId }) {
  const {
    loading,
    error,
    getSettings,
    updateElection,
    updatePhases,
    uploadLogo
  } = useElectionCRUD(electionId);
  
  const [settings, setSettings] = useState(null);
  
  useEffect(() => {
    async function load() {
      const data = await getSettings();
      setSettings(data);
    }
    load();
  }, [electionId]);
  
  async function handleSubmit(formData) {
    await updateElection(formData);
    await updatePhases(formData.phases);
    toast.success('Settings updated!');
  }
  
  return (
    <div>
      {loading && <Spinner />}
      {error && <ErrorMessage message={error} />}
      {settings && (
        <ElectionForm 
          initialData={settings}
          onSubmit={handleSubmit}
        />
      )}
    </div>
  );
}
```

---

## 📚 Documentation Files

- **ELECTION_SETTINGS_API.md** - GET endpoints (Read)
- **ELECTION_UPDATE_API.md** - UPDATE endpoints
- **ELECTION_SETTINGS_QUICK_REFERENCE.md** - Quick ref for GET
- **ELECTION_UPDATE_QUICK_REFERENCE.md** - Quick ref for UPDATE
- **ELECTION_CRUD_COMPLETE.md** - This file (Complete CRUD)

---

## ✅ Checklist: Complete Election Setup

- [ ] Create election (`POST /elections`)
- [ ] Update phases (`PUT /elections/{id}/phases`)
- [ ] Configure mode settings (`PUT /elections/{id}/settings/mode`)
- [ ] Upload primary logo (`POST /elections/{id}/branding/logo/primary`)
- [ ] Upload secondary logo (optional)
- [ ] Verify settings (`GET /elections/{id}/settings`)
- [ ] Import voters (`POST /elections/{id}/voters/import`)
- [ ] Add candidates (`POST /elections/{id}/candidates`)
- [ ] Test voting flow
- [ ] Open voting (`POST /elections/{id}/actions/open-voting`)

---

**Created**: 2025-11-25  
**Status**: ✅ Production Ready  
**API Version**: v1
