# Admin Election Management - Quick Reference

## Quick Start

### 1. Create Election
```bash
curl -X POST http://localhost:8080/api/v1/admin/elections \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "year": 2024,
    "name": "Pemilu Raya 2024",
    "slug": "pemira-2024",
    "online_enabled": true,
    "tps_enabled": true
  }'
```

### 2. Open Voting
```bash
curl -X POST http://localhost:8080/api/v1/admin/elections/1/open-voting \
  -H "Authorization: Bearer <admin_token>"
```

### 3. Close Voting
```bash
curl -X POST http://localhost:8080/api/v1/admin/elections/1/close-voting \
  -H "Authorization: Bearer <admin_token>"
```

---

## Common Tasks

### List All Elections
```bash
curl http://localhost:8080/api/v1/admin/elections \
  -H "Authorization: Bearer <admin_token>"
```

### Filter by Year
```bash
curl "http://localhost:8080/api/v1/admin/elections?year=2024" \
  -H "Authorization: Bearer <admin_token>"
```

### Filter by Status
```bash
curl "http://localhost:8080/api/v1/admin/elections?status=VOTING_OPEN" \
  -H "Authorization: Bearer <admin_token>"
```

### Search by Name
```bash
curl "http://localhost:8080/api/v1/admin/elections?search=pemira" \
  -H "Authorization: Bearer <admin_token>"
```

### Get Election Detail
```bash
curl http://localhost:8080/api/v1/admin/elections/1 \
  -H "Authorization: Bearer <admin_token>"
```

---

## Toggle Voting Mode

### Disable Online Voting
```bash
curl -X PUT http://localhost:8080/api/v1/admin/elections/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"online_enabled": false}'
```

### Enable Online Voting
```bash
curl -X PUT http://localhost:8080/api/v1/admin/elections/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"online_enabled": true}'
```

### TPS Only Mode
```bash
curl -X PUT http://localhost:8080/api/v1/admin/elections/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "online_enabled": false,
    "tps_enabled": true
  }'
```

### Online Only Mode
```bash
curl -X PUT http://localhost:8080/api/v1/admin/elections/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "online_enabled": true,
    "tps_enabled": false
  }'
```

### Hybrid Mode (Both)
```bash
curl -X PUT http://localhost:8080/api/v1/admin/elections/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "online_enabled": true,
    "tps_enabled": true
  }'
```

---

## Update Election Info

### Update Name
```bash
curl -X PUT http://localhost:8080/api/v1/admin/elections/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Pemilu Raya 2024 (Updated)"}'
```

### Update Multiple Fields
```bash
curl -X PUT http://localhost:8080/api/v1/admin/elections/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pemilu Raya 2024 v2",
    "slug": "pemira-2024-v2",
    "online_enabled": false
  }'
```

---

## Election Status Flow

```
CREATE (DRAFT)
    ↓
OPEN VOTING (VOTING_OPEN)
    ↓
CLOSE VOTING (VOTING_CLOSED)
```

### Status Enum
- `DRAFT`: Election created, not yet open
- `REGISTRATION`: Registration phase (optional)
- `CAMPAIGN`: Campaign phase (optional)
- `VOTING_OPEN`: Voting is active
- `VOTING_CLOSED`: Voting has ended
- `CLOSED`: Final results published (optional)
- `ARCHIVED`: Election archived

---

## Code Examples

### Go Client

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type CreateElectionRequest struct {
    Year          int    `json:"year"`
    Name          string `json:"name"`
    Slug          string `json:"slug"`
    OnlineEnabled bool   `json:"online_enabled"`
    TPSEnabled    bool   `json:"tps_enabled"`
}

func createElection(token string) error {
    req := CreateElectionRequest{
        Year:          2024,
        Name:          "Pemilu Raya 2024",
        Slug:          "pemira-2024",
        OnlineEnabled: true,
        TPSEnabled:    true,
    }

    body, _ := json.Marshal(req)
    
    httpReq, _ := http.NewRequest(
        "POST",
        "http://localhost:8080/api/v1/admin/elections",
        bytes.NewBuffer(body),
    )
    
    httpReq.Header.Set("Authorization", "Bearer "+token)
    httpReq.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    fmt.Printf("Status: %d\n", resp.StatusCode)
    return nil
}

func openVoting(token string, electionID int) error {
    url := fmt.Sprintf(
        "http://localhost:8080/api/v1/admin/elections/%d/open-voting",
        electionID,
    )
    
    httpReq, _ := http.NewRequest("POST", url, nil)
    httpReq.Header.Set("Authorization", "Bearer "+token)
    
    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    fmt.Printf("Status: %d\n", resp.StatusCode)
    return nil
}
```

### JavaScript/TypeScript Client

```typescript
const API_BASE = 'http://localhost:8080/api/v1';

