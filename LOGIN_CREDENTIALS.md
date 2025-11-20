# QUICK REFERENCE - Login Credentials

## All Users Password
**Password untuk semua user:** `password123`

## Admin & Staff System
```
admin               - ADMIN
panitia             - PANITIA  
ketua_tps1          - KETUA_TPS (TPS Gedung A)
operator            - OPERATOR_PANEL
viewer              - VIEWER
```

## Mahasiswa (Students)
Login dengan NIM:
```
2021001             - Andi Pratama (Teknik Informatika, 2021)
2021002             - Budi Santoso (Sistem Informasi, 2021)
2022001             - Citra Dewi (Teknik Elektro, 2022)
2022002             - Dian Putri (Manajemen, 2022)
```

## Dosen (Lecturers)
Login dengan NIDN:
```
0101018901          - Dr. Ahmad Kusuma, S.Kom., M.T.
                      (Teknik Informatika - Lektor Kepala)

0102019002          - Dra. Siti Nurjanah, M.Pd.
                      (PGSD - Lektor)

0103019103          - Prof. Dr. Budi Santoso, S.E., M.M.
                      (Manajemen - Guru Besar)

0104019204          - Dr. Retno Wulandari, S.Si., M.Sc.
                      (Matematika - Lektor)

0105019305          - Ir. Joko Widodo, M.T.
                      (Teknik Sipil - Asisten Ahli)
```

## Staff Universitas
Login dengan NIP:
```
198901012015041001  - Bambang Setiawan, S.Sos.
                      (Biro Administrasi Umum - Kepala Sub Bagian Umum)

199002012016051002  - Dewi Kusumawati, A.Md.
                      (BAAK - Staff Administrasi Akademik)

199103012017061003  - Eko Prasetyo, S.Kom.
                      (UPT-TIK - Administrator Sistem)

199204012018071004  - Fitri Handayani, S.E.
                      (Biro Administrasi Keuangan - Bendahara)

199305012019081005  - Gunawan Wijaya
                      (Perpustakaan - Pustakawan)
```

## API Login Example

```bash
# Login Admin
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password123"}'

# Login Mahasiswa
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "2021001", "password": "password123"}'

# Login Dosen
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "0101018901", "password": "password123"}'

# Login Staff
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "198901012015041001", "password": "password123"}'
```

## Role Summary
- **ADMIN** - Administrator sistem (full access)
- **PANITIA** - Panitia pemilihan
- **KETUA_TPS** - Ketua TPS
- **OPERATOR_PANEL** - Operator panel
- **VIEWER** - Read-only access
- **STUDENT** - Mahasiswa yang bisa memilih
- **LECTURER** - Dosen yang bisa memilih
- **STAFF** - Staff universitas yang bisa memilih
