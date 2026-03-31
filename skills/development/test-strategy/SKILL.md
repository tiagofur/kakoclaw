---
name: test-strategy
title: Testing Strategy Expert
description: Expert guidance on comprehensive testing strategies, frameworks, and best practices
when_to_use: When designing test suites, writing unit/integration/e2e tests, applying TDD, improving test coverage, or choosing a testing framework
version: 1.0.0
author: MakoClaw Development Team
category: testing
tags: [testing, tdd, coverage, quality, testing-frameworks]
---

# Testing Strategy Expert

You are an expert in software testing with deep knowledge of testing strategies, frameworks, and best practices. When developing or reviewing test suites, follow these principles:

## Testing Pyramid

```
        /\
       /  \    E2E Tests (10%)
      /____\   - Critical user paths
     /      \  - Slow, expensive, brittle
    /        \
   /          \ Integration Tests (30%)
  /__________\ - Component interactions
 /            \- Database, API, services
/              \
 Unit Tests (60%)
- Fast, focused, reliable
- Test individual functions/classes
```

## Test Categories

### 1. Unit Tests

**Purpose:** Verify individual units of code in isolation

**Characteristics:**
- Fast (milliseconds)
- Isolated (no external dependencies)
- Deterministic (same result every time)
- Focus on one behavior per test

**Best Practices:**
```go
// Good unit test
func TestCalculatePrice_WithDiscount_ReturnsDiscountedPrice(t *testing.T) {
    // Arrange
    calculator := NewPriceCalculator()
    discount := 0.2 // 20% discount

    // Act
    result := calculator.CalculatePrice(100, discount)

    // Assert
    expected := 80.0
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

**When to Write:**
- Business logic
- Validation rules
- Data transformations
- Edge cases and error conditions

### 2. Integration Tests

**Purpose:** Verify that components work together correctly

**Characteristics:**
- Slower than unit tests
- Test real interactions (database, APIs)
- May use test doubles for external services
- Focus on integration points

**Best Practices:**
```go
func TestUserService_CreateUser_SavesToDatabase(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create service with real DB
    service := NewUserService(db)

    // Act
    user, err := service.CreateUser("test@example.com", "password")

    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if user.ID == 0 {
        t.Error("expected user ID to be set")
    }

    // Verify in database
    saved, err := db.GetUser(user.ID)
    if err != nil {
        t.Fatalf("failed to fetch user: %v", err)
    }
    if saved.Email != "test@example.com" {
        t.Errorf("expected email 'test@example.com', got %v", saved.Email)
    }
}
```

### 3. End-to-End (E2E) Tests

**Purpose:** Verify critical user workflows from start to finish

**Characteristics:**
- Slowest (seconds to minutes)
- Test through the UI or API
- Most brittle, require maintenance
- Test only critical paths

**Best Practices:**
```javascript
// E2E test with Playwright
test('user can complete checkout flow', async ({ page }) => {
    // Navigate to site
    await page.goto('https://shop.example.com')

    // Add item to cart
    await page.click('text=Product A')
    await page.click('text=Add to Cart')

    // View cart
    await page.click('text=Cart')
    await expect(page.locator('.cart-item')).toHaveCount(1)

    // Checkout
    await page.click('text=Checkout')
    await page.fill('[name="email"]', 'test@example.com')
    await page.click('text=Place Order')

    // Verify success
    await expect(page.locator('.order-confirmation')).toBeVisible()
})
```

## Test-Driven Development (TDD)

**The TDD Cycle:**
1. 🔴 **Red**: Write a failing test for new behavior
2. 🟢 **Green**: Write the minimum code to make the test pass
3. 🔵 **Refactor**: Improve the code while tests pass

**Example:**

```go
// Step 1: Write failing test
func TestAdd_TwoNumbers_ReturnsSum(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("expected 5, got %v", result)
    }
}
// Run: FAIL - function doesn't exist

// Step 2: Write minimum code
func Add(a, b int) int {
    return 5 // passes this test only
}
// Run: PASS

// Step 3: Refactor to general solution
func Add(a, b int) int {
    return a + b
}
// Run: PASS
```

## Testing Patterns

### Table-Driven Tests

**When:** Testing same logic with different inputs

```go
func TestValidateEmail_ValidEmails_ReturnsTrue(t *testing.T) {
    tests := []struct {
        name  string
        email string
        want  bool
    }{
        {"valid simple", "user@example.com", true},
        {"valid with dots", "user.name@example.com", true},
        {"valid with plus", "user+tag@example.com", true},
        {"invalid no at", "userexample.com", false},
        {"invalid no domain", "user@", false},
        {"empty", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ValidateEmail(tt.email)
            if got != tt.want {
                t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.want)
            }
        })
    }
}
```

### Test Helpers and Fixtures

**When:** Setting up complex test scenarios

```go
// testutil/helpers.go
type TestUser struct {
    ID    int
    Name  string
    Email string
}

func CreateTestUser(t *testing.T, db *sql.DB) TestUser {
    t.Helper() // Marks this as a test helper

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

    return user
}

// Usage in test
func TestDeleteUser_DeletesFromDatabase(t *testing.T) {
    db := setupTestDB(t)
    user := CreateTestUser(t, db)

    err := DeleteUser(db, user.ID)
    if err != nil {
        t.Fatalf("DeleteUser failed: %v", err)
    }

    // Verify deleted
    var count int
    db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", user.ID).Scan(&count)
    if count != 0 {
        t.Error("user was not deleted")
    }
}
```

### Mocking and Fakes

**When:** Isolating from external dependencies

```go
// Mock implementation
type MockEmailService struct {
    SentTo      []string
    SendError   error
    CallCount   int
}

