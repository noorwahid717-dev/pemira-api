# LEAPCELL SSL FIX - TROUBLESHOOTING GUIDE

## 🚨 MASALAH: ERR_SSL_PROTOCOL_ERROR

Leapcell force redirect HTTP → HTTPS tapi SSL certificate rusak/belum ready!

---

## ⚡ SOLUSI 1: CONTACT LEAPCELL SUPPORT (FASTEST)

1. Buka Leapcell Dashboard
2. Cari menu **Support** atau **Help**
3. Report issue:
   ```
   Subject: SSL Certificate Error - ERR_SSL_PROTOCOL_ERROR
   
   App URL: https://pemira-api-noorwahid717-dev9346-hfvko405.apn.leapcell.dev
   
   Issue: SSL handshake failing. Getting TLSv1 alert internal error.
   HTTP redirects to HTTPS but HTTPS connection fails.
   
   Please re-provision SSL certificate or disable force HTTPS redirect.
   ```

---

## ⚡ SOLUSI 2: RE-DEPLOY APP

Kadang SSL certificate tidak ter-provision otomatis saat first deploy.

### Steps:
1. Buka Leapcell Dashboard
2. Ke project pemira-api
3. Click **Settings** atau **Deployment**
4. Trigger **Re-deploy** atau **Restart**
5. Tunggu 5-10 menit untuk SSL provisioning
6. Test lagi: `curl https://your-app.apn.leapcell.dev/health`

---

## ⚡ SOLUSI 3: CHECK ENVIRONMENT VARIABLES

Pastikan env vars sudah set di Leapcell:

```
DATABASE_URL=postgresql://postgres.xqzfrodnznhjstfstvyz:AZcIF926bLLeeVRQ@aws-1-ap-southeast-1.pooler.supabase.com:6543/postgres
JWT_SECRET=your-secure-secret-min-32-chars
APP_ENV=production
HTTP_PORT=8080
LOG_LEVEL=info
```

Jika env vars belum set → App crash → SSL tidak ter-provision!

---

## ⚡ SOLUSI 4: CHECK APP LOGS

Di Leapcell Dashboard → Logs, cari error:

### Kalau ada error seperti ini:
```
failed to connect to database
panic: JWT_SECRET is required
```

→ Fix environment variables dulu!

### Kalau app running normal:
```
{"level":"INFO","msg":"connected to database"}
{"level":"INFO","msg":"starting server","port":"8080"}
```

→ SSL issue di Leapcell side, contact support!

---

## ⚡ SOLUSI 5: DEPLOY KE PLATFORM LAIN (RECOMMENDED!)

Kalau Leapcell susah, deploy ke platform lain dengan SSL auto:

### A. RAILWAY (GRATIS, SSL AUTO)
1. https://railway.app
2. Login with GitHub
3. New Project → Deploy from GitHub
4. Pilih repo pemira-api
5. Set environment variables
6. **DONE! SSL auto dalam 2 menit!**

### B. RENDER (GRATIS, SSL AUTO)
1. https://render.com
2. Login with GitHub
3. New → Web Service
4. Connect repo pemira-api
5. Use `render.yaml` config yang sudah ada
6. **DONE! SSL auto dalam 5 menit!**

### C. FLY.IO (GRATIS, SSL AUTO)
1. Install CLI: `curl -L https://fly.io/install.sh | sh`
2. `fly auth login`
3. `fly launch` (pilih region Singapore)
4. Set secrets: `fly secrets set DATABASE_URL=... JWT_SECRET=...`
5. `fly deploy`
6. **DONE! SSL auto!**

---

## 🔍 VERIFY SSL WORKS

```bash
# Test HTTPS
curl -v https://your-app-url.com/health

# Should return:
# < HTTP/2 200
# {"status":"ok"}

# NOT:
# TLS connect error
# ERR_SSL_PROTOCOL_ERROR
```

---

## 📊 DEBUGGING COMMANDS

```bash
# Check SSL certificate
openssl s_client -connect your-app.apn.leapcell.dev:443 -servername your-app.apn.leapcell.dev

# Check DNS
dig your-app.apn.leapcell.dev

# Check HTTP redirect
curl -I http://your-app.apn.leapcell.dev/health

# Test endpoint
curl https://your-app.apn.leapcell.dev/health
```

---

## 🎯 KESIMPULAN

**Leapcell SSL issue kemungkinan:**
1. ❌ Certificate belum provisioned (tunggu/redeploy)
2. ❌ App crash → SSL tidak ter-setup (fix env vars)
3. ❌ Platform bug (contact support)

**SOLUSI TERCEPAT:**
→ **Deploy ke Railway/Render** (5 menit, SSL auto, GRATIS!)

---

## 🚀 NEXT STEPS

1. ✅ Code sudah ready (health checks added)
2. ✅ Database sudah ready (myschema, 25 tables)
3. ✅ Binary sudah built (build/pemira-api)
4. ⏳ Deploy ke platform dengan SSL working

**CHOOSE ONE:**
- Fix Leapcell (contact support, tunggu)
- Deploy Railway (5 menit, recommended!)
- Deploy Render (5 menit, recommended!)

---

**STATUS: READY TO DEPLOY TO RAILWAY/RENDER! 🚀**