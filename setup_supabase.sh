#!/bin/bash
set -e

# Database URL
DB_URL="postgresql://postgres.xqzfrodnznhjstfstvyz:AZcIF926bLLeeVRQ@aws-1-ap-southeast-1.pooler.supabase.com:6543/postgres"

echo "🚀 Setting up PEMIRA database on Supabase..."

# Set environment variable
export DATABASE_URL="$DB_URL"

echo "✅ Schema 'pemira' already created"

# Set default search_path untuk user postgres
echo "📝 Setting default search_path..."
psql "$DB_URL" -c "ALTER DATABASE postgres SET search_path TO pemira, public;" 2>&1 | grep -v "ALTER DATABASE" || true

echo "🔄 Running migrations..."
goose -dir migrations postgres "$DB_URL" up

echo "📊 Checking tables in pemira schema..."
psql "$DB_URL" -c "SET search_path TO pemira, public; \dt" 2>&1 | head -20

echo ""
echo "✅ Database setup complete!"
echo ""
echo "🔧 Next steps:"
echo "1. Update DATABASE_URL in Leapcell:"
echo "   DATABASE_URL=$DB_URL"
echo ""
echo "2. (Optional) Seed master data:"
echo "   psql \"\$DB_URL\" -c \"SET search_path TO pemira, public;\" < seeds/028_seed_master_tables.sql"
echo ""
echo "3. Restart your Leapcell service"
