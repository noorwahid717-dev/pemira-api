# QUICK START - PEMIRA API

## ✅ DATABASE RESTORED SUCCESSFULLY

Database production telah di-restore ke Supabase baru.

## 🔧 Database Configuration

```
Host: aws-1-ap-southeast-1.pooler.supabase.com
Port: 6543
Database: postgres
User: postgres.xqzfrodnznhjstfstvyz
Schema: myschema
```

**Connection String:**
```
postgresql://postgres.xqzfrodnznhjstfstvyz:AZcIF926bLLeeVRQ@aws-1-ap-southeast-1.pooler.supabase.com:6543/postgres
```

## 📊 Database Status

| Table | Rows |
|-------|------|
| elections | 2 |
| voters | 51 |
| candidates | 0 |
| votes | 0 |
| tps | 1 |
| user_accounts | 42 |

Total tables: 25 tables

## 🚀 Quick Commands

### Verify Database
```bash
make db-verify
```

### Build Application
```bash
go build -o build/pemira-api cmd/api/main.go
```

### Run Application
```bash
# Copy environment file
cp .env.production .env

# Edit JWT_SECRET if needed
nano .env

# Run
./build/pemira-api
```

## 📝 Environment Variables

Create `.env` file with:

```env
APP_ENV=production
HTTP_PORT=8080
DATABASE_URL=postgresql://postgres.xqzfrodnznhjstfstvyz:AZcIF926bLLeeVRQ@aws-1-ap-southeast-1.pooler.supabase.com:6543/postgres
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-min-32-chars
JWT_EXPIRATION=24h
REDIS_URL=
LOG_LEVEL=info
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

## ⚠️ Important Notes

1. **NO GOOSE** - Migration system removed. Database uses direct SQL restore.
2. **Schema: myschema** - All tables are in `myschema` schema, not `public`.
3. **Search Path** - Application automatically sets `search_path` to `myschema,public`.
4. **Backup Location** - `backups/pemira_production_backup_20251209_165647.sql`

## 🔍 Verify Tables

```bash
PGPASSWORD="AZcIF926bLLeeVRQ" psql \
  -h aws-1-ap-southeast-1.pooler.supabase.com \
  -p 6543 \
  -U postgres.xqzfrodnznhjstfstvyz \
  -d postgres \
  -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'myschema' ORDER BY table_name;"
```

## 🎯 Next Steps

1. Update JWT_SECRET in `.env`
2. Configure CORS_ALLOWED_ORIGINS for your frontend domain
3. Build: `go build -o build/pemira-api cmd/api/main.go`
4. Deploy the binary to your server
5. Set environment variables on server
6. Run the application

## 🛠️ Troubleshooting

**Connection Issues:**
- Verify Supabase project is active
- Check connection pooler settings
- Ensure IP is whitelisted (if needed)

**Schema Not Found:**
- Run `make db-verify` to check connection
- Verify `search_path` is set correctly in `pkg/database/postgres.go`

**Missing Data:**
- Restore from backup: `PGPASSWORD="..." psql ... < backups/pemira_cleaned.sql`
- Run migration script: `psql ... -f move_to_myschema.sql`

## 📚 Files Reference

- `restore_db.sh` - Full database restore script
- `move_to_myschema.sql` - Move tables from public to myschema
- `schema.sql` - Schema reference documentation
- `.env.production` - Production environment template
- `backups/pemira_production_backup_20251209_165647.sql` - Original backup
- `backups/pemira_cleaned.sql` - Cleaned backup (ownership fixed)

## ✨ Status: READY FOR DEPLOYMENT