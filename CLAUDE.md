# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests
go test ./... -race

# Run a single test
go test ./internal/services -run TestFunctionName -v

# Lint
go tool golangci-lint run ./...

# Vulnerability scan
go tool govulncheck ./...

# Run server
go run main.go server
```

## Architecture

The backend is a layered Go REST API with manual dependency injection:

```
Handler → Service → Repository / Storage
```

- **Handlers** (`internal/handlers/`) use [Fuego](https://github.com/go-fuego/fuego) and depend on service interfaces
- **Services** (`internal/services/`) hold business logic and depend on repository/storage interfaces
- **Repositories** (`internal/repositories/`) implement DB access via GORM
- **Storage** (`internal/storage/`) implements file I/O via `FileStorage` interface (DiskStorage uses UUID as directory name)
- **Middleware** (`internal/middleware/`) validates JWT and injects `user_id` into context; handlers retrieve it via `middleware.GetUserID(c.Context())`

Dependencies are wired in `cmd/server.go`: repos → services → handlers, all registered on a Fuego server with CORS.

All DB and storage methods take `context.Context` as first arg. Repository queries always filter by `proprietary_id` (owner's user ID) for row-level security.

## Mocking Pattern

All mocks live in `internal/testutil/mocks.go` and use function fields — no code generation:

```go
type MockUserService struct {
    GetUserFunc func(ctx context.Context, id uuid.UUID) (*models.User, error)
}
func (m *MockUserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
    return m.GetUserFunc(ctx, id)
}
```

Tests set behavior and assertions inline on the func field. Use `fuego.NewMockContextNoBody()` / `fuego.NewMockContext[B, R]()` for handler tests.

## Key Conventions

- All HTTP requests in tests must use `httptest.NewRequestWithContext(t.Context(), ...)` (enforced by `noctx` linter)
- Successful responses are wrapped in `ApiResponse[T]{Data, Message}`
- Test function names follow `TestWhat_ExpectedOutcome_Condition`
- `internal/testutil/` is excluded from Codecov coverage
