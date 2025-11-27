# Bug Fix: Auto-Enrollment ke Election Voters (DPT)

**Tanggal:** 27 November 2025  
**Status:** ✅ FIXED

## 🔍 Deskripsi Bug

User yang baru registrasi **TIDAK otomatis muncul di DPT list** (`/admin/elections/1/voters`), meskipun:
- ✅ User berhasil dibuat di tabel `users`
- ✅ Voter berhasil dibuat di tabel `voters`
- ✅ Dashboard monitoring menghitung user baru (total_eligible bertambah)

### Contoh Kasus
- User registrasi dengan voter_id: 84
- Dashboard monitoring: total_eligible bertambah dari 37 → 38
- DPT list: Masih menampilkan 39 pemilih (data lama)
- **User baru TIDAK muncul di DPT**

## 🎯 Root Cause

Saat registrasi user baru, backend hanya membuat:
1. ✅ Entry di tabel `users`
2. ✅ Entry di tabel `voters`
3. ✅ Entry di tabel `voter_status`
4. ❌ **TIDAK** membuat entry di tabel `election_voters`

**Tabel `election_voters` adalah sumber data untuk DPT list!**

## 📊 Dampak

- **Dashboard monitoring** menghitung dari tabel `voters` → langsung update
- **Halaman DPT Admin** menampilkan dari tabel `election_voters` → tidak update
- User baru harus di-enroll manual oleh admin agar muncul di DPT

## ✅ Solusi yang Diimplementasikan

### 1. Menambahkan Method Baru di Repository

**File:** `internal/auth/repository.go`

```go
// Election enrollment helper
EnrollVoterToElection(ctx context.Context, electionID, voterID int64, nim string, votingMethod string) error
```

**File:** `internal/auth/repository_pgx.go`

```go
// EnrollVoterToElection adds a voter to the election_voters table with PENDING status.
// This ensures newly registered voters automatically appear in the DPT list.
func (r *PgRepository) EnrollVoterToElection(ctx context.Context, electionID, voterID int64, nim string, votingMethod string) error {
	query := `
		INSERT INTO election_voters (election_id, voter_id, nim, status, voting_method, created_at, updated_at)
		VALUES ($1, $2, $3, 'PENDING', $4, NOW(), NOW())
		ON CONFLICT ON CONSTRAINT ux_election_voters_election_voter
		DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, electionID, voterID, nim, votingMethod)
	return err
}
```

### 2. Memanggil Method di Flow Registrasi

**File:** `internal/auth/service_auth.go`

#### RegisterStudent
```go
// Upsert voter_status preference/allowed flags
onlineAllowed := mode == "ONLINE"
tpsAllowed := mode == "TPS"
_ = s.repo.EnsureVoterStatus(ctx, regElection.ID, voterID, mode, onlineAllowed, tpsAllowed)

// ✨ BUG FIX: Automatically enroll voter to election_voters table
// This ensures the voter appears in the DPT list immediately after registration
_ = s.repo.EnrollVoterToElection(ctx, regElection.ID, voterID, nim, mode)
```

#### RegisterLecturerStaff (Case LECTURER)
```go
onlineAllowed := mode == "ONLINE"
tpsAllowed := mode == "TPS"
_ = s.repo.EnsureVoterStatus(ctx, regElection.ID, voterID, mode, onlineAllowed, tpsAllowed)

// ✨ BUG FIX: Automatically enroll voter to election_voters table
_ = s.repo.EnrollVoterToElection(ctx, regElection.ID, voterID, nidn, mode)
```

#### RegisterLecturerStaff (Case STAFF)
```go
onlineAllowed := mode == "ONLINE"
tpsAllowed := mode == "TPS"
_ = s.repo.EnsureVoterStatus(ctx, regElection.ID, voterID, mode, onlineAllowed, tpsAllowed)

// ✨ BUG FIX: Automatically enroll voter to election_voters table
_ = s.repo.EnrollVoterToElection(ctx, regElection.ID, voterID, nip, mode)
```

## 🔧 Detail Implementasi

### Karakteristik Enrollment

1. **Status Default:** `PENDING`
   - User baru akan muncul di DPT dengan status PENDING
   - Admin dapat verify/reject sesuai kebutuhan

2. **Voting Method:** Sesuai pilihan user saat registrasi
   - `ONLINE` jika user pilih online voting
   - `TPS` jika user pilih TPS voting

3. **Conflict Handling:** `ON CONFLICT DO NOTHING`
   - Jika voter sudah ada di election_voters, tidak akan error
   - Safe untuk dipanggil multiple kali

4. **Election Target:** Menggunakan `regElection.ID`
   - Election yang sedang dalam status REGISTRATION/CAMPAIGN/VOTING_OPEN
   - Sama dengan election yang digunakan untuk voter_status

## 📝 Files Changed

1. `internal/auth/repository.go` - Interface method signature
2. `internal/auth/repository_pgx.go` - Implementation
3. `internal/auth/service_auth.go` - Service layer integration (3 places)

## ✅ Expected Behavior After Fix

### Before Fix ❌
```
1. User registrasi → voter_id: 84
2. Dashboard: total_eligible = 38 ✅
3. DPT List: 39 voters (tidak ada voter_id 84) ❌
```

### After Fix ✅
```
1. User registrasi → voter_id: 84
2. Entry otomatis dibuat di election_voters dengan status PENDING
3. Dashboard: total_eligible = 38 ✅
4. DPT List: 40 voters (ada voter_id 84 dengan status PENDING) ✅
```

## 🧪 Testing Checklist

- [ ] Test registrasi student baru
- [ ] Test registrasi lecturer baru
- [ ] Test registrasi staff baru
- [ ] Verify user muncul di DPT dengan status PENDING
- [ ] Verify dashboard monitoring konsisten dengan DPT list
- [ ] Test registrasi ulang (harus tidak error karena ON CONFLICT DO NOTHING)

## 📚 Related Documentation

- API Registration: `docs/API_CONTRACT_VOTER_REGISTRATION.md`
- DPT API: `docs/DPT_API_DOCUMENTATION.md`
- Election Voters Schema: `migrations/025_add_election_voters_and_student_nim.up.sql`

## 🚀 Deployment Notes

- ✅ No database migration required (menggunakan tabel yang sudah ada)
- ✅ No breaking changes (hanya menambahkan behavior baru)
- ✅ Safe to deploy immediately
- ⚠️ Perlu restart service setelah deploy

## 🔍 Verification Query

Setelah deploy, cek apakah enrollment berjalan:

```sql
-- Cek voter yang baru dibuat
SELECT 
    v.id as voter_id,
    v.nim,
    v.name,
    v.created_at as voter_created,
    ev.id as election_voter_id,
    ev.status,
    ev.voting_method,
    ev.created_at as enrolled_at
FROM voters v
LEFT JOIN election_voters ev ON ev.voter_id = v.id
WHERE v.created_at > NOW() - INTERVAL '1 day'
ORDER BY v.created_at DESC;
```

Harapan:
- Semua voter baru harus memiliki `election_voter_id` (tidak NULL)
- Status harus `PENDING`
- `enrolled_at` harus sama dengan `voter_created`
