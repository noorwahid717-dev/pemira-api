#!/bin/bash

set -e

echo "=== RESTORING PEMIRA PRODUCTION DATABASE ==="

DB_URL="postgresql://postgres.xqzfrodnznhjstfstvyz:AZcIF926bLLeeVRQ@aws-1-ap-southeast-1.pooler.supabase.com:6543/postgres"
BACKUP_FILE="backups/pemira_production_backup_20251209_165647.sql"

echo "Checking backup file..."
if [ ! -f "$BACKUP_FILE" ]; then
    echo "ERROR: Backup file not found: $BACKUP_FILE"
    exit 1
fi

echo "Connecting to database..."
echo "Dropping old schema if exists..."
PGPASSWORD="AZcIF926bLLeeVRQ" psql -h aws-1-ap-southeast-1.pooler.supabase.com -p 6543 -U postgres.xqzfrodnznhjstfstvyz -d postgres << EOF
DROP SCHEMA IF EXISTS myschema CASCADE;
DROP SCHEMA IF EXISTS pemira CASCADE;
EOF

echo "Restoring from backup..."
PGPASSWORD="AZcIF926bLLeeVRQ" psql -h aws-1-ap-southeast-1.pooler.supabase.com -p 6543 -U postgres.xqzfrodnznhjstfstvyz -d postgres < "$BACKUP_FILE"

echo "Verifying restoration..."
PGPASSWORD="AZcIF926bLLeeVRQ" psql -h aws-1-ap-southeast-1.pooler.supabase.com -p 6543 -U postgres.xqzfrodnznhjstfstvyz -d postgres << EOF
\dt myschema.*
SELECT 'Elections:', COUNT(*) FROM myschema.elections;
SELECT 'Voters:', COUNT(*) FROM myschema.voters;
SELECT 'Candidates:', COUNT(*) FROM myschema.candidates;
SELECT 'Votes:', COUNT(*) FROM myschema.votes;
EOF

echo ""
echo "=== RESTORE COMPLETE ==="
echo "Database URL: $DB_URL"
echo "Schema: myschema"
