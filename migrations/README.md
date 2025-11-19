# PEMIRA UNIWA Database Migrations

## Urutan Eksekusi Migration

**PENTING**: Migration harus dijalankan dengan urutan yang benar karena dependency antar tabel.

### Urutan yang Benar:

1. **003_create_core_tables** - Tabel inti (elections, voters, vote_tokens)
2. **004_create_supporting_tables** - Tabel pendukung (user_accounts, candidates)  
3. **002_create_tps_foundation** - Sistem TPS (tps, tps_qr, tps_checkins, voter_status, votes)

### Skema Database Lengkap

```
elections (master pemilu)
  ├─> candidates (paslon)
  ├─> voters (mahasiswa pemilih)
  │     └─> voter_status (status per election)
  │     └─> vote_tokens (bukti suara)
  ├─> tps (tempat pemungutan suara)
  │     ├─> tps_qr (QR code TPS)
  │     └─> tps_checkins (log scan QR)
  └─> votes (suara terenkripsi)

user_accounts (panitia & admin)
  └─> tps_checkins.approved_by_id
```

## ENUM Types

- `election_status`: DRAFT, REGISTRATION, CAMPAIGN, VOTING_OPEN, CLOSED, ARCHIVED
- `academic_status`: ACTIVE, GRADUATED, ON_LEAVE, DROPPED, INACTIVE
- `user_role`: ADMIN, PANITIA, KETUA_TPS, OPERATOR_PANEL, VIEWER
- `candidate_status`: PENDING, APPROVED, REJECTED, WITHDRAWN
- `tps_status`: DRAFT, ACTIVE, CLOSED
- `tps_checkin_status`: PENDING, APPROVED, REJECTED, USED, EXPIRED
- `voting_method`: ONLINE, TPS
- `vote_channel`: ONLINE, TPS

## Fitur Keamanan & Integritas

✅ **Foreign Key Constraints** dengan ON DELETE strategy yang tepat  
✅ **Unique Constraints** untuk business rules (NIM, token_hash, election_code)  
✅ **Check Constraints** untuk validasi state (voter_status, elections)  
✅ **Composite Indexes** untuk query optimization  
✅ **Partial Indexes** untuk kondisi spesifik (WHERE token_hash IS NOT NULL)  
✅ **Auto-update Triggers** untuk timestamp tracking  

## Catatan Migrasi

Jika migration 001 sudah dijalankan, bisa di-drop dulu atau skip karena sudah digantikan oleh skema baru yang lebih lengkap.

### Drop old migration (opsional):
```sql
-- Jika perlu clean slate
DROP TABLE IF EXISTS tps_checkins CASCADE;
DROP TABLE IF EXISTS tps_panitia CASCADE;
DROP TABLE IF EXISTS tps_qr CASCADE;
DROP TABLE IF EXISTS tps CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column();
```
