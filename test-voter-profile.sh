#!/bin/bash

# Voter Profile API Test Script
# Usage: ./test-voter-profile.sh <voter_token>

set -e

BASE_URL="http://localhost:8080/api/v1"
TOKEN="${1:-}"

if [ -z "$TOKEN" ]; then
    echo "❌ Error: Bearer token required"
    echo "Usage: ./test-voter-profile.sh <voter_token>"
    exit 1
fi

echo "🧪 Testing Voter Profile API Endpoints"
echo "========================================"
echo ""

# Test 1: Get Complete Profile
echo "1️⃣  Testing GET /voters/me/complete-profile"
curl -s -X GET \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/voters/me/complete-profile" | jq '.' || echo "❌ Failed"
echo ""
echo "---"
echo ""

# Test 2: Update Profile
echo "2️⃣  Testing PUT /voters/me/profile"
curl -s -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test.update@uniwa.ac.id",
    "phone": "08123456789"
  }' \
  "$BASE_URL/voters/me/profile" | jq '.' || echo "❌ Failed"
echo ""
echo "---"
echo ""

# Test 3: Get Participation Stats
echo "3️⃣  Testing GET /voters/me/participation-stats"
curl -s -X GET \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/voters/me/participation-stats" | jq '.' || echo "❌ Failed"
echo ""
echo "---"
echo ""

# Test 4: Update Voting Method (will fail if no election)
echo "4️⃣  Testing PUT /voters/me/voting-method"
curl -s -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "election_id": 1,
    "preferred_method": "ONLINE"
  }' \
  "$BASE_URL/voters/me/voting-method" | jq '.' || echo "⚠️  May fail if no election"
echo ""
echo "---"
echo ""

# Test 5: Change Password (use with caution!)
echo "5️⃣  Testing POST /voters/me/change-password (SKIPPED - would change password)"
echo "   To test, use:"
echo "   curl -X POST -H 'Authorization: Bearer \$TOKEN' \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"current_password\":\"old\",\"new_password\":\"new123\",\"confirm_password\":\"new123\"}' \\"
echo "     $BASE_URL/voters/me/change-password"
echo ""
echo "---"
echo ""

# Test 6: Delete Photo
echo "6️⃣  Testing DELETE /voters/me/photo"
curl -s -X DELETE \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/voters/me/photo" | jq '.' || echo "❌ Failed"
echo ""
echo "---"
echo ""

echo "✅ Test suite completed!"
echo ""
echo "📝 Notes:"
echo "  - Make sure the server is running on $BASE_URL"
echo "  - Token must be from a voter account"
echo "  - Some tests may fail if data doesn't exist"
