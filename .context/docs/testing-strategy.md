---
type: doc
name: testing-strategy
description: Test frameworks, patterns, coverage requirements, and quality gates
category: testing
generated: 2026-02-23
status: filled
scaffoldVersion: "2.0.0"
---

# MakoClaw Testing Strategy

This document outlines the comprehensive testing strategy for MakoClaw, including frameworks, patterns, coverage requirements, and quality gates.

## Testing Philosophy

**"Quality is not an act, it is a habit." - Aristotle**

At MakoClaw, we believe in:
- **Testing early and often** - Tests are written alongside code (TDD when appropriate)
- **Testing pyramid** - More unit tests, fewer integration tests, minimal E2E tests
- **Fast feedback** - Tests should run quickly and provide immediate feedback
- **Reliability** - Tests must be deterministic and repeatable
- **Maintainability** - Tests should be as well-written as production code

## Testing Pyramid Distribution

```
        E2E (5%)
       /      \
      /        \
     /          \
    / Integration (25%)
   /            \
  /              \
 /                \
/   Unit Tests (70%) \
```

**Target Distribution:**
- **Unit Tests:** 70% - Fast, isolated, comprehensive
- **Integration Tests:** 25% - Component interactions, database, APIs
- **E2E Tests:** 5% - Critical user journeys only

## Test Coverage Requirements

### Minimum Coverage Standards

| Component Type | Minimum Coverage | Critical Path Coverage |
|---------------|------------------|------------------------|
| Core Business Logic | 90% | 100% |
| API Handlers | 85% | 100% |
| Utilities/Helpers | 90% | 100% |
| Data Models | 80% | 95% |
| UI Components | 70% | 90% |
| Configuration | 60% | 80% |

### Coverage Enforcement

```bash
# Run tests with coverage
go test ./... -coverprofile=coverage.out

# Show coverage by function
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Check coverage threshold (requires 80%)
go test ./... -coverprofile=coverage.out && \
  go tool cover -func=coverage.out | grep total | \
  awk '{if ($3+0 < 80.0) {print "Coverage below 80%"; exit 1}}'
```

## Test Framework Stack

### Go Backend Tests

**Primary Framework:** `testing` package (Go standard library)

**Helper Libraries:**
```go
import (
    "testing"           // Standard testing
    "github.com/stretchr/testify/assert"  // Assertions
    "github.com/stretchr/testify/mock"    // Mocking
    "github.com/stretchr/testify/suite"  // Test suites
)
```

**Example Test Structure:**
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"invalid email", "not-an-email", true},
        {"empty email", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db := setupTestDB(t)
            defer cleanupTestDB(t, db)

            service := NewUserService(db)
            user, err := service.CreateUser(tt.email, "password")

            if (err != nil) != tt.wantErr {
                t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr && user.Email != tt.email {
                t.Errorf("CreateUser() email = %v, want %v", user.Email, tt.email)
            }
        })
    }
}
```

### Frontend Tests

**Framework:** Vitest (Fast, native ESM support)

**Component Testing:** Vue Test Utils

**E2E Testing:** Playwright

**Example Test:**
```javascript
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import LoginForm from '@/components/Auth/LoginForm.vue'

describe('LoginForm', () => {
  it('shows validation errors for empty email', async () => {
    const wrapper = mount(LoginForm)
    await wrapper.find('input[type="email"]').setValue('')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.find('.error').text()).toContain('Email is required')
  })

  it('emits submit event with credentials', async () => {
    const wrapper = mount(LoginForm)
    await wrapper.find('input[type="email"]').setValue('test@example.com')
    await wrapper.find('input[type="password"]').setValue('password123')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')[0]).toEqual([{
      email: 'test@example.com',
      password: 'password123'
    }])
  })
})
```

### Integration Tests

**Database Integration:**
```go
func TestAPIEndpoints(t *testing.T) {
    // Set up test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Set up test server
    server := NewTestServer(db)
    defer server.Close()

    // Test endpoint
    resp := server.Post("/api/users", map[string]string{
        "email": "test@example.com",
        "password": "password123",
    })

    assert.Equal(t, 201, resp.StatusCode)
}
```

### E2E Tests

**Framework:** Playwright

```javascript
import { test, expect } from '@playwright/test'

