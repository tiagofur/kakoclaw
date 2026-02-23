---
name: code-review-checklist
title: Code Review Checklist
description: Comprehensive checklist for thorough and effective code reviews
version: 1.0.0
author: MakoClaw Development Team
category: code-review
tags: [code-review, quality, best-practices, checklist]
---

# Code Review Checklist

You are conducting a thorough code review. Use this comprehensive checklist to ensure high-quality, maintainable, and secure code.

## Review Process

### Before Starting
1. **Understand the Context**
   - [ ] Read the PR/issue description carefully
   - [ ] Understand what problem this solves
   - [ ] Review related documentation/tickets
   - [ ] Check if this is the right approach

2. **Set Up Environment**
   - [ ] Check out the branch locally
   - [ ] Run the application
   - [ ] Verify the described behavior
   - [ ] Run tests successfully

### During Review

## 1. Functionality & Requirements

### Does It Work?
- [ ] Code implements the requirements correctly
- [ ] Edge cases are handled properly
- [ ] Error cases are handled appropriately
- [ ] Code works in all supported environments
- [ ] Manual testing confirms expected behavior

### Completeness
- [ ] All requirements are addressed
- [ ] No obvious features missing
- [ ] No TODO comments left for production code
- [ ] Configuration is documented

### User Experience
- [ ] User-facing strings are clear and typo-free
- [ ] Error messages are helpful
- [ ] Loading states are handled
- [ ] UX is consistent with the rest of the application

## 2. Code Quality

### Code Style & Formatting
- [ ] Follows project's style guide
- [ ] Consistent formatting throughout
- [ ] No commented-out code
- [ ] Meaningful variable/function names
- [ ] Proper indentation and spacing

**Examples:**

```javascript
// ❌ Bad - unclear naming
const d = getData().filter(x => x.v > 0);

// ✅ Good - descriptive naming
const activeProducts = fetchProducts().filter(p => p.value > 0);
```

### Code Organization
- [ ] Functions are small and focused (<50 lines)
- [ ] Files are organized logically
- [ ] Related code is grouped together
- [ ] Separation of concerns is maintained
- [ ] No duplicated code (DRY principle)

### Complexity
- [ ] Cyclomatic complexity is reasonable
- [ ] Deep nesting is avoided (max 3-4 levels)
- [ ] Long parameter lists are avoided
- [ ] Complex logic is commented or refactored

```go
// ❌ Too nested
func ProcessUser(id int) error {
    if id > 0 {
        user, err := db.GetUser(id)
        if err == nil {
            if user.Active {
                for _, item := range user.Items {
                    if item.Valid {
                        // ... more nesting
                    }
                }
            }
        }
    }
}

// ✅ Early returns reduce nesting
func ProcessUser(id int) error {
    if id <= 0 {
        return errors.New("invalid id")
    }

    user, err := db.GetUser(id)
    if err != nil {
        return fmt.Errorf("get user: %w", err)
    }
    if !user.Active {
        return nil
    }

    return processUserItems(user.Items)
}
```

## 3. Architecture & Design

### Design Patterns
- [ ] Appropriate design patterns used
- [ ] No anti-patterns present
- [ ] SOLID principles followed
- [ ] Code is extensible without modification

### Dependencies & Coupling
- [ ] Low coupling between components
- [ ] High cohesion within modules
- [ ] Dependencies are minimal
- [ ] No circular dependencies
- [ ] Appropriate use of interfaces

### Abstractions
- [ ] Right level of abstraction
- [ ] Not over-engineered
- [ ] Not under-engineered
- [ ] Leaks are properly encapsulated

```go
// ❌ Over-engineered for simple case
type UserManagerFactoryProvider interface {
    CreateFactory() UserManagerFactory
}

type UserManagerFactory interface {
    CreateUserManager() UserManager
}

// ✅ Simple and appropriate
func NewUserService(db *Database) *UserService {
    return &UserService{db: db}
}
```

## 4. Performance & Scalability

