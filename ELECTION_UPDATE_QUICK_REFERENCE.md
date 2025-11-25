# Election Update API - Quick Reference Card

## 🚀 Update Endpoints Summary

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/admin/elections/{id}` | PUT | Update general info (full) |
| `/admin/elections/{id}` | PATCH | Update general info (partial) |
| `/admin/elections/{id}/phases` | PUT | Update phases schedule |
| `/admin/elections/{id}/settings/mode` | PUT | Update mode settings |
| `/admin/elections/{id}/branding/logo/{slot}` | POST | Upload logo |
| `/admin/elections/{id}/branding/logo/{slot}` | DELETE | Delete logo |

---

## ⚡ Quick Examples

### 1. Update Election Name (Partial)
```bash
curl -X PATCH "http://localhost:8080/api/v1/admin/elections/1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "New Election Name"}'
```

### 2. Update Full Election Info
```bash
curl -X PUT "http://localhost:8080/api/v1/admin/elections/1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pemilihan Raya BEM 2024",
    "description": "Description",
    "year": 2024,
    "slug": "PEMIRA-2024",
    "academic_year": "2023/2024"
  }'
```

### 3. Update Phases
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
      },
      ... (all 6 phases)
    ]
  }'
```

### 4. Update Mode Settings
```bash
curl -X PUT "http://localhost:8080/api/v1/admin/elections/1/settings/mode" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "online_enabled": true,
    "tps_enabled": false
  }'
```

### 5. Upload Logo
```bash
curl -X POST "http://localhost:8080/api/v1/admin/elections/1/branding/logo/primary" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@logo.png"
```

### 6. Delete Logo
```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/elections/1/branding/logo/primary" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🎨 Frontend Integration

### React Form Submit Handler
```typescript
async function handleSubmit(formData: ElectionFormData) {
  const token = localStorage.getItem('access_token');
  
  try {
    // Update general info
    await fetch(`/api/v1/admin/elections/${electionId}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        name: formData.name,
        description: formData.description,
        year: formData.year,
        slug: formData.slug,
        academic_year: formData.academicYear
      })
    });
    
    // Update phases
    await fetch(`/api/v1/admin/elections/${electionId}/phases`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ phases: formData.phases })
    });
    
    // Update mode settings
    await fetch(`/api/v1/admin/elections/${electionId}/settings/mode`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        online_enabled: formData.onlineEnabled,
        tps_enabled: formData.tpsEnabled,
        tps_settings: formData.tpsSettings
      })
    });
    
    toast.success('Settings updated successfully!');
  } catch (error) {
    toast.error('Failed to update settings');
  }
}
```

### Upload Logo Handler
```typescript
async function handleLogoUpload(file: File, slot: 'primary' | 'secondary') {
  const token = localStorage.getItem('access_token');
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetch(
    `/api/v1/admin/elections/${electionId}/branding/logo/${slot}`,
    {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: formData
    }
  );
  
  if (response.ok) {
    const result = await response.json();
    console.log('Logo uploaded:', result.id);
  }
}
```

---

## 📋 Request Body Templates

### General Info (PUT)
```json
{
  "name": "string (required)",
  "description": "string",
  "year": 2024,
  "slug": "string (required, unique)",
  "academic_year": "2023/2024"
}
```

### Phases (PUT)
```json
{
  "phases": [
    {
      "key": "REGISTRATION|VERIFICATION|CAMPAIGN|QUIET_PERIOD|VOTING|RECAP",
      "label": "string",
      "start_at": "2025-11-01T00:00:00+07:00",
      "end_at": "2025-11-30T23:59:59+07:00"
    }
  ]
}
```

### Mode Settings (PUT)
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

---

## ✅ Validation Checklist

### Before Update General Info
- [ ] Name is not empty
- [ ] Slug is unique
- [ ] Year is valid
- [ ] All required fields present

### Before Update Phases
- [ ] All 6 phases included
- [ ] Phase keys are valid
- [ ] start_at < end_at for each phase
- [ ] Dates in ISO 8601 format with timezone

### Before Update Mode
- [ ] At least one mode enabled (online OR tps)
- [ ] Boolean values for flags
- [ ] TPS settings valid if tps_enabled = true

### Before Upload Logo
- [ ] File size < 2 MB
- [ ] File format: PNG, JPG, JPEG, or SVG
- [ ] Slot is 'primary' or 'secondary'

---

## 🐛 Common Errors

| Error Code | Reason | Solution |
|------------|--------|----------|
| `VALIDATION_ERROR` | Invalid request body | Check required fields |
| `INVALID_PHASE_KEY` | Wrong phase key | Use valid keys (REGISTRATION, etc) |
| `UNAUTHORIZED` | Token expired | Login again |
| `ELECTION_NOT_FOUND` | Invalid election ID | Verify election exists |
| `FILE_TOO_LARGE` | Logo > 2MB | Compress image |

---

## 🔄 Update Flow Diagram

```
┌─────────────────────────────────────────┐
│  1. Load Current Settings               │
│     GET /elections/{id}/settings        │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  2. User Modifies Form                  │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  3. Update General Info                 │
│     PUT /elections/{id}                 │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  4. Update Phases                       │
│     PUT /elections/{id}/phases          │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  5. Update Mode Settings                │
│     PUT /elections/{id}/settings/mode   │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  6. Upload Logo (if changed)            │
│     POST /elections/{id}/branding/logo  │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  7. Show Success Message                │
│     Reload settings to verify           │
└─────────────────────────────────────────┘
```

---

## 💡 Pro Tips

1. **Use PATCH for single field updates** - Lebih efisien daripada PUT
2. **Validate on client side first** - Kurangi error responses
3. **Handle errors gracefully** - Show user-friendly messages
4. **Optimistic updates** - Update UI immediately, rollback on error
5. **Debounce rapid updates** - Avoid rate limiting
6. **Cache logo uploads** - Don't re-upload unchanged logos

---

## 🔗 Related Endpoints

| Purpose | Endpoint | Method |
|---------|----------|--------|
| Get all settings | `/elections/{id}/settings` | GET |
| Get phases only | `/elections/{id}/phases` | GET |
| Get mode settings | `/elections/{id}/settings/mode` | GET |
| Get branding | `/elections/{id}/branding` | GET |
| Open voting | `/elections/{id}/actions/open-voting` | POST |
| Close voting | `/elections/{id}/actions/close-voting` | POST |

---

**📚 Full Documentation**: See `ELECTION_UPDATE_API.md`
