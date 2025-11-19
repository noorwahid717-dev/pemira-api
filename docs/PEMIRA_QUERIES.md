# Paket Query Utama PEMIRA UNIWA

Query collection untuk reporting & dashboard PEMIRA. Semua query pakai PostgreSQL dengan parameter `$1` untuk `election_id` (kecuali disebutkan lain).

## 📊 1. Total Suara per Kandidat

Hitung suara per kandidat dalam satu pemilu.

```sql
-- Total suara per kandidat di satu election
SELECT
    c.id              AS candidate_id,
    c.candidate_number AS candidate_number,
    c.chairman_name   AS chairman_name,
    c.vice_chairman_name AS vice_chairman_name,
    COUNT(v.id)       AS total_votes
FROM candidates c
LEFT JOIN votes v
    ON v.candidate_id = c.id
   AND v.election_id = $1
WHERE c.election_id = $1
GROUP BY c.id, c.candidate_number, c.chairman_name, c.vice_chairman_name
ORDER BY c.candidate_number;
```

**Variant: Sort by votes (descending)**
```sql
ORDER BY total_votes DESC, c.candidate_number;
```

---

## 📈 2. Breakdown ONLINE vs TPS per Kandidat

Pecah suara per channel (ONLINE/TPS) untuk setiap kandidat.

```sql
-- Suara per kandidat, dipecah per channel
SELECT
    c.id                         AS candidate_id,
    c.candidate_number           AS candidate_number,
    c.chairman_name              AS chairman_name,
    c.vice_chairman_name         AS vice_chairman_name,
    COALESCE(SUM(CASE WHEN v.channel = 'ONLINE' THEN 1 ELSE 0 END), 0) AS votes_online,
    COALESCE(SUM(CASE WHEN v.channel = 'TPS'    THEN 1 ELSE 0 END), 0) AS votes_tps,
    COUNT(v.id)                  AS total_votes
FROM candidates c
LEFT JOIN votes v
    ON v.candidate_id = c.id
   AND v.election_id = $1
WHERE c.election_id = $1
GROUP BY c.id, c.candidate_number, c.chairman_name, c.vice_chairman_name
ORDER BY c.candidate_number;
```

---

## 🏢 3. Suara per Kandidat per TPS

Lihat distribusi suara per TPS untuk analisis geografis.

```sql
-- Hasil per TPS: berapa suara tiap kandidat di TPS tersebut
SELECT
    t.id               AS tps_id,
    t.code             AS tps_code,
    t.name             AS tps_name,
    c.id               AS candidate_id,
    c.candidate_number AS candidate_number,
    c.chairman_name    AS chairman_name,
    COUNT(v.id)        AS total_votes
FROM tps t
JOIN votes v
    ON v.tps_id = t.id
   AND v.election_id = $1
JOIN candidates c
    ON c.id = v.candidate_id
   AND c.election_id = $1
WHERE t.election_id = $1
GROUP BY t.id, t.code, t.name, c.id, c.candidate_number, c.chairman_name
ORDER BY t.code, c.candidate_number;
```

---

## 🎓 4. Partisipasi (Turnout) per Fakultas

Turnout rate berdasarkan fakultas untuk analisis demografis.

```sql
-- Turnout per fakultas
WITH base AS (
    SELECT
        vs.election_id,
        v.faculty_code,
        v.faculty_name,
        vs.has_voted
    FROM voter_status vs
    JOIN voters v
      ON v.id = vs.voter_id
    WHERE vs.election_id = $1
      AND vs.is_eligible = TRUE
)
SELECT
    faculty_code,
    faculty_name,
    COUNT(*)                                 AS total_eligible,
    SUM(CASE WHEN has_voted THEN 1 ELSE 0 END) AS total_voted,
    ROUND(
        SUM(CASE WHEN has_voted THEN 1 ELSE 0 END)::NUMERIC
        / NULLIF(COUNT(*), 0) * 100,
        2
    ) AS turnout_percent
FROM base
GROUP BY faculty_code, faculty_name
ORDER BY faculty_name;
```

**Variant: Turnout per Prodi**
```sql
-- Tambahkan di base CTE:
v.study_program_code,
v.study_program_name,

-- Tambahkan di GROUP BY & SELECT:
GROUP BY faculty_code, faculty_name, study_program_code, study_program_name
```

---

## 🌐 5. Partisipasi Global (Turnout Overall)

Turnout rate total untuk satu pemilu.

```sql
-- Turnout total satu pemilu
SELECT
    COUNT(*) FILTER (WHERE is_eligible)                      AS total_eligible,
    COUNT(*) FILTER (WHERE is_eligible AND has_voted)        AS total_voted,
    ROUND(
        COUNT(*) FILTER (WHERE is_eligible AND has_voted)::NUMERIC
        / NULLIF(COUNT(*) FILTER (WHERE is_eligible), 0) * 100,
        2
    ) AS turnout_percent
FROM voter_status
WHERE election_id = $1;
```

---

## 🗳️ 6. Partisipasi per TPS

Jumlah pemilih yang voting via TPS, dengan usage percentage.