test.describe('Authentication Flow', () => {
  test('user can log in and view dashboard', async ({ page }) => {
    await page.goto('http://localhost:18880')

    // Navigate to login
    await page.click('text=Login')

    // Fill credentials
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'admin123')
    await page.click('button[type="submit"]')

    // Verify dashboard
    await expect(page).toHaveURL(/.*dashboard/)
    await expect(page.locator('h1')).toContainText('Dashboard')
  })
})
```

## Testing Patterns

### 1. Table-Driven Tests

**When:** Testing same logic with multiple inputs

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        email string
        valid bool
    }{
        {"user@example.com", true},
        {"user.name@example.com", true},
        {"user+tag@example.com", true},
        {"invalid", false},
        {"@example.com", false},
        {"", false},
    }

    for _, tt := range tests {
        t.Run(tt.email, func(t *testing.T) {
            got := ValidateEmail(tt.email)
            if got != tt.valid {
                t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.valid)
            }
        })
    }
}
```

### 2. Test Helpers and Fixtures

**Setup Test Database:**
```go
// testutil/db.go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to open test database: %v", err)
    }

    // Run migrations
    if err := runMigrations(db); err != nil {
        t.Fatalf("failed to run migrations: %v", err)
    }

    return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
    t.Helper()
    if err := db.Close(); err != nil {
        t.Errorf("failed to close database: %v", err)
    }
}
```

**Create Test Data:**
```go
// testutil/fixtures.go
type TestUser struct {
    ID    int
    Name  string
    Email string
}

func CreateTestUser(t *testing.T, db *sql.DB) TestUser {
    t.Helper()

    user := TestUser{
        Name:  "Test User",
        Email: "test@example.com",
    }

    err := db.QueryRow(
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        user.Name, user.Email,
    ).Scan(&user.ID)

    if err != nil {
        t.Fatalf("failed to create test user: %v", err)
    }

    t.Cleanup(func() {
        _, _ = db.Exec("DELETE FROM users WHERE id = $1", user.ID)
    })

    return user
}
```

### 3. Mocking External Dependencies

```go
// mock_llm_provider.go
type MockLLMProvider struct {
    mock.Mock
}

func (m *MockLLMProvider) Complete(ctx context.Context, prompt string) (string, error) {
    args := m.Called(ctx, prompt)
    return args.String(0), args.Error(1)
}

// Usage in test
func TestAgent_ProcessMessage(t *testing.T) {
    mockProvider := new(MockLLMProvider)
    mockProvider.On("Complete", mock.Anything, "hello").
        Return("Hi there!", nil)

    agent := NewAgent(mockProvider)
    response, err := agent.ProcessMessage(context.Background(), "hello")

    assert.NoError(t, err)
    assert.Equal(t, "Hi there!", response)
    mockProvider.AssertExpectations(t)
}
```

### 4. Golden File Testing

**For reproducible outputs:**
```go
func TestRenderTemplate(t *testing.T) {
    data := TemplateData{
        Title: "Test Page",
        Items: []string{"Item 1", "Item 2"},
    }

    got := renderTemplate(data)

    golden, err := os.ReadFile("testdata/template.golden")
    if err != nil {
        t.Fatalf("failed to read golden file: %v", err)
    }

    if got != string(golden) {
        t.Errorf("renderTemplate() mismatch")
        // Update golden file with: go test -update-golden
    }
}
```

## Test Organization

### Directory Structure

```
pkg/
├── config/
│   ├── config.go
│   └── config_test.go           # Unit tests alongside code
├── web/
│   ├── handlers.go
│   └── handlers_test.go
└── storage/
    ├── sqlite.go
    └── sqlite_test.go

test/
├── integration/                 # Integration tests
│   ├── api_test.go
│   └── database_test.go
├── e2e/                         # E2E tests
│   ├── auth_flow_test.go
│   └── task_management_test.go
├── fixtures/                    # Test data
│   ├── users.json
│   └── tasks.json
└── testutil/                    # Test helpers
    ├── db.go
    ├── auth.go
    └── fixtures.go

pkg/web/frontend/
├── src/
│   └── components/
│       ├── LoginForm.vue
│       └── LoginForm.spec.ts    # Component tests
└── e2e/                         # E2E tests
    ├── auth.spec.ts
    └── tasks.spec.ts
```

## Quality Gates

### Pre-commit Checks

**Install pre-commit hooks:**
```bash
# .git/hooks/pre-commit
#!/bin/bash

# Format code
echo "Formatting code..."
go fmt ./...

# Run linter
echo "Running linter..."
golangci-lint run

# Run tests
echo "Running tests..."
go test ./... -race -cover

# Check coverage
echo "Checking coverage..."
go test ./... -coverprofile=coverage.out
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage $COVERAGE% is below 80%"
    exit 1
fi

echo "All checks passed!"
```

