#!/bin/bash

# Pemira API - Authentication System Test Script
# Usage: ./test-auth.sh

API_URL="${API_URL:-http://localhost:8080}"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║          PEMIRA API - Authentication System Test             ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Test 1: Login
echo "📝 Test 1: POST /auth/login"
echo "---"
LOGIN_RESPONSE=$(curl -s -X POST "${API_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "student123",
    "password": "password123"
  }')

echo "$LOGIN_RESPONSE" | jq '.'
echo ""

# Extract access token
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token // empty')

if [ -z "$ACCESS_TOKEN" ]; then
    echo "❌ Login failed! No access token received."
    exit 1
fi

echo "✅ Login successful!"
echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
echo ""

# Test 2: Get Me
echo "📝 Test 2: GET /auth/me"
echo "---"
ME_RESPONSE=$(curl -s -X GET "${API_URL}/auth/me" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

echo "$ME_RESPONSE" | jq '.'
echo ""

USER_ID=$(echo "$ME_RESPONSE" | jq -r '.id // empty')
if [ -n "$USER_ID" ]; then
    echo "✅ Get Me successful!"
    echo "User ID: $USER_ID"
    echo "Username: $(echo "$ME_RESPONSE" | jq -r '.username')"
    echo "Role: $(echo "$ME_RESPONSE" | jq -r '.role')"
else
    echo "❌ Get Me failed!"
fi
echo ""

# Test 3: Refresh Token
echo "📝 Test 3: POST /auth/refresh"
echo "---"
REFRESH_RESPONSE=$(curl -s -X POST "${API_URL}/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"${REFRESH_TOKEN}\"
  }")

echo "$REFRESH_RESPONSE" | jq '.'
echo ""

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token // empty')
if [ -n "$NEW_ACCESS_TOKEN" ]; then
    echo "✅ Refresh successful!"
    echo "New Access Token: ${NEW_ACCESS_TOKEN:0:50}..."
else
    echo "❌ Refresh failed!"
fi
echo ""

# Test 4: Invalid token
echo "📝 Test 4: GET /auth/me (with invalid token)"
echo "---"
INVALID_RESPONSE=$(curl -s -X GET "${API_URL}/auth/me" \
  -H "Authorization: Bearer invalid-token-here")

echo "$INVALID_RESPONSE" | jq '.'
echo ""

ERROR_CODE=$(echo "$INVALID_RESPONSE" | jq -r '.code // empty')
if [ "$ERROR_CODE" = "INVALID_TOKEN" ] || [ "$ERROR_CODE" = "UNAUTHORIZED" ]; then
    echo "✅ Invalid token properly rejected!"
else
    echo "⚠️  Unexpected response for invalid token"
fi
echo ""

# Test 5: Logout
echo "📝 Test 5: POST /auth/logout"
echo "---"
LOGOUT_RESPONSE=$(curl -s -X POST "${API_URL}/auth/logout" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"${REFRESH_TOKEN}\"
  }")

echo "$LOGOUT_RESPONSE" | jq '.'
echo ""

LOGOUT_MESSAGE=$(echo "$LOGOUT_RESPONSE" | jq -r '.message // empty')
if [ -n "$LOGOUT_MESSAGE" ]; then
    echo "✅ Logout successful!"
else
    echo "❌ Logout failed!"
fi
echo ""

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                      Tests Complete! ✅                       ║"
echo "╚══════════════════════════════════════════════════════════════╝"
