# TPS Ballot QR Voting API (Offline Device)

Kontrak API untuk perangkat pemilih (mode OFFLINE/TPS) ketika memindai QR paslon dan mengirim suara. Fokus pada dua endpoint: parse QR (helper UI) dan cast vote dari QR (final commit).

## 0. Scope, Prasyarat, dan Aturan Global
- **Auth**: `Authorization: Bearer <JWT>` yang ter-resolve ke `user_account` → `voter_id`. Role wajib `VOTER`.
- **Mode**: `voter.mode` untuk pemilu terkait harus `OFFLINE_TPS` (bukan online).
- **DPT**: Voter terdaftar di DPT pemilu tersebut.
- **TPS Check-in**: Untuk cast suara harus sudah `CHECKED_IN` di TPS (row di `tps_checkins` untuk election itu).
- **Status**: `voter_status.has_voted` masih `false`.
- **Response envelope**: Selalu gunakan `success` + `data` atau `error`.
  ```json
  {
    "success": true,
    "data": { ... }
  }
  ```
  ```json
  {
    "success": false,
    "error": { "code": "ALREADY_VOTED", "message": "..." }
  }
  ```

## 1. QR Paslon — Format dan Resolusi
- **Payload mentah**: string hasil scan, contoh `PEMIRA-UNIWA|E:3|C:7|V:1`.
- **Rules parse**:
  - Prefix wajib `PEMIRA-UNIWA`.
  - `E:{election_id}` integer.
  - `C:{candidate_id}` integer.
  - `V:{version}` integer.
- **Lookup**: Cari di `candidate_qr_codes` dengan `(election_id, candidate_id, version, is_active = true)`. Kalau tidak ketemu → `INVALID_BALLOT_QR`.
- **Election match**: `election_id` di QR harus sama dengan election yang sedang dijalani voter (explicit `election_id` request atau resolved dari QR / current offline election).

## 2. Endpoint — Parse QR Paslon (Opsional, UI Helper)
Tujuan: setelah scan, sebelum commit, menampilkan konfirmasi “Anda akan memilih PASLON 02 – Budi & Rian”.

- **Method & Path**: `POST /voting/tps/ballots/parse-qr`
- **Auth**: Bearer JWT pemilih (role `VOTER`)
- **Body**:
  ```json
  { "ballot_qr_payload": "PEMIRA-UNIWA|E:3|C:7|V:1" }
  ```
- **Validasi**:
  1) JWT valid, role `VOTER`.
  2) Voter mode = `OFFLINE_TPS`; jika tidak → `NOT_TPS_VOTER`.
  3) Parse payload; jika gagal/prefix salah → `INVALID_BALLOT_QR`.
  4) Resolve election_id (dari QR atau request), pastikan voter terkait election itu dan election cocok; jika tidak → `ELECTION_MISMATCH`.
  5) Resolve kandidat via `candidate_qr_codes` aktif; jika tidak → `INVALID_BALLOT_QR`.
  6) Tidak ada perubahan status/DB.
- **Response 200 (sukses)**:
  ```json
  {
    "success": true,
    "data": {
      "election_id": 3,
      "election_name": "PEMIRA UNIWA 2025",
      "candidate_id": 7,
      "candidate_number": "02",
      "candidate_name": "Budi Pratama",
      "candidate_vice_name": "Rian Darmawan",
      "version": 1
    }
  }
  ```
- **Error samples**:
  - `INVALID_BALLOT_QR`: QR tidak dikenali/format salah.
  - `ELECTION_MISMATCH`: QR bukan untuk election aktif pemilih.
  - `NOT_TPS_VOTER`: mode pemilih bukan TPS.

## 3. Endpoint — Cast Vote dari QR Paslon (Final Commit)
Endpoint utama saat pemilih menekan “KIRIM SUARA”.

- **Method & Path**: `POST /voting/tps/ballots/cast-from-qr`
- **Auth**: Bearer JWT pemilih (role `VOTER`)
- **Body**:
  ```json
  { "ballot_qr_payload": "PEMIRA-UNIWA|E:3|C:7|V:1" }
  ```
  Opsional: sertakan `election_id` eksplisit.
  ```json
  { "election_id": 3, "ballot_qr_payload": "PEMIRA-UNIWA|E:3|C:7|V:1" }
  ```

