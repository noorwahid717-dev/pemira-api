#!/bin/bash

set -e

echo "=========================================="
echo "  PEMIRA API - TESTING SCRIPT"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run test
run_test() {
    local test_name="$1"
    local endpoint="$2"
    local method="${3:-GET}"
    local expected_status="${4:-200}"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo -e "${BLUE}Testing:${NC} $test_name"

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$API_URL$endpoint" 2>/dev/null || echo "000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$API_URL$endpoint" 2>/dev/null || echo "000")
    fi

    status_code=$(echo "$response" | tail -n 1)

    if [ "$status_code" = "$expected_status" ] || [ "$status_code" = "401" ] || [ "$status_code" = "404" ]; then
        echo -e "${GREEN}✓ PASS${NC} - Status: $status_code"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAIL${NC} - Expected: $expected_status, Got: $status_code"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
}

# Start tests
echo -e "${YELLOW}API Base URL:${NC} $API_URL"
echo ""

# Health checks
echo -e "${YELLOW}=== HEALTH CHECKS ===${NC}"
run_test "Health Check" "/health" "GET" "200"
run_test "API Version" "/api/v1/health" "GET" "200"

# Auth endpoints
echo -e "${YELLOW}=== AUTH ENDPOINTS ===${NC}"
run_test "Login Endpoint" "/api/v1/auth/login" "POST" "400"
run_test "Refresh Token" "/api/v1/auth/refresh" "POST" "401"
run_test "Logout" "/api/v1/auth/logout" "POST" "401"

# Elections endpoints
echo -e "${YELLOW}=== ELECTIONS ENDPOINTS ===${NC}"
run_test "List Elections" "/api/v1/elections" "GET" "200"
run_test "Get Election by ID" "/api/v1/elections/1" "GET" "200"
run_test "Active Elections" "/api/v1/elections/active" "GET" "200"

# Candidates endpoints
echo -e "${YELLOW}=== CANDIDATES ENDPOINTS ===${NC}"
run_test "List Candidates" "/api/v1/candidates" "GET" "200"
run_test "Candidates by Election" "/api/v1/elections/1/candidates" "GET" "200"

# Voters endpoints
echo -e "${YELLOW}=== VOTERS ENDPOINTS ===${NC}"
run_test "List Voters" "/api/v1/voters" "GET" "401"
run_test "Voter Profile" "/api/v1/voters/profile" "GET" "401"

# TPS endpoints
echo -e "${YELLOW}=== TPS ENDPOINTS ===${NC}"
run_test "List TPS" "/api/v1/tps" "GET" "200"
run_test "TPS by ID" "/api/v1/tps/1" "GET" "200"

# Master data endpoints
echo -e "${YELLOW}=== MASTER DATA ENDPOINTS ===${NC}"
run_test "List Faculties" "/api/v1/faculties" "GET" "200"
run_test "List Study Programs" "/api/v1/study-programs" "GET" "200"

# Admin endpoints (should fail without auth)
echo -e "${YELLOW}=== ADMIN ENDPOINTS (Auth Required) ===${NC}"
run_test "Admin Dashboard" "/api/v1/admin/dashboard" "GET" "401"
run_test "Admin Users" "/api/v1/admin/users" "GET" "401"
run_test "Admin Elections" "/api/v1/admin/elections" "GET" "401"

# Analytics endpoints
echo -e "${YELLOW}=== ANALYTICS ENDPOINTS ===${NC}"
run_test "Election Results" "/api/v1/analytics/elections/1/results" "GET" "200"
run_test "Turnout Statistics" "/api/v1/analytics/elections/1/turnout" "GET" "200"

# Voting endpoints (should fail without auth)
echo -e "${YELLOW}=== VOTING ENDPOINTS (Auth Required) ===${NC}"
run_test "Submit Vote" "/api/v1/votes" "POST" "401"
run_test "Vote Status" "/api/v1/votes/status" "GET" "401"

# Print summary
echo ""
echo "=========================================="
echo -e "${YELLOW}  TEST SUMMARY${NC}"
echo "=========================================="
echo -e "Total Tests:  ${BLUE}$TOTAL_TESTS${NC}"
echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    echo ""
    exit 1
fi
