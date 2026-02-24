#!/bin/bash
# Pre-Production Checklist for MakoClaw
# Run this script before deploying to production

set -e

echo "🦈 MakoClaw Pre-Production Validation"
echo "======================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SUCCESS=0
WARNINGS=0
ERRORS=0

check() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $1"
        ((SUCCESS++))
    else
        echo -e "${RED}✗${NC} $1"
        ((ERRORS++))
    fi
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS++))
}

echo "1️⃣  Testing Go Build"
echo "===================="
go build -o /tmp/makoclaw-test ./cmd/makoclaw > /dev/null 2>&1
check "Go binary compiles successfully"
rm -f /tmp/makoclaw-test

echo ""
echo "2️⃣  Running Unit Tests"
echo "===================="
go test ./... -v > /tmp/makoclaw-tests.log 2>&1
if grep -q "FAIL" /tmp/makoclaw-tests.log; then
    echo -e "${RED}✗${NC} Some tests failed"
    grep "FAIL" /tmp/makoclaw-tests.log
    ((ERRORS++))
else
    check "All unit tests passed"
fi

echo ""
echo "3️⃣  Static Analysis"
echo "===================="
go vet ./... > /dev/null 2>&1
check "go vet passes with no issues"

go fmt ./... > /tmp/makoclaw-fmt.log 2>&1
if [ -s /tmp/makoclaw-fmt.log ]; then
    warn "Some files need formatting (run 'make fmt')"
else
    check "Code is properly formatted"
fi

echo ""
echo "4️⃣  Security Checks"
echo "===================="

# Check for hardcoded credentials
if grep -r "password.*=.*\"" --include="*.go" pkg/ cmd/ | grep -v "Password.*string" | grep -v "test" > /dev/null; then
    warn "Possible hardcoded passwords found (review manually)"
else
    check "No obvious hardcoded passwords"
fi

# Check for SQL injection risks
if grep -r "fmt.Sprintf.*SELECT\|INSERT\|UPDATE\|DELETE" --include="*.go" pkg/ cmd/ | grep -v "_test.go" > /dev/null; then
    warn "Possible SQL injection risk (review SQL query construction)"
else
    check "No obvious SQL injection patterns"
fi

# Check for command injection
if grep -r "exec.Command.*+" --include="*.go" pkg/ cmd/ | grep -v "_test.go" > /dev/null; then
    warn "Command execution with string concat (review for injection)"
else
    check "No obvious command injection patterns"
fi

echo ""
echo "5️⃣  Docker Build"
echo "===================="
docker build -t makoclaw:pre-prod-test . > /tmp/makoclaw-docker.log 2>&1
check "Docker image builds successfully"

echo ""
echo "6️⃣  Configuration Validation"
echo "===================="

# Check if example config exists
if [ -f "config.example.json" ]; then
    check "config.example.json exists"
else
    warn "config.example.json not found"
fi

# Check if docker-compose.yml exists
if [ -f "docker-compose.yml" ]; then
    check "docker-compose.yml exists"
else
    warn "docker-compose.yml not found"
fi

# Check if Dockerfile uses multi-stage build
if grep -q "AS frontend-builder" Dockerfile && grep -q "AS backend-builder" Dockerfile; then
    check "Dockerfile uses multi-stage build"
else
    warn "Dockerfile may not use optimal multi-stage build"
fi

echo ""
echo "7️⃣  Documentation"
echo "===================="

if [ -f "README.md" ]; then
    check "README.md exists"
else
    warn "README.md not found"
fi

if [ -d "docs" ]; then
    check "docs/ directory exists"
else
    warn "docs/ directory not found"
fi

echo ""
echo "8️⃣  Dependencies"
echo "===================="

# Check go.mod
if [ -f "go.mod" ]; then
    check "go.mod exists"
else
    echo -e "${RED}✗${NC} go.mod not found"
    ((ERRORS++))
fi

# Check for known vulnerable packages (example)
if go list -m all | grep -q "golang.org/x/"; then
    check "Using golang.org/x/ packages (ensure up-to-date)"
fi

echo ""
echo "9️⃣  Frontend Build"
echo "===================="

if [ -f "pkg/web/frontend/package.json" ]; then
    check "Frontend package.json exists"
    
    # Check if node_modules exists or can be installed
    if [ -d "pkg/web/frontend/node_modules" ] || (cd pkg/web/frontend && npm install > /dev/null 2>&1); then
        check "Frontend dependencies installable"
        
        # Try building frontend
        if (cd pkg/web/frontend && npm run build > /dev/null 2>&1); then
            check "Frontend builds successfully"
        else
            warn "Frontend build failed (check npm run build)"
        fi
    else
        warn "Frontend dependencies installation failed"
    fi
else
    warn "Frontend package.json not found"
fi

echo ""
echo "🔟 Database Migrations"
echo "===================="

if grep -q "migrate()" pkg/storage/*.go; then
    check "Migration code exists in storage package"
else
    warn "No obvious migration code found"
fi

echo ""
echo "====================================="
echo "📊 SUMMARY"
echo "====================================="
echo -e "${GREEN}Success:${NC} $SUCCESS checks"
echo -e "${YELLOW}Warnings:${NC} $WARNINGS checks"
echo -e "${RED}Errors:${NC} $ERRORS checks"
echo ""

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}❌ PRE-PRODUCTION CHECK FAILED${NC}"
    echo "Fix the errors above before deploying to production."
    exit 1
elif [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}⚠️  PRE-PRODUCTION CHECK PASSED WITH WARNINGS${NC}"
    echo "Review the warnings above before deploying."
    exit 0
else
    echo -e "${GREEN}✅ PRE-PRODUCTION CHECK PASSED${NC}"
    echo "System is ready for production deployment."
    exit 0
fi
