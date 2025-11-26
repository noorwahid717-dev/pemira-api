# 🚀 Quick Deploy Guide - Leapcell

## Step-by-Step (10 menit)

### 1️⃣ Database Setup (Neon PostgreSQL recommended)

1. Buat account di [Neon](https://neon.tech) atau [Supabase](https://supabase.com)
2. Create database baru
3. Copy connection string:
   ```
   postgres://username:password@host/database?sslmode=require
   ```

4. Run migrations:
   ```bash
   # Download semua file migration dari repo, lalu:
   for f in migrations/*.sql; do
     psql "YOUR_DATABASE_URL" < "$f"
   done
   ```

### 2️⃣ Supabase Storage Setup

1. Login ke [Supabase](https://supabase.com)
2. Go to Storage → Create bucket → Name: `pemira`
3. Settings:
   - ✅ Public bucket
   - File size limit: 10 MB
   - Allowed MIME: image/png, image/jpeg
4. Copy dari Settings → API:
   - **Project URL**: https://xxx.supabase.co
   - **Service role key** (bukan anon key!)

### 3️⃣ Generate JWT Secret

```bash
openssl rand -base64 32
```
Copy output, ini akan jadi `JWT_SECRET`

### 4️⃣ Deploy ke Leapcell

1. Push code ke GitHub:
   ```bash
   git add .
   git commit -m "Ready for production"
   git push origin main
   ```

2. Login ke [Leapcell](https://leapcell.io)

3. **Create New Project** → Connect GitHub repo

4. **Set Environment Variables** (tab Environment):
   ```
   APP_ENV=production
   HTTP_PORT=8080
   DATABASE_URL=<dari step 1>
   JWT_SECRET=<dari step 3>
   JWT_EXPIRATION=24h
   SUPABASE_URL=<dari step 2>
   SUPABASE_SECRET_KEY=<dari step 2>
   SUPABASE_MEDIA_BUCKET=pemira
   SUPABASE_BRANDING_BUCKET=pemira
   CORS_ALLOWED_ORIGINS=https://your-frontend-domain.com
   LOG_LEVEL=info
   ```

5. Click **Deploy** 🚀

### 5️⃣ Verify Deployment

```bash
# Replace dengan URL Leapcell kamu
API_URL="https://your-app.leapcell.io"

# Health check
curl $API_URL/health

# Login test
curl -X POST $API_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password123"}'
```

### 6️⃣ Seed Initial Data

```bash
# Insert admin user dan data awal
psql "YOUR_DATABASE_URL" < seed_data.sql
psql "YOUR_DATABASE_URL" < seed_election_schedule.sql
```

### 7️⃣ Change Default Password

```bash
# Login ke app, lalu ubah password admin default!
```

---

## 📁 Important Files

| File | Purpose |
|------|---------|
| `DEPLOYMENT.md` | Full deployment documentation |
| `DEPLOYMENT_CHECKLIST.md` | Step-by-step checklist |
| `.env.example` | Environment variable template |
| `Dockerfile` | Container configuration |
| `.dockerignore` | Files to exclude from image |

## 🆘 Common Issues

**Database connection failed**
→ Check `sslmode=require` in DATABASE_URL

**Upload returns 500**
→ Verify Supabase bucket is public

**CORS error**
→ Add exact frontend domain to CORS_ALLOWED_ORIGINS (with https://)

**Build fails**
→ Leapcell auto-detects Dockerfile, no custom build command needed

---

## ✅ Success Checklist

- [ ] Database created & migrations ran
- [ ] Supabase bucket "pemira" created (public)
- [ ] All env vars set in Leapcell
- [ ] Deployment successful (green status)
- [ ] Health endpoint returns 200
- [ ] Admin can login
- [ ] Upload logo works
- [ ] Frontend can access API (no CORS errors)

**Total time: ~10-15 minutes** ⏱️

---

Need help? Check:
- 📖 `DEPLOYMENT.md` - Detailed guide
- 📋 `DEPLOYMENT_CHECKLIST.md` - Complete checklist
- 🔧 Leapcell docs: https://docs.leapcell.io
