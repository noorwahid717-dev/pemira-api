#!/bin/bash
echo "🚀 Starting PEMIRA API..."
echo ""

# Start app in background
export DATABASE_URL="postgresql://postgres.xqzfrodnznhjstfstvyz:AZcIF926bLLeeVRQ@aws-1-ap-southeast-1.pooler.supabase.com:6543/postgres"
export JWT_SECRET="test-secret-key-minimum-32-characters-long-for-production"
export APP_ENV="production"
export HTTP_PORT="8080"

./build/pemira-api &
APP_PID=$!

sleep 3

echo "🌐 Starting ngrok tunnel..."
./ngrok http 8080

# Cleanup
kill $APP_PID