```sql
-- Partisipasi per TPS dengan usage percentage
SELECT
    t.id              AS tps_id,
    t.code            AS tps_code,
    t.name            AS tps_name,
    t.capacity_estimate,
    COUNT(vs.id)      AS total_voted_tps,
    ROUND(
        COUNT(vs.id)::NUMERIC / NULLIF(t.capacity_estimate, 0) * 100,
        2
    ) AS usage_percent
FROM tps t
LEFT JOIN voter_status vs
    ON vs.tps_id = t.id
   AND vs.election_id = t.election_id
   AND vs.has_voted = TRUE
   AND vs.voting_method = 'TPS'
WHERE t.election_id = $1
GROUP BY t.id, t.code, t.name, t.capacity_estimate
ORDER BY t.code;
```

---

## 👥 7. Statistik Check-in TPS

### 7.1. Ringkasan per TPS

Status check-in untuk monitoring dashboard TPS.

```sql
-- Ringkasan check-in per TPS untuk dashboard TPS / admin
SELECT
    t.id   AS tps_id,
    t.code AS tps_code,
    t.name AS tps_name,

    COUNT(tc.id)                                         AS total_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'PENDING')    AS pending_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'APPROVED')   AS approved_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'REJECTED')   AS rejected_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'USED')       AS used_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'EXPIRED')    AS expired_checkins

FROM tps t
LEFT JOIN tps_checkins tc
       ON tc.tps_id = t.id
      AND tc.election_id = t.election_id
WHERE t.election_id = $1
GROUP BY t.id, t.code, t.name
ORDER BY t.code;
```

### 7.2. List Antrian PENDING (Panel TPS)

Daftar pemilih yang menunggu approval di TPS tertentu.

```sql
-- Antrian PENDING untuk satu TPS, urut dari paling awal
-- Parameter: $1 = tps_id (bukan election_id)
SELECT
    tc.id                   AS checkin_id,
    tc.scan_at,
    v.id                    AS voter_id,
    v.nim,
    v.name,
    v.faculty_name,
    v.study_program_name,
    v.cohort_year
FROM tps_checkins tc
JOIN voters v
  ON v.id = tc.voter_id
WHERE tc.tps_id = $1
  AND tc.status = 'PENDING'
ORDER BY tc.scan_at ASC
LIMIT 100;
```

---

## 💻 8. Suara Online vs TPS (Ringkasan Global)

Breakdown global votes by channel.

```sql
-- Berapa suara via ONLINE vs TPS untuk satu election
SELECT
    SUM(CASE WHEN channel = 'ONLINE' THEN 1 ELSE 0 END) AS total_online,
    SUM(CASE WHEN channel = 'TPS'    THEN 1 ELSE 0 END) AS total_tps,
    COUNT(*)                                            AS total_votes
FROM votes
WHERE election_id = $1;
```

---

## 📋 9. Dashboard Admin Utama (All-in-One)

Query comprehensive untuk main dashboard admin. Satu query, semua metric penting.

```sql
WITH
vs_agg AS (
    SELECT
        election_id,
        COUNT(*) FILTER (WHERE is_eligible)                    AS total_eligible,
        COUNT(*) FILTER (WHERE is_eligible AND has_voted)      AS total_voted,
        COUNT(*) FILTER (WHERE is_eligible AND NOT has_voted)  AS total_not_voted
    FROM voter_status
    WHERE election_id = $1
    GROUP BY election_id
),
votes_agg AS (
    SELECT
        election_id,
        COUNT(*)                                              AS total_votes,
        COUNT(*) FILTER (WHERE channel = 'ONLINE')             AS total_online,
        COUNT(*) FILTER (WHERE channel = 'TPS')                AS total_tps
    FROM votes
    WHERE election_id = $1
    GROUP BY election_id
),
tps_agg AS (
    SELECT
        election_id,
        COUNT(*)                                             AS total_tps,
        COUNT(*) FILTER (WHERE status = 'ACTIVE')            AS active_tps
    FROM tps
    WHERE election_id = $1
    GROUP BY election_id
)
SELECT
    e.id                  AS election_id,
    e.code                AS election_code,
    e.name                AS election_name,
    e.status              AS election_status,
    e.voting_start_at,
    e.voting_end_at,

    vs.total_eligible,
    vs.total_voted,
    vs.total_not_voted,
    ROUND(
        vs.total_voted::NUMERIC
        / NULLIF(vs.total_eligible, 0) * 100,
        2
    ) AS turnout_percent,

    vagg.total_votes,
    vagg.total_online,
    vagg.total_tps,

    tagg.total_tps,
    tagg.active_tps

FROM elections e
LEFT JOIN vs_agg   vs   ON vs.election_id = e.id
LEFT JOIN votes_agg vagg ON vagg.election_id = e.id
LEFT JOIN tps_agg  tagg ON tagg.election_id = e.id
WHERE e.id = $1;
```

**Output fields:**
- Election metadata (id, code, name, status, dates)
- Voter statistics (eligible, voted, not_voted, turnout_percent)
- Vote statistics (total_votes, online, tps)
- TPS statistics (total_tps, active_tps)

