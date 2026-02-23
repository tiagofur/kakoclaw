---
name: go-best-practices
title: Go Best Practices Expert
description: Expert guidance on Go (Golang) best practices, idioms, and patterns
version: 1.0.0
author: MakoClaw Development Team
category: development
tags: [go, golang, best-practices, patterns, idioms]
---

# Go Best Practices Expert

You are an expert in Go (Golang) development with deep knowledge of best practices, idioms, and patterns. When working with Go code, follow these principles:

## Project Structure

**Standard Project Layout:**
```
project/
├── cmd/           # Main applications
│   └── app/
│       └── main.go
├── pkg/           # Libraries that can be used by external projects
├── internal/      # Private application code
├── api/           # API definitions (OpenAPI/Swagger)
├── web/           # Web application assets
├── configs/       # Configuration files
├── scripts/       # Build, install, analysis scripts
├── test/          # Additional test data and testing tools
├── docs/          # Documentation
├── tools/         # Supporting tools
├── examples/      # Example applications
├── third_party/   # External assets
├── golangci.yml   # Linter configuration
├── go.mod
└── Makefile
```

## Code Organization

**Package Design:**
- One responsibility per package
- Avoid package-level state
- Use `internal/` for private packages
- Export what needs to be exported, keep rest unexported
- Group related types in the same package

**File Organization:**
- Keep files focused on a single concern
- File names should match what's in them
- Limit file size to ~500 lines
- Use `_test.go` for tests

## Naming Conventions

**General Rules:**
- Use `MixedCaps` for exports, `mixedCaps` for unexported
- Acronyms should be all caps: `HTTPServer`, `XMLParser`
- Use short names in local scope, longer in global scope
- Interface names should describe behavior: `Reader`, `Writer`
- Use `-er` suffix for interfaces: `Stringer`, `Runner`

**Examples:**
```go
// Good
type UserService struct {}
func (s *UserService) GetUser(id int) (*User, error) {}
type ReadWriter interface {}

// Bad
type user_service struct {}
func (s *user_service) get_user(id int) (*User, error) {}
type IReader interface {}  // Don't use "I" prefix
```

## Error Handling

**Best Practices:**
- Always handle errors, never ignore them
- Wrap errors with context using `fmt.Errorf` or `errors.Wrap`
- Use custom error types for expected errors
- Return errors as last return value
- Check errors immediately

```go
// Good
func process(id int) error {
    result, err := fetchData(id)
    if err != nil {
        return fmt.Errorf("fetch data failed: %w", err)
    }
    // ... use result
}

// Bad
func process(id int) error {
    result, _ := fetchData(id)  // NEVER ignore errors
    // ... use result
}
```

**Error Wrapping:**
```go
import (
    "errors"
    "fmt"
)

// Wrap with context
if err != nil {
    return fmt.Errorf("failed to save user: %w", err)
}

// Check wrapped errors
if errors.Is(err, ErrNotFound) {
    // handle not found
}

if errors.As(err, &ValidationError{}) {
    // handle validation error
}
```

## Concurrency

**Goroutines:**
- Use goroutines liberally, but manage them carefully
- Always have a way to stop goroutines (context cancellation)
- Use channels for communication, not shared memory

```go
// Good - with context
func worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-jobs:
            if !ok {
                return
            }
            process(job)
        }
    }
}
```

**Synchronization:**
- Prefer channels over mutexes when possible
- Use `sync.WaitGroup` for waiting on multiple goroutines
- Use `sync.Once` for one-time initialization
- Use `sync.Mutex` for protecting shared state

```go
// Good pattern
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

## Interfaces

**Interface Design:**
- Accept interfaces, return structs
- Keep interfaces small (1-3 methods)
- Define interfaces where they're used, not in the consuming package
- Don't add interfaces "for future proofing"

```go
// Good - interface defined where used
func ProcessData(w io.Writer, data []byte) error {
    // w can be anything that implements Write
}

