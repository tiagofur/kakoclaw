#!/bin/bash

# =============================================================================
# Pre-Deployment Check Script for MakoClaw
# =============================================================================
# This script verifies that everything is ready for deployment
#
# Usage: ./pre-deploy-check.sh
# =============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PASS=0
WARN=0
FAIL=0

print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}\n"
}

print_check() {
    echo -n "  $1... "
}

print_pass() {
    echo -e "${GREEN}✓ PASS${NC}"
    ((PASS++))
}

print_warn() {
    echo -e "${YELLOW}⚠ WARNING${NC}: $1"
    ((WARN++))
}

print_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((FAIL++))
}

print_info() {
    echo -e "    ${BLUE}ℹ${NC} $1"
}

# Start checks
echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║         MakoClaw Pre-Deployment Check                     ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# =============================================================================
# 1. Repository Checks
# =============================================================================
print_header "Repository Checks"

print_check "Git repository initialized"
if [ -d .git ]; then
    print_pass
else
    print_fail "Not a git repository"
fi

print_check "Working directory is clean"
if [ -z "$(git status --porcelain)" ]; then
    print_pass
else
    print_warn "You have uncommitted changes"
    git status --short | sed 's/^/    /'
fi

# =============================================================================
# 2. Required Files
# =============================================================================
print_header "Required Files"

required_files=(
    "Dockerfile"
    "docker-compose.yml"
    "docker-compose.prod.yml"
    ".dockerignore"
    ".env.example"
    "deploy.sh"
    "go.mod"
    "go.sum"
)

for file in "${required_files[@]}"; do
    print_check "File exists: $file"
    if [ -f "$file" ]; then
        print_pass
    else
        print_fail "File not found: $file"
    fi
done

# =============================================================================
# 3. Go Environment
# =============================================================================
print_header "Go Environment"

print_check "Go is installed"
if command -v go &> /dev/null; then
    print_pass
    print_info "Version: $(go version | awk '{print $3}')"
else
    print_fail "Go is not installed"
fi

print_check "Go modules are valid"
if go mod verify &> /dev/null; then
    print_pass
else
    print_fail "Go modules verification failed"
fi

print_check "Dependencies are up to date"
if go mod tidy -diff &> /dev/null; then
    print_pass
else
    print_warn "Dependencies may need updating (run: go mod tidy)"
fi

# =============================================================================
# 4. Build Tests
# =============================================================================
print_header "Build Tests"

print_check "Project compiles"
if go build -o /tmp/makoclaw-test ./cmd/makoclaw &> /dev/null; then
    print_pass
    rm -f /tmp/makoclaw-test
else
    print_fail "Compilation failed"
    echo "    Run 'go build ./cmd/makoclaw' for details"
fi

print_check "All tests pass"
if go test ./... &> /tmp/makoclaw-test-output.txt; then
    print_pass
    rm -f /tmp/makoclaw-test-output.txt
else
    print_fail "Some tests failed"
    echo "    Run 'go test ./...' for details"
    if [ -f /tmp/makoclaw-test-output.txt ]; then
        echo "    Last few lines:"
        tail -10 /tmp/makoclaw-test-output.txt | sed 's/^/    /'
        rm -f /tmp/makoclaw-test-output.txt
    fi
fi

# =============================================================================
# 5. Docker Files Validation
# =============================================================================
print_header "Docker Configuration"

print_check "Dockerfile syntax"
if docker build -f Dockerfile --target frontend-builder -t test-frontend . &> /dev/null; then
    print_pass
    docker rmi test-frontend &> /dev/null || true
else
    print_warn "Cannot verify Dockerfile (Docker not available or build failed)"
fi

print_check "docker-compose.yml syntax"
if docker-compose -f docker-compose.yml config &> /dev/null; then
    print_pass
else
    print_warn "Cannot verify docker-compose.yml (docker-compose not available)"
fi

print_check "docker-compose.prod.yml syntax"
if docker-compose -f docker-compose.prod.yml config &> /dev/null; then
    print_pass
else
    print_warn "Cannot verify docker-compose.prod.yml (docker-compose not available)"
fi

print_check ".dockerignore exists and is not empty"
if [ -f .dockerignore ] && [ -s .dockerignore ]; then
    print_pass
    print_info "Lines in .dockerignore: $(wc -l < .dockerignore)"