---

## 📈 10. Analytics & Advanced Reporting

### Timeline Votes per Hour

Time-series data untuk grafik line/bar.

**File:** `queries/analytics_01_timeline_votes_per_hour.sql`

Generates hourly buckets dari voting_start_at sampai voting_end_at dengan COALESCE untuk jam yang tidak ada suara.

### Timeline by Channel (ONLINE vs TPS)

Stacked bar chart: ONLINE vs TPS per jam.

**File:** `queries/analytics_02_timeline_votes_by_channel.sql`

### Timeline per Candidate

Multi-line chart dengan satu series per kandidat.

**File:** `queries/analytics_03_timeline_votes_per_candidate.sql`

### Heatmap: Faculty × Candidate

Matrix visualization untuk melihat preferensi fakultas.

**Files:**
- `analytics_04_heatmap_faculty_candidate.sql` - Raw counts
- `analytics_05_heatmap_faculty_candidate_percent.sql` - Percentage within faculty

### Cumulative Turnout Timeline

Grafik kumulatif: "dari jam ke jam berapa % turnout".

**File:** `analytics_06_turnout_cumulative_timeline.sql`

Includes `votes_in_hour`, `cumulative_votes`, dan `cumulative_turnout_percent`.

### Demographic Breakdowns

**By Cohort Year:**
```sql
-- File: analytics_07_votes_by_cohort_candidate.sql
-- Output: votes grouped by cohort_year × candidate
```

**By Prodi (Granular):**
```sql
-- File: analytics_08_votes_by_prodi_candidate.sql
-- Output: faculty + prodi + candidate breakdown
```

### Peak Hours Analysis

Ranking jam tersibuk untuk capacity planning.

**File:** `analytics_09_peak_hours_analysis.sql`

Output: Top 20 hours dengan votes, hour_of_day, day_name, rank.

### Voting Velocity

Statistical analysis: berapa rata-rata gap (menit) antar suara?

**File:** `analytics_10_voting_velocity.sql`

Output: avg, min, max, median, p95 gap_minutes.

---

## 🔍 Bonus Queries

### Cek Duplikasi Vote (Audit)

```sql
-- Cek apakah ada token_hash yang dipakai lebih dari 1x
SELECT
    token_hash,
    COUNT(*) AS usage_count
FROM votes
WHERE election_id = $1
GROUP BY token_hash
HAVING COUNT(*) > 1;
```

### Top 5 TPS Tersibuk

```sql
-- TPS dengan jumlah suara terbanyak
SELECT
    t.code,
    t.name,
    COUNT(v.id) AS total_votes
FROM tps t
LEFT JOIN votes v
    ON v.tps_id = t.id
   AND v.election_id = $1
WHERE t.election_id = $1
GROUP BY t.id, t.code, t.name
ORDER BY total_votes DESC
LIMIT 5;
```

### Voter Eligible tapi Belum Vote

```sql
-- List pemilih yang belum vote untuk reminder campaign
SELECT
    v.id,
    v.nim,
    v.name,
    v.email,
    v.faculty_name
FROM voter_status vs
JOIN voters v ON v.id = vs.voter_id
WHERE vs.election_id = $1
  AND vs.is_eligible = TRUE
  AND vs.has_voted = FALSE
ORDER BY v.faculty_name, v.name
LIMIT 100;
```

---

## 🎯 Tips Penggunaan

1. **Indexing**: Semua query sudah dioptimalkan dengan index yang ada di schema
2. **Parameter Binding**: Selalu pakai prepared statement untuk prevent SQL injection
3. **Pagination**: Tambahkan `LIMIT` dan `OFFSET` untuk large dataset
4. **Real-time**: Query #9 (Dashboard) bisa di-cache 1-5 menit untuk reduce DB load
5. **Export**: Semua query bisa export ke CSV via `COPY TO` atau aplikasi layer

---

## 📦 Implementation Example (Go)

```go
// Query #1: Total votes per candidate
type CandidateVotes struct {
    CandidateID     int64  `db:"candidate_id"`
    CandidateNumber int    `db:"candidate_number"`
    ChairmanName    string `db:"chairman_name"`
    ViceChairmanName string `db:"vice_chairman_name"`
    TotalVotes      int64  `db:"total_votes"`
}

func (r *Repository) GetCandidateVotes(ctx context.Context, electionID int64) ([]CandidateVotes, error) {
    query := `
        SELECT
            c.id AS candidate_id,
            c.candidate_number,
            c.chairman_name,
            c.vice_chairman_name,
            COUNT(v.id) AS total_votes
        FROM candidates c
        LEFT JOIN votes v ON v.candidate_id = c.id AND v.election_id = $1
        WHERE c.election_id = $1
        GROUP BY c.id, c.candidate_number, c.chairman_name, c.vice_chairman_name
        ORDER BY c.candidate_number
    `
    
    var results []CandidateVotes
    err := r.db.SelectContext(ctx, &results, query, electionID)
    return results, err
}
```