### Performance
- [ ] No obvious performance issues
- [ ] Database queries are optimized
- [ ] No N+1 query problems
- [ ] Caching used where appropriate
- [ ] Resource usage is reasonable

### Algorithmic Complexity
- [ ] Appropriate data structures used
- [ ] Efficient algorithms for the scale
- [ ] No unnecessary loops or iterations
- [ ] Pagination for large datasets

```javascript
// ❌ O(n²) - nested loop
const findDuplicates = (arr) => {
    const duplicates = [];
    for (let i = 0; i < arr.length; i++) {
        for (let j = i + 1; j < arr.length; j++) {
            if (arr[i] === arr[j]) duplicates.push(arr[i]);
        }
    }
    return duplicates;
};

// ✅ O(n) - using Set
const findDuplicates = (arr) => {
    const seen = new Set();
    const duplicates = [];
    for (const item of arr) {
        if (seen.has(item)) {
            duplicates.push(item);
        } else {
            seen.add(item);
        }
    }
    return duplicates;
};
```

### Scalability
- [ ] Handles increased load gracefully
- [ ] No single points of failure
- [ ] Asynchronous operations used appropriately
- [ ] Background jobs for long tasks

## 5. Security

### Input Validation
- [ ] All inputs are validated
- [ ] Sanitization of user input
- [ ] Protection against injection attacks
- [ ] SQL injection prevention
- [ ] XSS prevention

```go
// ✅ Parameterized query prevents SQL injection
func GetUser(db *sql.DB, username string) (*User, error) {
    var user User
    err := db.QueryRow(
        "SELECT id, name FROM users WHERE username = $1",
        username, // Safely parameterized
    ).Scan(&user.ID, &user.Name)
    return &user, err
}
```

### Authentication & Authorization
- [ ] Authentication is properly implemented
- [ ] Authorization checks on all endpoints
- [ ] No hardcoded credentials
- [ ] Sensitive data is encrypted
- [ ] Secrets are properly managed

### Data Protection
- [ ] Sensitive data not logged
- [ ] PII is protected
- [ ] HTTPS enforced
- [ ] Secure headers configured
- [ ] No information leakage in errors

### Dependencies
- [ ] No known vulnerabilities in dependencies
- [ ] Dependencies are up to date
- [ ] License compatibility checked

```bash
# Check for vulnerabilities
npm audit         # JavaScript
go mod audit      # Go
pip check         # Python
safety check      # Python
```

## 6. Testing

### Test Coverage
- [ ] Tests cover new functionality
- [ ] Tests cover edge cases
- [ ] Tests cover error paths
- [ ] Coverage is not decreased
- [ ] Critical paths have 100% coverage

### Test Quality
- [ ] Tests are readable and maintainable
- [ ] Tests are independent
- [ ] Tests are deterministic
- [ ] No hardcoded test values
- [ ] Proper test setup and teardown

```go
// ✅ Good test - descriptive and isolated
func TestUser_Create_WithValidData_SavesToDatabase(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    service := NewUserService(db)

    // Act
    user, err := service.Create("test@example.com", "password")

    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if user.Email != "test@example.com" {
        t.Errorf("expected email 'test@example.com', got %v", user.Email)
    }
}
```

### Test Types
- [ ] Unit tests for business logic
- [ ] Integration tests where appropriate
- [ ] E2E tests for critical flows
- [ ] Performance tests for critical paths

## 7. Error Handling

### Error Cases
- [ ] Errors are handled, not ignored
- [ ] Error messages are helpful
- [ ] Errors include context
- [ ] Proper error types used
- [ ] Errors are logged appropriately

```go
// ❌ Bad - error ignored
user, _ := db.GetUser(id)

// ✅ Good - error handled
user, err := db.GetUser(id)
if err != nil {
    return fmt.Errorf("failed to get user %d: %w", id, err)
}
```

### Recovery
- [ ] Graceful degradation where possible
- [ ] No panics in production code
- [ ] Database transaction rollback on error
- [ ] Resource cleanup on error

