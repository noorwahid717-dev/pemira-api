# 🚀 DEPLOY KE RENDER - 5 MENIT!

## STEP BY STEP:

### 1️⃣ Buka Render
https://render.com

### 2️⃣ Login
- Click "Get Started" atau "Login"
- Login pakai GitHub

### 3️⃣ New Web Service
- Dashboard → Click "New +"
- Pilih "Web Service"

### 4️⃣ Connect Repository
- Connect GitHub account (kalau belum)
- Cari repo: pemira-api
- Click "Connect"

### 5️⃣ Configure (AUTO DETECT!)
Render akan auto detect dari render.yaml:
- Name: pemira-api ✅
- Region: Singapore ✅
- Build Command: go build... ✅
- Start Command: ./build/pemira-api ✅

### 6️⃣ IMPORTANT: Set Environment Variables
Kalau tidak auto detect, set manual:

```
DATABASE_URL = postgresql://postgres.xqzfrodnznhjstfstvyz:AZcIF926bLLeeVRQ@aws-1-ap-southeast-1.pooler.supabase.com:6543/postgres

JWT_SECRET = [GENERATE RANDOM STRING 32+ CHARS]

APP_ENV = production

HTTP_PORT = 8080

LOG_LEVEL = info

CORS_ALLOWED_ORIGINS = https://your-frontend-domain.com
```

### 7️⃣ Deploy!
- Click "Create Web Service"
- Tunggu 3-5 menit build & deploy
- **DONE!**

### 8️⃣ Get Your URL
Render kasih URL seperti:
```
https://pemira-api-xxxx.onrender.com
```

**PAKAI URL INI DI FRONTEND!** ✅ SSL AUTO!

---

## ✅ ADVANTAGES RENDER:

- ✅ Gratis 750 jam/bulan (cukup!)
- ✅ SSL auto
- ✅ No credit card required
- ✅ Auto deploy from GitHub
- ✅ Lebih generous dari Railway

## ⚠️ Note:

Free tier sleep after 15 menit inactive.
First request setelah sleep = cold start ~30 detik.

Upgrade ke paid ($7/month) = always on.

---

## 🔥 ALTERNATIVE: FLY.IO

Kalau mau always-on gratis, pakai Fly.io:

```bash
curl -L https://fly.io/install.sh | sh
fly auth login
fly launch --region sin
fly secrets set DATABASE_URL="..." JWT_SECRET="..."
fly deploy
```

Fly.io gratis: 3 VM always-on!

---

**PILIH RENDER ATAU FLY.IO, DEPLOY SEKARANG!** 🚀