### 3.1 Pre-check Business Rules
1) JWT valid, role `VOTER`.
2) Voter ada di `voters` + `voter_status` untuk election tersebut; `mode = OFFLINE_TPS` / `channel_preferred = TPS`; `has_voted = false`; jika sudah → `ALREADY_VOTED`.
3) Ada check-in aktif: ambil `tps_checkins` terbaru (election_id, voter_id) status `CHECKED_IN`; jika tidak → `NO_ACTIVE_CHECKIN`.
4) Parse QR payload; prefix & format valid; jika tidak → `INVALID_BALLOT_QR`.
5) Resolve QR ke `candidate_qr_codes` (election_id, candidate_id, version, is_active = true); jika tidak → `INVALID_BALLOT_QR`.
6) Election di QR = election yang sedang dijalani voter; jika tidak → `ELECTION_MISMATCH`.

### 3.2 Transaction Flow (all-or-nothing)
1) `BEGIN;`
2) Lock `voter_status` row (`FOR UPDATE`) untuk election_id, voter_id; re-check `has_voted = false`.
3) Re-check check-in status masih `CHECKED_IN` (optional lock row).
4) Insert `tps_ballot_scans`:
   - election_id, tps_id (dari check-in), checkin_id, voter_id
   - candidate_id, candidate_qr_id
   - raw_payload, payload_valid = true
   - status = `APPLIED`
   - scanned_by_user_id = user_id
   - dapat `ballot_scan_id`
5) Insert `votes`:
   - election_id, voter_id, candidate_id
   - channel = `TPS`, tps_id (dari check-in)
   - candidate_qr_id, ballot_scan_id
   - unique constraint (election_id, voter_id) → kalau sudah ada: abort dengan `ALREADY_VOTED` atau `DUPLICATE_VOTE_ATTEMPT`.
6) Update `voter_status`:
   - `has_voted = true`
   - `last_vote_channel = 'TPS'`
   - `last_tps_id = tps_id`
   - `last_vote_at = now()`
7) Update `tps_checkins`:
   - `status = 'VOTED'`
   - `voted_at = now()`
8) `COMMIT;`

### 3.3 Response 200 (sukses)
```json
{
  "success": true,
  "data": {
    "election_id": 3,
    "voted_at": "2025-11-21T10:23:45Z",
    "channel": "TPS",
    "tps": { "id": 3, "code": "TPS03", "name": "TPS Aula Utama" },
    "status": "VOTED"
  }
}
```
Catatan: nama paslon boleh disertakan untuk feedback ke pemilih, tetapi jangan dipakai/expose di panel TPS.

### 3.4 Error cases utama
- `INVALID_BALLOT_QR` (400/422): Format QR salah / QR tidak terdaftar aktif.
- `ELECTION_MISMATCH` (400): QR bukan untuk election pemilih.
- `NOT_TPS_VOTER` (400): Mode pemilih bukan TPS.
- `NO_ACTIVE_CHECKIN` (400/404): Belum check-in TPS.
- `ALREADY_VOTED` (409): Suara sudah tercatat.
- `DUPLICATE_VOTE_ATTEMPT` (409): Race/double submit; treat sama dengan already voted.

## 4. Error Code Ringkas
| Code | HTTP | Deskripsi |
|------|------|-----------|
| INVALID_BALLOT_QR | 400/422 | QR tidak bisa diparse atau tidak terdaftar aktif |
| ELECTION_MISMATCH | 400 | QR bukan untuk election pemilih |
| NOT_TPS_VOTER | 400 | Voter mode bukan TPS/offline |
| NO_ACTIVE_CHECKIN | 400/404 | Tidak ada check-in TPS aktif |
| ALREADY_VOTED | 409 | Sudah pernah voting untuk election ini |
| DUPLICATE_VOTE_ATTEMPT | 409 | Double submit / race condition terdeteksi |
| UNAUTHORIZED | 401 | JWT invalid/expired |

## 5. Pseudo-Swagger (ringkas)
```yaml
paths:
  /voting/tps/ballots/parse-qr:
    post:
      summary: Parse QR paslon (tanpa merekam suara)
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [ballot_qr_payload]
              properties:
                ballot_qr_payload:
                  type: string
      responses:
        "200":
          description: Parsed successfully or error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiResponse'

  /voting/tps/ballots/cast-from-qr:
    post:
      summary: Rekam suara TPS dari QR paslon (device pemilih)
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [ballot_qr_payload]
              properties:
                election_id:
                  type: integer
                  nullable: true
                ballot_qr_payload:
                  type: string
      responses:
        "200":
          description: Vote recorded or error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiResponse'
```

## 6. Keamanan & Privasi
- Endpoint ini hanya untuk device pemilih, bukan panel TPS.
- Panel TPS hanya membaca agregat (`tps_checkins.status`, agregasi votes per TPS).
- Tidak ada API yang memaparkan “pemilih X memilih kandidat Y” ke pihak lain; feedback nama paslon hanya boleh ke pemilih itu sendiri jika diperlukan UI.
