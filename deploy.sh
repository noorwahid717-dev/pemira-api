#!/bin/bash

set -e

echo "=========================================="
echo "  PEMIRA API - DEPLOYMENT SCRIPT"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${RED}ERROR: .env file not found!${NC}"
    echo "Please create .env file from .env.production template"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

echo -e "${YELLOW}Step 1:${NC} Checking database connection..."
PGPASSWORD="AZcIF926bLLeeVRQ" psql \
  -h aws-1-ap-southeast-1.pooler.supabase.com \
  -p 6543 \
  -U postgres.xqzfrodnznhjstfstvyz \
  -d postgres \
  -c "SELECT COUNT(*) as tables FROM information_schema.tables WHERE table_schema = 'myschema';" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database connection OK${NC}"
else
    echo -e "${RED}✗ Database connection FAILED${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 2:${NC} Running tests..."
go test ./... -short
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Tests passed${NC}"
else
    echo -e "${RED}✗ Tests FAILED${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 3:${NC} Building application..."
go build -ldflags="-s -w" -o build/pemira-api cmd/api/main.go
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build FAILED${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 4:${NC} Verifying binary..."
if [ -f build/pemira-api ]; then
    SIZE=$(du -h build/pemira-api | cut -f1)
    echo -e "${GREEN}✓ Binary created (Size: $SIZE)${NC}"
else
    echo -e "${RED}✗ Binary not found${NC}"
    exit 1
fi

echo ""
echo "=========================================="
echo -e "${GREEN}  DEPLOYMENT READY!${NC}"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Copy binary to server: scp build/pemira-api user@server:/path/"
echo "  2. Copy .env file: scp .env user@server:/path/"
echo "  3. Run on server: ./pemira-api"
echo ""
echo "Or run locally:"
echo "  ./build/pemira-api"
echo ""
echo "API will be available at: http://localhost:${HTTP_PORT:-8080}"
echo ""