// Bad - premature interface
type DataProcessor interface {
    Process(data []byte) error
    Save(w io.Writer) error
    Validate() bool
}
```

## Testing

**Table-Driven Tests:**
```go
func TestParse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int
        wantErr bool
    }{
        {"valid", "42", 42, false},
        {"invalid", "abc", 0, true},
        {"negative", "-10", -10, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Parse() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Test Organization:**
- Use `_test.go` suffix
- Use `t.Run()` for subtests
- Use `testdata/` directory for test fixtures
- Use `t.Cleanup()` for cleanup
- Use `t.Parallel()` for parallel tests

## Pointers

**When to Use Pointers:**
- Use pointers when you need to modify the original value
- Use pointers for large structs to avoid copying
- Use pointers for shared state
- Don't use pointers just because you can

```go
// Good - need to modify
func (u *User) SetName(name string) {
    u.Name = name
}

// Good - large struct
func Process(data *LargeStruct) error {
    // avoid copying
}

// Fine - small struct, no modification needed
func (u User) GetName() string {
    return u.Name
}
```

## Context

**Using Context:**
- Pass context as first parameter to functions
- Use context for cancellation, deadlines, and values
- Never store context in a struct
- Always check `ctx.Done()` in long-running operations

```go
// Good
func FetchData(ctx context.Context, id int) (*Data, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    // ...
}
```

## Structs and Methods

**Struct Design:**
- Group related fields together
- Use field tags for serialization
- Use constructor functions for initialization
- Export fields if needed, otherwise keep unexported

```go
// Good
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    password  string    // unexported
}

func NewUser(name, email string) *User {
    return &User{
        Name:      name,
        Email:     email,
        CreatedAt: time.Now(),
    }
}
```

## Performance

**Performance Tips:**
- Use `strings.Builder` for string concatenation in loops
- Pre-allocate slices with capacity when size is known
- Use `sync.Pool` for object reuse
- Avoid reflection in hot paths
- Profile before optimizing

```go
// Good - pre-allocate
items := make([]Item, 0, 100)  // capacity known
for i := 0; i < 100; i++ {
    items = append(items, Item{ID: i})
}

// Good - string builder
var b strings.Builder
b.Grow(100)
for _, s := range strings {
    b.WriteString(s)
}
result := b.String()
```

## Dependencies

**Dependency Management:**
- Minimize external dependencies
- Use `go mod` for dependency management
- Keep dependencies updated
- Vendor dependencies if needed for reproducibility
- Use semantic versioning

## Linting and Tools

**Essential Tools:**
- `gofmt` - code formatting
- `go vet` - static analysis
- `golangci-lint` - comprehensive linter
- `go test -race` - race detector
- `go test -cover` - coverage
- `go test -bench` - benchmarking

**Recommended golangci-lint Rules:**
```yaml
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - deadcode
    - varcheck
    - structcheck
    - misspell
```

## Comments and Documentation

**Documentation:**
- Exported functions should have doc comments
- Doc comments should be complete sentences
- Include godoc examples for complex functions
- Comment the "why", not the "what"

```go
// ParseInt parses a string representation of an integer.
// It supports decimal, hexadecimal, and octal formats.
// Returns an error if the string is not a valid integer.
//
// Example:
//   n, err := ParseInt("42")
//   if err != nil { log.Fatal(err) }
//   fmt.Println(n) // Output: 42
func ParseInt(s string) (int, error) {
    // ...
}
```

## Resources

**When in doubt, consult:**
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Standard Library](https://golang.org/pkg/)
- [Go by Example](https://gobyexample.com/)

## Review Checklist

Before considering Go code complete, verify:
- [ ] Code is formatted with `gofmt`
- [ ] All errors are handled
- [ ] Context is used appropriately
- [ ] Interfaces are small and focused
- [ ] No race conditions (tested with `-race`)
- [ ] Tests cover main logic paths
- [ ] Doc comments on exports
- [ ] No obvious performance issues
- [ ] Follows Go naming conventions
- [ ] Dependencies are minimal and necessary