else
    print_warn ".dockerignore is missing or empty"
fi

# =============================================================================
# 6. Environment Configuration
# =============================================================================
print_header "Environment Configuration"

print_check ".env.example exists"
if [ -f .env.example ]; then
    print_pass
else
    print_fail ".env.example is missing"
fi

print_check ".env.example has required variables"
required_vars=(
    "MAKOCLAW_WEB_PASSWORD"
    "MAKOCLAW_AGENTS_DEFAULTS_PROVIDER"
    "MAKOCLAW_AGENTS_DEFAULTS_MODEL"
)

all_vars_present=true
for var in "${required_vars[@]}"; do
    if ! grep -q "^${var}=" .env.example 2>/dev/null; then
        all_vars_present=false
        print_fail "Missing variable in .env.example: $var"
    fi
done

if $all_vars_present; then
    print_pass
fi

print_check ".env file for deployment"
if [ -f .env ]; then
    print_warn ".env file found in repository - ensure it's in .gitignore"

    # Check if sensitive data is present
    if grep -q "changeme" .env 2>/dev/null; then
        print_warn ".env contains default password - change before deployment!"
    fi
else
    print_info ".env not found (will be created during deployment)"
fi

# =============================================================================
# 7. Security Checks
# =============================================================================
print_header "Security Checks"

print_check ".env is in .gitignore"
if grep -q "^\.env$" .gitignore 2>/dev/null; then
    print_pass
else
    print_fail ".env is not in .gitignore - SECURITY RISK!"
fi

print_check "No hardcoded secrets in code"
secret_patterns=("password=" "api_key=" "token=" "secret=")
found_secrets=false

for pattern in "${secret_patterns[@]}"; do
    if grep -r -i "$pattern" --include="*.go" --include="*.yml" --include="*.yaml" . 2>/dev/null | grep -v ".env" | grep -v "#" | head -1; then
        found_secrets=true
    fi
done

if ! $found_secrets; then
    print_pass
else
    print_warn "Potential hardcoded secrets found - review carefully"
fi

print_check "deploy.sh is executable"
if [ -x deploy.sh ]; then
    print_pass
else
    print_warn "deploy.sh is not executable (run: chmod +x deploy.sh)"
fi

# =============================================================================
# 8. Frontend Build
# =============================================================================
print_header "Frontend Checks"

print_check "Frontend package.json exists"
if [ -f pkg/web/frontend/package.json ]; then
    print_pass
else
    print_fail "pkg/web/frontend/package.json not found"
fi

print_check "Node.js is available (optional)"
if command -v node &> /dev/null; then
    print_pass
    print_info "Version: $(node --version)"
else
    print_warn "Node.js not installed (will be installed in Docker)"
fi

# =============================================================================
# 9. Documentation
# =============================================================================
print_header "Documentation"

docs=(
    "README.md"
    "DEPLOYMENT.md"
    "DEPLOY_CHECKLIST.md"
)

for doc in "${docs[@]}"; do
    print_check "Documentation: $doc"
    if [ -f "$doc" ]; then
        print_pass
    else
        print_warn "Documentation file missing: $doc"
    fi
done

# =============================================================================
# Summary
# =============================================================================
print_header "Summary"

TOTAL=$((PASS + WARN + FAIL))

echo "  Total checks: $TOTAL"
echo -e "  ${GREEN}Passed:${NC} $PASS"
echo -e "  ${YELLOW}Warnings:${NC} $WARN"
echo -e "  ${RED}Failed:${NC} $FAIL"

echo ""

if [ $FAIL -eq 0 ] && [ $WARN -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed! Ready for deployment.${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Commit and push your changes"
    echo "  2. SSH to your VPS"
    echo "  3. Clone the repository"
    echo "  4. Run: ./deploy.sh --install-prereqs --build --start"
    echo ""
    exit 0
elif [ $FAIL -eq 0 ]; then
    echo -e "${YELLOW}⚠ Checks passed with warnings. Review warnings before deployment.${NC}"
    echo ""
    echo "You can proceed with deployment, but address the warnings first if possible."
    echo ""
    exit 0
else
    echo -e "${RED}✗ Some checks failed. Fix the issues before deployment.${NC}"
    echo ""
    echo "Review the failed checks above and fix the issues."
    echo ""
    exit 1
fi