func (m *MockEmailService) Send(to, subject, body string) error {
    m.CallCount++
    m.SentTo = append(m.SentTo, to)
    return m.SendError
}

// Usage in test
func TestWelcomeEmail_OnRegistration_SendsEmail(t *testing.T) {
    mockEmail := &MockEmailService{}
    userService := NewUserService(mockEmail)

    userService.Register("test@example.com", "password")

    if mockEmail.CallCount != 1 {
        t.Errorf("expected 1 email, got %d", mockEmail.CallCount)
    }
    if len(mockEmail.SentTo) != 1 || mockEmail.SentTo[0] != "test@example.com" {
        t.Error("email not sent to correct address")
    }
}
```

## Test Coverage

**Coverage Goals:**
- **Overall:** >80% code coverage
- **Critical Path:** 100% coverage
- **Error Handling:** 100% coverage
- **Edge Cases:** All boundary conditions

**Measuring Coverage:**

```bash
# Go
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# JavaScript/TypeScript
npm test -- --coverage
npm test -- --coverage --collectCoverageFrom='src/**/*.{js,ts}'

# Python
pytest --cov=src --cov-report=html
```

## Test Organization

**Directory Structure:**
```
project/
├── src/
│   ├── user/
│   │   ├── user.go
│   │   ├── user_test.go        # Unit tests alongside code
│   │   └── mocks.go
│   └── api/
│       ├── handlers.go
│       └── handlers_test.go
├── test/
│   ├── integration/            # Integration tests
│   │   ├── database_test.go
│   │   └── api_test.go
│   ├── e2e/                    # E2E tests
│   │   ├── checkout.spec.ts
│   │   └── auth.spec.ts
│   ├── fixtures/               # Test data
│   │   └── users.json
│   └── testutil/               # Test helpers
│       └── helpers.go
```

## Property-Based Testing

**When:** Testing invariants with random inputs

```go
// Using testing/quick
func TestAppendAndLength(t *testing.T) {
    property := func(xs []int, x int) bool {
        xs = append(xs, x)
        return xs[len(xs)-1] == x && len(xs) > 0
    }

    if err := quick.Check(property, nil); err != nil {
        t.Error(err)
    }
}
```

## Performance Testing

**Benchmarking:**

```go
func BenchmarkFibonacci(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Fibonacci(20)
    }
}

// Run with: go test -bench=.
```

**Load Testing:**

```javascript
// Using k6
import http from 'k6/http';
import { check } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 100 },  // Ramp up
        { duration: '1m', target: 100 },    // Stay at 100
        { duration: '30s', target: 0 },     // Ramp down
    ],
};

export default function () {
    let res = http.get('https://api.example.com/users');
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
    });
}
```

## Continuous Testing

**Pre-commit Hooks:**
```bash
# .git/hooks/pre-commit
#!/bin/bash
go fmt ./...
go vet ./...
go test ./... -race
```

**CI Pipeline:**
```yaml
# GitHub Actions
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go test ./... -race -coverprofile=coverage.out
      - uses: codecov/codecov-action@v2
```

## Testing Anti-Patterns

**❌ Avoid:**

1. **Testing Implementation Details**
```go
// Bad - tests internal structure
if user.internalID != 123 { ... }

// Good - tests observable behavior
if user.GetID() != 123 { ... }
```

2. **Flaky Tests**
```go
// Bad - depends on timing
time.Sleep(100 * time.Millisecond)

// Good - use explicit synchronization
wait := make(chan struct{})
go func() {
    doSomething()
    close(wait)
}()
<-wait
```

3. **Testing Multiple Things**
```go
// Bad - multiple unrelated assertions
func TestUser(t *testing.T) {
    user := CreateUser()
    if user.Name == "" { t.Error() }
    if user.Age < 0 { t.Error() }
    if user.Email == "" { t.Error() }
}

// Good - separate tests
func TestUser_Name_NotEmpty(t *testing.T) { ... }
func TestUser_Age_NonNegative(t *testing.T) { ... }
func TestUser_Email_NotEmpty(t *testing.T) { ... }
```

## Testing Checklist

Before considering a feature complete:

**Unit Tests:**
- [ ] All public functions have tests
- [ ] Edge cases covered
- [ ] Error paths tested
- [ ] No hardcoded values
- [ ] Tests are deterministic

**Integration Tests:**
- [ ] Database interactions tested
- [ ] External API integrations tested
- [ ] Component interactions verified
- [ ] Test data properly cleaned up

**E2E Tests:**
- [ ] Critical user paths covered
- [ ] Tests are reliable
- [ ] Tests run in CI
- [ ] Failure notifications set up

**General:**
- [ ] Coverage measured
- [ ] All tests pass locally
- [ ] All tests pass in CI
- [ ] No race conditions detected
- [ ] Tests run quickly enough

## Framework-Specific Resources

**Go:**
- [testing package](https://golang.org/pkg/testing/)
- [testify assertions](https://github.com/stretchr/testify)
- [gomock](https://github.com/golang/mock)

**JavaScript/TypeScript:**
- [Jest](https://jestjs.io/)
- [Vitest](https://vitest.dev/)
- [Playwright](https://playwright.dev/)
- [MSW (API mocking)](https://mswjs.io/)

**Python:**
- [pytest](https://docs.pytest.org/)
- [unittest.mock](https://docs.python.org/3/library/unittest.mock.html)
- [hypothesis (PBT)](https://hypothesis.works/)
