# Deployment Guide - Leapcell

## Prerequisites

1. Account Leapcell aktif
2. Database PostgreSQL (Neon, Supabase, atau lainnya)
3. Supabase Storage untuk media files
4. Git repository pushed ke GitHub/GitLab

## Environment Variables yang Diperlukan

Set environment variables berikut di Leapcell dashboard:

```bash
# Application
APP_ENV=production
HTTP_PORT=8080

# Database (Required)
DATABASE_URL=postgres://username:password@host:5432/database?sslmode=require

# Auth (Required)
JWT_SECRET=<generate-random-secret-key>
JWT_EXPIRATION=24h

# Logging
LOG_LEVEL=info

# CORS (Required - sesuaikan dengan frontend domain)
CORS_ALLOWED_ORIGINS=https://your-frontend-domain.com,https://www.your-frontend-domain.com

# Supabase Storage (Required untuk upload foto kandidat & branding logo)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SECRET_KEY=your-supabase-service-role-key
SUPABASE_MEDIA_BUCKET=pemira
SUPABASE_BRANDING_BUCKET=pemira

# Redis (Optional - jika menggunakan Redis untuk caching)
# REDIS_URL=redis://default:password@host:6379
```

## Steps Deployment ke Leapcell

### 1. Prepare Repository

```bash
# Ensure all changes are committed
git add .
git commit -m "Prepare for Leapcell deployment"
git push origin main
```

### 2. Setup Database

Jika menggunakan **Neon PostgreSQL** atau **Supabase PostgreSQL**:

1. Create database baru
2. Copy connection string
3. Run migrations:

```bash
# Manually run migrations via psql
psql "postgres://username:password@host:5432/database" < migrations/001_init.sql
psql "postgres://username:password@host:5432/database" < migrations/002_*.sql
# ... dst untuk semua migration files
```

### 3. Setup Supabase Storage

1. Login ke Supabase dashboard
2. Buat bucket "pemira" dengan konfigurasi:
   - **Public bucket**: Yes (agar files bisa diakses public)
   - **File size limit**: 10MB
   - **Allowed MIME types**: image/jpeg, image/png
3. Copy:
   - Project URL (e.g., https://xxx.supabase.co)
   - Service role key (dari Settings > API)

### 4. Deploy ke Leapcell

#### Via Leapcell Dashboard:

1. Login ke [Leapcell](https://leapcell.io)
2. Create New Project:
   - **Name**: pemira-api
   - **Region**: Pilih yang terdekat dengan users
   - **Source**: Connect GitHub repository
3. Configure Build:
   - **Build Command**: (auto-detected from Dockerfile)
   - **Port**: 8080
4. Add Environment Variables (semua yang listed di atas)
5. Click "Deploy"

#### Build Verification:

Dockerfile sudah dikonfigurasi untuk:
- ✅ Multi-stage build (golang:1.22-alpine → distroless)
- ✅ Static binary dengan CGO disabled
- ✅ Includes migrations folder
- ✅ Runs as non-root user
- ✅ Exposes port 8080

### 5. Post-Deployment Checks

```bash
# Check health
curl https://your-api.leapcell.io/health

# Test login
curl -X POST https://your-api.leapcell.io/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password123"}'

# Test public endpoint
curl https://your-api.leapcell.io/api/v1/elections
```

### 6. Initial Data Setup

Setelah deployment, insert initial data:

```bash
# Run seed scripts if needed
psql "YOUR_DATABASE_URL" < seed_data.sql
psql "YOUR_DATABASE_URL" < seed_election_schedule.sql
# etc.
```

## Security Checklist

- [ ] `JWT_SECRET` menggunakan random string minimal 32 characters
- [ ] `DATABASE_URL` menggunakan SSL (`sslmode=require`)
- [ ] `SUPABASE_SECRET_KEY` adalah service role key (bukan anon key)
- [ ] `CORS_ALLOWED_ORIGINS` hanya include domain frontend yang valid
- [ ] Default admin password sudah diubah
- [ ] APP_ENV set ke `production`

## Monitoring & Logs

Leapcell menyediakan:
- Real-time logs di dashboard
- Metrics (CPU, Memory, Network)
- Auto-scaling berdasarkan load

## Troubleshooting

### Database Connection Failed
- Verify DATABASE_URL format
- Ensure database allows connections from Leapcell IPs
- Check SSL mode (`sslmode=require` for production)

### CORS Errors
- Update CORS_ALLOWED_ORIGINS dengan frontend domain yang tepat
- Include protocol (https://)
- Tambahkan subdomain www jika perlu

### Upload Fails
- Verify Supabase credentials
- Check bucket permissions (must be public)
- Ensure bucket name matches env var

### Migration Issues
- Migrations must be run manually before first deployment
- Check migrations folder is included in Docker image

## Rollback

Jika ada issue:
1. Di Leapcell dashboard, pilih deployment sebelumnya
2. Click "Redeploy"

## Support

- Leapcell Docs: https://docs.leapcell.io
- Supabase Docs: https://supabase.com/docs