### CI Pipeline

**GitHub Actions:**
```yaml
name: Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: makoclaw_test
          POSTGRES_USER: makoclaw
          POSTGRES_PASSWORD: makoclaw
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install dependencies
        run: go mod download

      - name: Run tests
        run: go test ./... -race -coverprofile=coverage.out -covermode=atomic

      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage below 80%"
            exit 1
          fi

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

      - name: Run linter
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  frontend-test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-node@v3
        with:
          node-version: '20'

      - name: Install dependencies
        working-directory: ./pkg/web/frontend
        run: npm ci

      - name: Run tests
        working-directory: ./pkg/web/frontend
        run: npm test -- --coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./pkg/web/frontend/coverage/lcov.info

  e2e-test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Build MakoClaw
        run: make build

      - name: Start MakoClaw
        run: ./makoclaw web &

      - uses: actions/setup-node@v3
        with:
          node-version: '20'

      - name: Install Playwright
        working-directory: ./pkg/web/frontend
        run: npx playwright install

      - name: Run E2E tests
        working-directory: ./pkg/web/frontend
        run: npm run test:e2e
```

## Performance Testing

### Benchmarking

**Go Benchmarks:**
```go
func BenchmarkAgent_ProcessMessage(b *testing.B) {
    agent := setupTestAgent(b)
    message := "test message"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = agent.ProcessMessage(context.Background(), message)
    }
}

// Run with: go test -bench=. -benchmem
```

**Load Testing:**
```javascript
// k6 load test
import http from 'k6/http';
import { check } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '30s', target: 0 },
    ],
};

export default function () {
    let res = http.post('http://localhost:18880/api/messages', {
        message: 'test message',
        channel: 'test',
    });

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
    });
}
```

## Continuous Testing Strategy

### Development Workflow

1. **Write Test First (TDD)**
   - Red: Write failing test
   - Green: Write minimum code to pass
   - Refactor: Improve code while tests pass

2. **Run Tests Frequently**
   - Run unit tests after each change
   - Run integration tests before committing
   - Run full suite before pushing

3. **Monitor Coverage**
   - Check coverage after new features
   - Address coverage gaps immediately
   - Aim for continuous improvement

### Test Metrics to Track

- **Coverage percentage** - Overall and per module
- **Test execution time** - Keep tests fast
- **Flaky test rate** - Should be near zero
- **Test pass rate** - Should be >95%
- **Time to feedback** - How quickly tests run

## Best Practices

### DO ✅

- **Test behavior, not implementation**
- **Use descriptive test names**
- **Test one thing per test**
- **Use setup/teardown for common code**
- **Mock external dependencies**
- **Test edge cases and error paths**
- **Keep tests fast and simple**
- **Make tests deterministic**
- **Use test fixtures for complex data**

### DON'T ❌

- **Don't test implementation details**
- **Don't write brittle tests**
- **Don't ignore test failures**
- **Don't write tests that depend on each other**
- **Don't use sleep/wait for timing**
- **Don't hardcode values in tests**
- **Don't write tests that are too complex**
- **Don't skip tests without fixing them**
- **Don't test third-party libraries**

## Resources

- [Go Testing Guide](https://golang.org/doc/tutorial/add-a-test)
- [Testify Assertions](https://github.com/stretchr/testify/tree/master/assert)
- [Vitest Documentation](https://vitest.dev/)
- [Playwright Documentation](https://playwright.dev/)
- [Testing Anti-Patterns](https://testing.googleblog.com/2015/01/testing-on-toilet-dont-put-everything.html)

## Review Checklist

Before considering testing complete:

**Test Coverage:**
- [ ] All new code has tests
- [ ] Coverage thresholds met
- [ ] Edge cases covered
- [ ] Error paths tested

**Test Quality:**
- [ ] Tests are readable and maintainable
- [ ] Tests are deterministic
- [ ] No flaky tests
- [ ] Tests run quickly

**CI/CD:**
- [ ] All tests pass in CI
- [ ] Coverage reports generated
- [ ] No regressions introduced
- [ ] E2E tests passing

**Documentation:**
- [ ] Complex test logic commented
- [ ] Test fixtures documented
- [ ] Testing guidelines followed
