# 🚀 Pre-Deployment Checklist

## ✅ Code & Build
- [x] Go version updated to 1.24.0
- [x] Dependencies resolved (validator downgraded to v10.22.1)
- [x] Docker build successful
- [x] .dockerignore configured
- [x] Migrations included in Docker image

## ✅ Database Schema
- [x] `branding_files` table uses `storage_path` column (TEXT)
- [x] `branding_settings` table configured
- [x] Foreign keys properly set up
- [x] All migrations ready to run

## ✅ Configuration Files
- [x] `.env.example` updated with Supabase config
- [x] `Dockerfile` optimized (multi-stage, distroless)
- [x] `DEPLOYMENT.md` guide created

## 📋 Before Deploying to Leapcell

### 1. Database Setup
```bash
# Create PostgreSQL database (Neon/Supabase recommended)
# Connection string format:
postgres://username:password@host:5432/database?sslmode=require

# Run all migrations:
psql "YOUR_DATABASE_URL" < migrations/001_init.sql
psql "YOUR_DATABASE_URL" < migrations/002_*.sql
# ... (all migration files in order)

# Verify tables exist:
psql "YOUR_DATABASE_URL" -c "\dt"
```

### 2. Supabase Storage Setup
- [ ] Create bucket "pemira" (public)
- [ ] Allow MIME types: image/jpeg, image/png
- [ ] File size limit: 10MB
- [ ] Copy Project URL and Service Role Key

### 3. Generate Secrets
```bash
# Generate JWT_SECRET (32+ characters)
openssl rand -base64 32

# Or use:
head -c 32 /dev/urandom | base64
```

### 4. Environment Variables (Set di Leapcell)

**Required:**
```
APP_ENV=production
HTTP_PORT=8080
DATABASE_URL=postgres://...
JWT_SECRET=<generated-secret>
JWT_EXPIRATION=24h
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_SECRET_KEY=<service-role-key>
SUPABASE_MEDIA_BUCKET=pemira
SUPABASE_BRANDING_BUCKET=pemira
CORS_ALLOWED_ORIGINS=https://your-frontend.com
LOG_LEVEL=info
```

**Optional:**
```
REDIS_URL=redis://...
```

### 5. Deploy to Leapcell
- [ ] Push code to GitHub
- [ ] Connect repository to Leapcell
- [ ] Set all environment variables
- [ ] Deploy
- [ ] Check deployment logs

### 6. Post-Deployment Testing
```bash
# Health check
curl https://your-api.leapcell.io/health

# Auth test
curl -X POST https://your-api.leapcell.io/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password123"}'

# Public endpoint test
curl https://your-api.leapcell.io/api/v1/elections
```

### 7. Seed Initial Data
```bash
# Insert admin user, elections, etc.
psql "YOUR_DATABASE_URL" < seed_data.sql
psql "YOUR_DATABASE_URL" < seed_election_schedule.sql
```

### 8. Security Hardening
- [ ] Change default admin password
- [ ] Verify JWT_SECRET is strong
- [ ] Confirm DATABASE_URL uses SSL
- [ ] Test CORS with actual frontend domain
- [ ] Review Supabase bucket policies

## 🔍 Verification Points

### API Endpoints Working:
- [ ] `POST /api/v1/auth/login` - Admin login
- [ ] `GET /api/v1/elections` - List elections
- [ ] `GET /api/v1/elections/:id` - Get election detail
- [ ] `POST /api/v1/admin/elections` - Create election (admin)
- [ ] `POST /api/v1/admin/elections/:id/branding/logo/primary` - Upload logo
- [ ] `GET /api/v1/admin/elections/:id/branding/logo/primary` - Get logo (302 redirect)

### Upload Features:
- [ ] Candidate photo upload to Supabase
- [ ] Branding logo upload (primary/secondary)
- [ ] Files accessible via public URL
- [ ] Redirect (302) working correctly

## 📝 Known Issues & Fixes

### Issue: Logo upload returns 500
**Fix**: Ensure `branding_files` table uses `storage_path` (TEXT), not `data` (BYTEA)

### Issue: CORS error
**Fix**: Set `CORS_ALLOWED_ORIGINS` to exact frontend domain with protocol (https://)

### Issue: Database connection failed
**Fix**: Use `?sslmode=require` in DATABASE_URL for production

### Issue: Supabase upload fails
**Fix**: 
1. Verify bucket is public
2. Check SUPABASE_SECRET_KEY is service role key
3. Ensure bucket name matches env var

## 🎯 Success Criteria

Deployment dianggap sukses jika:
- ✅ API server running tanpa error
- ✅ Database connection established
- ✅ Admin bisa login dan mendapat token
- ✅ Public endpoints bisa diakses
- ✅ Upload foto & logo berfungsi
- ✅ GET logo redirect ke Supabase URL
- ✅ CORS tidak block frontend requests

## 📞 Support Resources

- **Leapcell Docs**: https://docs.leapcell.io
- **Supabase Docs**: https://supabase.com/docs/guides/storage
- **PostgreSQL Connection**: https://www.postgresql.org/docs/current/libpq-connect.html

---

**Last Updated**: $(date)
**Prepared By**: System
**Status**: Ready for Deployment ✅
