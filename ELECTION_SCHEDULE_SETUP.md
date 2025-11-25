# Election Schedule Setup Summary

## ✅ Jadwal Pemilihan Raya 2025

Jadwal tahapan pemilihan telah berhasil diatur untuk kedua election (ID 1 dan ID 2).

### 📅 Timeline Lengkap

| Tahapan | Tanggal Mulai | Tanggal Selesai | Durasi |
|---------|--------------|-----------------|--------|
| **Pendaftaran** | 01 November 2025 | 30 November 2025 | 30 hari |
| **Verifikasi Berkas** | 01 Desember 2025 | 07 Desember 2025 | 7 hari |
| **Kampanye** | 08 Desember 2025 | 10 Desember 2025 | 3 hari |
| **Masa Tenang** | 11 Desember 2025 | 14 Desember 2025 | 4 hari |
| **Voting** | 15 Desember 2025 | 17 Desember 2025 | 3 hari |
| **Rekapitulasi** | 21 Desember 2025 | 22 Desember 2025 | 2 hari |

## 🔧 Cara Update Jadwal

### Via API (Recommended)

```bash
# Login sebagai admin
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}' | jq -r '.access_token')

# Update schedule untuk election tertentu
curl -X PUT "http://localhost:8080/api/v1/admin/elections/{ELECTION_ID}/phases" \
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
  }'
```

### Via SQL

Jalankan file seed:
```bash
psql -U pemira -d pemira -f seed_election_schedule.sql
```

## 📊 Verifikasi Jadwal

### Via API
```bash
# Cek jadwal election tertentu
curl -s "http://localhost:8080/api/v1/admin/elections/{ELECTION_ID}/phases" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### Via SQL
```sql
SELECT 
  e.id AS election_id,
  e.name AS election_name,
  ep.phase_key,
  ep.phase_label,
  ep.start_at,
  ep.end_at
FROM election_phases ep
JOIN elections e ON e.id = ep.election_id
WHERE e.id = 1
ORDER BY ep.phase_order;
```

## 📝 Phase Keys yang Tersedia

| Key | Label Default |
|-----|---------------|
| `REGISTRATION` | Pendaftaran |
| `VERIFICATION` | Verifikasi Berkas |
| `CAMPAIGN` | Masa Kampanye |
| `QUIET_PERIOD` | Masa Tenang |
| `VOTING` | Voting |
| `RECAP` | Rekapitulasi |

## ⚠️ Catatan Penting

1. **Timezone**: Semua waktu menggunakan WIB (UTC+7)
2. **Format Waktu**: ISO 8601 dengan timezone (`2025-11-01T00:00:00+07:00`)
3. **Validasi**: API akan memvalidasi bahwa tidak ada overlap waktu antar tahapan
4. **Current Phase**: Sistem akan otomatis mendeteksi phase aktif berdasarkan waktu server
5. **Election Status**: Pastikan election dalam status yang sesuai sebelum mengubah jadwal

## 🎯 Status Applied

- ✅ Election ID 1 (Pemilihan Raya BEM 2024): Schedule updated
- ✅ Election ID 2 (Pemilihan Raya BEM 2025): Schedule updated
- ✅ SQL seed file created: `seed_election_schedule.sql`

## 🔄 Rollback

Jika perlu rollback ke jadwal sebelumnya, restore dari backup database atau update manual via API/SQL.

---

**Timestamp**: 2025-11-25 20:25:00 WIB
**Updated By**: Admin API
**Method**: PUT /api/v1/admin/elections/{id}/phases
