#!/bin/bash
# Verification script for cancel and background execution feature
# Tests that all components are properly deployed

set -e

BASE_URL="http://127.0.0.1:18880"
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🔍 KakoClaw Cancel & Background Execution Feature Verification"
echo "============================================================="
echo ""

# Test 1: Health check
echo "1. Testing health endpoint..."
HEALTH=$(curl -s "$BASE_URL/api/v1/health")
if echo "$HEALTH" | grep -q "ok"; then
    echo -e "${GREEN}✓${NC} Health endpoint OK"
else
    echo -e "${RED}✗${NC} Health endpoint FAILED"
    exit 1
fi

# Test 2: Web panel loads
echo "2. Testing web panel..."
WEB=$(curl -s "$BASE_URL/" | head -20)
if echo "$WEB" | grep -q "KakoClaw"; then
    echo -e "${GREEN}✓${NC} Web panel loads"
else
    echo -e "${RED}✗${NC} Web panel FAILED to load"
    exit 1
fi

# Test 3: Cancel endpoint exists (should return 401 without auth)
echo "3. Testing cancel endpoint..."
CANCEL=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/api/v1/chat/cancel")
if [ "$CANCEL" = "401" ] || [ "$CANCEL" = "405" ]; then
    echo -e "${GREEN}✓${NC} Cancel endpoint exists (requires auth)"
else
    echo -e "${RED}✗${NC} Cancel endpoint not found (got HTTP $CANCEL)"
    exit 1
fi

# Test 4: Active endpoint exists (should return 401 without auth)
echo "4. Testing active executions endpoint..."
ACTIVE=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/api/v1/chat/active")
if [ "$ACTIVE" = "401" ] || [ "$ACTIVE" = "200" ]; then
    echo -e "${GREEN}✓${NC} Active endpoint exists"
else
    echo -e "${RED}✗${NC} Active endpoint not found (got HTTP $ACTIVE)"
    exit 1
fi

# Test 5: Container running
echo "5. Checking Docker container..."
if docker ps | grep -q kakoclaw; then
    echo -e "${GREEN}✓${NC} Docker container is running"
else
    echo -e "${RED}✗${NC} Docker container not running"
    exit 1
fi

# Test 6: Check for compilation errors in logs
echo "6. Checking container logs for errors..."
LOGS=$(docker logs kakoclaw-kakoclaw-1 --tail 50 2>&1)
if echo "$LOGS" | grep -qi "panic\|fatal\|compilation error"; then
    echo -e "${RED}✗${NC} Found errors in logs:"
    echo "$LOGS" | grep -i "panic\|fatal\|compilation error"
    exit 1
else
    echo -e "${GREEN}✓${NC} No critical errors in logs"
fi

# Test 7: Frontend bundle - check in source (embedded in binary)
echo "7. Checking frontend source code..."
if [ -f "pkg/web/frontend/src/views/ChatView.vue" ]; then
    if grep -q "cancelExecution" pkg/web/frontend/src/views/ChatView.vue; then
        echo -e "${GREEN}✓${NC} Cancel button code found in frontend source"
    else
        echo -e "${RED}✗${NC} Cancel button code NOT found in source"
        exit 1
    fi
else
    echo -e "${RED}✗${NC} Frontend source not found"
    exit 1
fi

# Summary
echo ""
echo "============================================================="
echo -e "${GREEN}✓ All verification checks passed!${NC}"
echo ""
echo "Features verified:"
echo "  • Web server running on port 18880"
echo "  • Cancel endpoint: POST /api/v1/chat/cancel"
echo "  • Active endpoint: GET /api/v1/chat/active"
echo "  • Frontend includes cancel button UI"
echo "  • No critical errors in logs"
echo ""
echo "Next steps:"
echo "  1. Open http://127.0.0.1:18880 in your browser"
echo "  2. Login with your credentials"
echo "  3. Start a long-running task in chat"
echo "  4. Click the cancel button (red X) to test cancellation"
echo "  5. Start a task and navigate to Tasks to test background execution"
echo ""
echo "Documentation:"
echo "  • docs/features/cancel-and-background-execution.md"
echo "  • IMPLEMENTATION_SUMMARY_CANCEL_FEATURE.md"
