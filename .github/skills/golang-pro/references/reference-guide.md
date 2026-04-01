# Go Pro Reference: Idioms, Anti-Patterns, Project Structure, Performance, Tooling

## 1. Idiomatic Go Examples

### Simplicity and Clarity
```go
// Good: Clear and direct
func Add(a, b int) int {
    return a + b
}

// Bad: Overly clever
func Add(a, b int) (result int) {
    defer func() { result = a + b }()
    return
}
```

### Make Zero Value Useful
```go
type Counter struct {
    mu sync.Mutex
    count int // zero value is 0
}
```

### Accept Interfaces, Return Structs
```go
func ReadAll(r io.Reader) ([]byte, error) {
    return io.ReadAll(r)
}
```

## 2. Common Go Idioms

| Idiom | Description |
|-------|-------------|
| Accept interfaces, return structs | Functions accept interface params, return concrete types |
| Errors are values | Treat errors as first-class values |
| Don't communicate by sharing memory | Use channels for coordination |
| Make the zero value useful | Types should work without explicit initialization |
| Return early | Handle errors first, keep happy path unindented |

## 3. Anti-Patterns to Avoid

```go
// Bad: Naked returns in long functions
func process() (result int, err error) {
    // ... 50 lines ...
    return // What is being returned?
}

// Bad: Using panic for control flow
func GetUser(id string) *User {
    user, err := db.Find(id)
    if err != nil {
        panic(err)
    }
    return user
}

// Bad: Passing context in struct
// Good: Context as first parameter
func ProcessRequest(ctx context.Context, id string) error {
    // ...
}
```



## 5. Tooling & Linting

- **Build & Run:**
    - `go build ./...`
    - `go run ./cmd/myapp`
- **Testing:**
    - `go test ./...`
    - `go test -race ./...`
    - `go test -cover ./...`
- **Static Analysis:**
    - `go vet ./...`
    - `golangci-lint run`
- **Formatting:**
    - `gofmt -w .`
    - `goimports -w .`
- **Recommended .golangci.yml:**
```yaml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - unconvert
    - unparam
```

---

**Note:** Tham khảo thêm các file reference khác trong thư mục này để có hướng dẫn chuyên sâu về concurrency, interfaces, generics, testing, project structure.