## 8. Documentation

### Code Comments
- [ ] Complex logic is explained
- [ ] Non-obvious decisions are documented
- [ ] No obvious/redundant comments
- [ ] Public APIs are documented
- [ ] TODO/FIXME comments have issues

```go
// ❌ Bad comment - redundant
// Increment count by 1
count++

// ✅ Good comment - explains why
// Using exponential backoff to avoid thundering herd
time.Sleep(backoff.Duration())
```

### API Documentation
- [ ] API changes are documented
- [ ] Parameters are documented
- [ ] Return values are documented
- [ ] Examples are provided
- [ ] Breaking changes are highlighted

### README/Docs
- [ ] README is updated if needed
- [ ] Configuration changes documented
- [ ] New features documented
- [ ] Migration guide provided if needed
- [ ] Diagrams updated for architecture changes

## 9. Compatibility & Standards

### Backwards Compatibility
- [ ] No breaking changes for existing users
- [ ] Breaking changes are documented in CHANGELOG
- [ ] Deprecation warnings added for old APIs
- [ ] Migration path provided

### Standards Compliance
- [ ] Follows language idioms
- [ ] Follows project conventions
- [ ] Platform/standards compliance
- [ ] Accessibility standards (WCAG)

### Version Control
- [ ] Commit messages are clear
- [ ] Commits are atomic
- [ ] No merge commits in feature branch
- [ ] Branch follows naming convention

## 10. Deployment & Operations

### Deployment
- [ ] Database migrations included
- [ ] Configuration changes documented
- [ ] No manual steps required for deploy
- [ ] Rollback plan exists

### Monitoring & Logging
- [ ] Appropriate logging added
- [ ] Metrics added where needed
- [ ] Health checks updated
- [ ] Debug logs removed or marked

### Operations
- [ ] Runbooks updated if needed
- [ ] Alerts configured for new issues
- [ ] Documentation for ops team

## Review Feedback Guidelines

### How to Give Feedback

**Be Constructive:**
```
❌ "This code is terrible."

✅ "I think we could simplify this by extracting the validation
   logic into a separate function. That would make it easier to
   test and reuse."
```

**Provide Context:**
```
✅ "I noticed we're using the old API here. According to the
   migration guide, we should use the new API. Here's the link
   to the docs: [link]"
```

**Suggest, Don't Dictate:**
```
❌ "Change this to use async/await."

✅ "Have you considered using async/await here? It might make
   this code more readable."
```

### What to Approve vs. Comment On

**🟢 Approve:**
- Minor style issues that don't affect functionality
- Personal preferences (debatable)
- Things that can be improved in a follow-up PR
- Documentation typos (unless critical)

**🟡 Request Changes:**
- Security issues
- Performance problems
- Broken functionality
- Missing tests
- Poor error handling

**🔴 Block:**
- Security vulnerabilities
- Data loss risk
- Breaking changes without proper process
- Complete misunderstanding of requirements

## Post-Review Actions

### For Author
- [ ] Address all review comments
- [ ] Push changes to same branch (not new commits)
- [ ] Mark comments as resolved
- [ ] Add comments explaining why you didn't make a change
- [ ] Thank reviewers for their time

### For Reviewer
- [ ] Re-review after changes
- [ ] Approve once satisfied
- [ ] Delete branch after merge (if appropriate)
- [ ] Celebrate the successful merge! 🎉

## Quick Reference

**Critical Must-Haves:**
- ✅ Tests pass
- ✅ No security vulnerabilities
- ✅ No obvious performance issues
- ✅ Error handling in place
- ✅ Code is readable and maintainable

**Nice-to-Haves:**
- 📝 Comprehensive documentation
- 🧪 High test coverage
- 🎨 Consistent style
- 🔧 Easy to maintain
- 📚 Well-structured

**Deal Breakers:**
- ❌ Breaking changes without notice
- ❌ Security vulnerabilities
- ❌ Data loss potential
- ❌ No tests for critical code
- ❌ Hardcoded credentials/secrets
