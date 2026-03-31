---
name: go-unit-test
description: "Write Go unit tests for BankEase microservices (identity-service, user-profile-service, saving-service, bff-service). Use when: creating unit tests, adding test coverage, writing tests for gRPC handlers, HTTP handlers, DB providers, JWT, utils, or interceptors. Produces test files following the project standard: testify/assert, go-sqlmock, newTestServer(t) constructor, table-driven tests, minimum 90% coverage. Also handles running tests with coverage profile, SonarQube integration."
argument-hint: "Package or file to test (e.g. 'server/api/bff_auth_api.go')"
---

# Go Unit Test — BankEase Services

## When to Use

- Writing new unit tests for any service (`identity-service`, `user-profile-service`, `saving-service`, `bff-service`)
- Adding coverage for an untested package
- Verifying test output and coverage ≥ 90%
- Setting up SonarQube coverage reporting

---

## Procedure

### Step 1 — Understand the Target File

Read the file to test. Identify:
- Package name (same package test = white-box; `_test` suffix = black-box)
- Structs, exported functions, and methods to cover
- External dependencies: `*sql.DB`, gRPC clients, JWT manager, logger
- Which layer it belongs to (see [Layer Map](#layer-map))

### Step 2 — Pick the Right Test Harness

| Layer | Harness |
|---|---|
| `server/api/` (HTTP + gRPC handlers) | `newTestServer(t)` constructor — see [Test Server Pattern](#test-server-pattern) |
| `server/db/` (DB providers) | `sqlmock.New()` directly — see [DB Provider Pattern](#db-provider-pattern) |
| `server/jwt/` | Real struct with test secrets |
| `server/lib/database/` | Combination of `sql.Open` mocks + interface assertions |
| `server/utils/`, `server/constant/` | Plain unit tests, no mocks needed |
| `server/api/` interceptors | Bare handler funcs + `mockServerStream` — see [Interceptor Pattern](#interceptor-pattern) |

### Step 3 — Write the Test File

**Naming**: `<original_file_name>_test.go` in the **same directory**  
**Package**: same package as source file (white-box), e.g. `package api`

**Required imports:**
```go
import (
    "testing"

    "github.com/stretchr/testify/assert"
    // add as needed:
    "github.com/DATA-DOG/go-sqlmock"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)
```

### Step 4 — Cover All Cases per Function

For every exported function/method, write at minimum:
1. **Happy path** — valid input, expected success result
2. **Error path(s)** — DB errors, invalid tokens, missing fields, wrong types
3. **Edge/security cases** — empty strings, nil pointers, JWT "none" algorithm (for JWT tests)

Use **table-driven tests** when a function has 3+ input variants:
```go
func TestFoo(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"success", "valid", false},
        {"empty input", "", true},
        {"db error", "causes_error", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Step 5 — Run with Coverage

```powershell
# From the service root (e.g. identity-service/)
go test -v -count=1 -parallel=4 -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Check that **total coverage ≥ 90%**.  
To view HTML report:
```powershell
go tool cover -html=coverage.out -o coverage.html
```

### Step 6 — Fix SonarQube Coverage (if needed)

Add to `sonar-project.properties` if not present:
```properties
sonar.go.coverage.reportPaths=coverage.out
```

Generate coverage before running SonarQube scan:
```powershell
go test -coverprofile=coverage.out ./...
```

---

## Layer Map

```
server/
├── api/              → HTTP handlers + gRPC handlers + interceptors
│   ├── *_api.go      → test: happy path + error path for each handler
│   ├── *_grpc_api.go → test: gRPC request/response + error codes
│   └── *_interceptor.go → test: each interceptor independently + chained
├── db/               → DB providers (sqlmock)
│   └── *_provider.go → test: each query: success, not found, DB error
├── jwt/              → JWT manager
│   └── manager.go    → test: generate, verify, expired, wrong secret, none-alg
├── lib/database/     → DB wrapper
│   └── *.go          → test: connect, retry, ping, transaction
├── utils/            → Pure helpers
│   └── utils.go      → test: each helper, including ctx injection
└── constant/         → Constants
    └── constant.go   → test: value assertions
```

---

## Test Server Pattern

For `server/api/` packages, create a `newTestServer(t)` constructor at the **top of the test file**:

```go
func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock) {
    t.Helper()
    db, mock, err := sqlmock.New()
    require.NoError(t, err)

    // Use the project's DatabaseMock or wrap *sql.DB directly
    dbMock := &databasemock.DatabaseMock{DB: db}

    logger, _ := zap.NewDevelopment()

    jwtMgr, err := jwt.NewJWTManager("test-secret-key-32chars!!", 24*time.Hour)
    require.NoError(t, err)

    provider := db.NewProvider(dbMock)
    srv := NewServer(provider, jwtMgr, logger)

    t.Cleanup(func() { db.Close() })
    return srv, mock
}
```

> Adapt field names to the actual `Server` struct and constructor signature of each service.

---

## DB Provider Pattern

```go
func TestCreateUser_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    mock.ExpectQuery(`INSERT INTO users`).
        WithArgs(sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), "081234").
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("uuid-123"))

    provider := NewProvider(&databasemock.DatabaseMock{DB: db})
    id, err := provider.CreateUser(context.Background(), "alice", "hashed", "081234")

    assert.NoError(t, err)
    assert.NotEmpty(t, id)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_DBError(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    mock.ExpectQuery(`INSERT INTO users`).
        WillReturnError(errors.New("connection refused"))

    provider := NewProvider(&databasemock.DatabaseMock{DB: db})
    _, err := provider.CreateUser(context.Background(), "alice", "hashed", "")

    assert.Error(t, err)
}
```

---

## Interceptor Pattern

For gRPC stream interceptors, define a minimal mock at the test file level:

```go
type mockServerStream struct {
    grpc.ServerStream
    ctx context.Context
}

func (m *mockServerStream) Context() context.Context { return m.ctx }
```

Test each interceptor as a standalone function:

```go
func TestAuthInterceptor_ValidToken(t *testing.T) {
    srv, _ := newTestServer(t)
    token, _ := srv.jwtMgr.Generate("user-id", "alice")

    md := metadata.Pairs("authorization", "Bearer "+token)
    ctx := metadata.NewIncomingContext(context.Background(), md)
    info := &grpc.UnaryServerInfo{FullMethod: "/IdentityService/GetMe"}

    called := false
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        called = true
        return nil, nil
    }

    _, err := srv.AuthInterceptor(ctx, nil, info, handler)
    assert.NoError(t, err)
    assert.True(t, called)
}
```

---

## JWT Security Tests (Required)

Always include the "none" algorithm rejection test:

```go
func TestVerify_TokenWithNoneAlgorithm(t *testing.T) {
    mgr, _ := jwt.NewJWTManager("secret", time.Hour)

    // Craft token with "none" alg manually
    header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
    payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"uid","exp":9999999999}`))
    noneToken := header + "." + payload + "."

    _, err := mgr.Verify(noneToken)
    assert.Error(t, err, "should reject 'none' algorithm")
}
```

---

## gRPC Error Code Assertions

```go
// Assert specific gRPC status code
st, ok := status.FromError(err)
assert.True(t, ok)
assert.Equal(t, codes.NotFound, st.Code())
assert.Contains(t, st.Message(), "not found")
```

---

## Coverage Checklist

Before marking done, verify:

- [ ] Every **exported function** in the target package has at least one test
- [ ] Every **error path** (DB error, validation error, auth error) is covered
- [ ] **Table-driven tests** used for functions with 3+ variants
- [ ] `mock.ExpectationsWereMet()` called after all sqlmock tests
- [ ] `assert.NoError(t, err)` used for setup steps; `require.NoError` for fatal setup
- [ ] **JWT none-algorithm** test present if testing JWT package
- [ ] Run `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out` — total ≥ 90%
- [ ] No `t.Skip()` or `// TODO: test` left in final output
- [ ] `sonar-project.properties` has `sonar.go.coverage.reportPaths=coverage.out`

---

## Quick Reference — Test Commands

```powershell
# Run all tests (verbose, no cache)
go test -v -count=1 -parallel=4 ./...

# Run with coverage
go test -v -count=1 -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Run a single package
go test -v -count=1 ./server/api/...

# Run a specific test by name
go test -v -run TestSignUp_Success ./server/api/...
```