interface CreateElectionRequest {
  year: number;
  name: string;
  slug: string;
  online_enabled: boolean;
  tps_enabled: boolean;
}

async function createElection(
  token: string,
  data: CreateElectionRequest
): Promise<any> {
  const response = await fetch(`${API_BASE}/admin/elections`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  
  return response.json();
}

async function openVoting(token: string, electionId: number): Promise<any> {
  const response = await fetch(
    `${API_BASE}/admin/elections/${electionId}/open-voting`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    }
  );
  
  return response.json();
}

async function closeVoting(token: string, electionId: number): Promise<any> {
  const response = await fetch(
    `${API_BASE}/admin/elections/${electionId}/close-voting`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    }
  );
  
  return response.json();
}

async function toggleOnlineMode(
  token: string,
  electionId: number,
  enabled: boolean
): Promise<any> {
  const response = await fetch(
    `${API_BASE}/admin/elections/${electionId}`,
    {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ online_enabled: enabled }),
    }
  );
  
  return response.json();
}
```

---

## Response Examples

### Success Response (Election Created)
```json
{
  "id": 1,
  "year": 2024,
  "name": "Pemilu Raya 2024",
  "slug": "pemira-2024",
  "status": "DRAFT",
  "voting_start_at": null,
  "voting_end_at": null,
  "online_enabled": true,
  "tps_enabled": true,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

### Success Response (Voting Opened)
```json
{
  "id": 1,
  "year": 2024,
  "name": "Pemilu Raya 2024",
  "slug": "pemira-2024",
  "status": "VOTING_OPEN",
  "voting_start_at": "2024-03-01T00:00:00Z",
  "voting_end_at": null,
  "online_enabled": true,
  "tps_enabled": true,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-03-01T00:00:00Z"
}
```

### Error Response
```json
{
  "error": {
    "code": "ELECTION_ALREADY_OPEN",
    "message": "Pemilu sudah dalam status voting terbuka."
  }
}
```

---

## Integration Flow

### Complete Election Lifecycle

1. **Create Election**
```bash
POST /admin/elections
→ Status: DRAFT
```

2. **Import Voters (DPT)**
```bash
POST /admin/elections/1/voters/import
→ Upload CSV/Excel
```

3. **Add Candidates**
```bash
POST /admin/elections/1/candidates
→ Create candidates
```

4. **Open Voting**
```bash
POST /admin/elections/1/open-voting
→ Status: VOTING_OPEN
```

5. **Monitor Voting**
```bash
# Check current election
GET /elections/current

# Check voter status
GET /elections/1/me/status

# Monitor stats (future feature)
GET /admin/elections/1/stats
```

6. **Toggle Mode (if needed)**
```bash
# Disable online voting
PUT /admin/elections/1
{"online_enabled": false}
```

7. **Close Voting**
```bash
POST /admin/elections/1/close-voting
→ Status: VOTING_CLOSED
```

8. **View Results**
```bash
# Get final results (future feature)
GET /admin/elections/1/results
```

---

## Troubleshooting

### Cannot Open Voting
**Error**: `ELECTION_ALREADY_OPEN`
- Election is already in VOTING_OPEN status
- Check current status: `GET /admin/elections/{id}`

**Error**: `INVALID_STATUS_CHANGE`
- Cannot open archived elections
- Create new election instead

### Cannot Close Voting
**Error**: `ELECTION_NOT_OPEN`
- Election must be in VOTING_OPEN status to close
- Check current status first

### Toggle Not Working
- Make sure to send valid JSON
- Check that you're updating the correct election ID
- Verify admin token is valid

---

## Best Practices

1. **Before Opening Voting**
   - ✅ Import all voters (DPT)
   - ✅ Add all candidates
   - ✅ Configure voting mode
   - ✅ Test with sample accounts

2. **During Voting**
   - ⚠️ Avoid changing election name/slug
   - ⚠️ Toggle voting mode only if necessary
   - ✅ Monitor voter status
   - ✅ Check for issues

3. **After Closing Voting**
   - ✅ Verify all votes are counted
   - ✅ Export results
   - ✅ Archive election if needed

---

## Related Documentation

- [Full API Documentation](./ADMIN_ELECTION_API.md)
- [Implementation Details](./ADMIN_ELECTION_IMPLEMENTATION.md)
- [DPT Management](./DPT_API_DOCUMENTATION.md)
- [Voting System](./VOTING_API_IMPLEMENTATION.md)
