# Implementation Plan: Search Feature API — user-profile-service

## TL;DR

Add 3 new endpoints to **user-profile-service** based on the Search Feature API spec (`api.txt`):

| Endpoint | Description |
|---|---|
| `GET /api/exchange-rates` | Returns list of currency exchange rates |
| `GET /api/interest-rates` | Returns list of deposit interest rates |
| `GET /api/branches?q={query}` | Returns BRI branches, optionally filtered by name |

Each domain follows the existing layered architecture:
**migration SQL → model → repository → HTTP handler → gRPC handler → proto + protogen → wiring**

Responses are **raw JSON arrays** (matching api.txt spec — no `{code, description}` wrapper).
All 3 endpoints exposed via **REST + gRPC**.

---

## Architecture Flow

```
api.txt spec
     │
     ▼
Migration SQL (DDL + seed data)
     │
     ▼
Models (internal/models/search.go)
     │
     ▼
Repositories (3 new files in internal/repository/)
     │
     ├──▶ HTTP Handlers (internal/handlers/search.go)
     │         │
     │         ▼
     │    Router (internal/server/router.go)
     │
     └──▶ gRPC Handlers (internal/grpchandler/search.go)
               │
               ▼
          Proto + Protogen updates
               │
               ▼
          Server wiring (internal/server/server.go)
```

---

## Phase 1 — Database: Migrations + Seed

> **Pattern:** Follow existing naming convention (`NNN_description.sql`). Auto-run by `internal/db/migrate.go` on startup.

### Files to Create

**`internal/db/migrations/004_add_exchange_rates.sql`**
```sql
CREATE TABLE IF NOT EXISTS exchange_rate (
    id           VARCHAR(50)    PRIMARY KEY,
    country      VARCHAR(100)   NOT NULL,
    currency     VARCHAR(10)    NOT NULL,
    country_code VARCHAR(10)    NOT NULL,
    buy          DECIMAL(15,4)  NOT NULL,
    sell         DECIMAL(15,4)  NOT NULL
);
```

**`internal/db/migrations/005_add_interest_rates.sql`**
```sql
CREATE TABLE IF NOT EXISTS interest_rate (
    id      VARCHAR(50)   PRIMARY KEY,
    kind    VARCHAR(20)   NOT NULL,
    deposit VARCHAR(10)   NOT NULL,
    rate    DECIMAL(5,2)  NOT NULL
);
```

**`internal/db/migrations/006_add_branches.sql`**
```sql
CREATE TABLE IF NOT EXISTS branch (
    id        VARCHAR(50)    PRIMARY KEY,
    name      VARCHAR(200)   NOT NULL,
    distance  VARCHAR(20)    NOT NULL,
    latitude  DECIMAL(10,6)  NOT NULL,
    longitude DECIMAL(10,6)  NOT NULL
);
```

### File to Update

**`seed.sql`** — append INSERT blocks with sample data from api.txt:

```sql
-- exchange_rate seed
INSERT INTO exchange_rate (id, country, currency, country_code, buy, sell) VALUES
  ('a1b2c3d4-0001-4000-8000-000000000001', 'Vietnam',   'VND', 'VN', 1.403,  1.746),
  ('a1b2c3d4-0001-4000-8000-000000000002', 'Nicaragua', 'NIO', 'NI', 9.123,  12.09),
  ('a1b2c3d4-0001-4000-8000-000000000003', 'Korea',     'KRW', 'KR', 3.704,  5.151),
  ('a1b2c3d4-0001-4000-8000-000000000004', 'China',     'CNY', 'CN', 1.725,  2.234)
ON CONFLICT (id) DO NOTHING;

-- interest_rate seed
INSERT INTO interest_rate (id, kind, deposit, rate) VALUES
  ('b2c3d4e5-0002-4000-8000-000000000001', 'individual', '1m',  4.5),
  ('b2c3d4e5-0002-4000-8000-000000000002', 'corporate',  '2m',  5.5),
  ('b2c3d4e5-0002-4000-8000-000000000004', 'corporate',  '6m',  2.5),
  ('b2c3d4e5-0002-4000-8000-000000000011', 'individual', '12m', 5.9)
ON CONFLICT (id) DO NOTHING;

-- branch seed
INSERT INTO branch (id, name, distance, latitude, longitude) VALUES
  ('c3d4e5f6-0003-4000-8000-000000000001', 'Bank 1656 Union Street',    '50m',    -6.2,   106.816),
  ('c3d4e5f6-0003-4000-8000-000000000002', 'Bank Secaucus',              '1,2 km', -6.205, 106.82),
  ('c3d4e5f6-0003-4000-8000-000000000003', 'Bank 1657 Riverside Drive',  '5,3 km', -6.195, 106.825),
  ('c3d4e5f6-0003-4000-8000-000000000004', 'Bank Rutherford',            '70m',    -6.21,  106.812),
  ('c3d4e5f6-0003-4000-8000-000000000005', 'Bank 1656 Union Street',    '30m',    -6.208, 106.814)
ON CONFLICT (id) DO NOTHING;
```

---

## Phase 2 — Models

> **Pattern:** Follow `internal/models/menu.go` — plain Go structs with `json` tags matching the API spec field names exactly.

### File to Create

**`internal/models/search.go`**

```go
package models

type ExchangeRate struct {
    ID          string  `json:"id"`
    Country     string  `json:"country"`
    Currency    string  `json:"currency"`
    CountryCode string  `json:"countryCode"`
    Buy         float64 `json:"buy"`
    Sell        float64 `json:"sell"`
}

type InterestRate struct {
    ID      string  `json:"id"`
    Kind    string  `json:"kind"`
    Deposit string  `json:"deposit"`
    Rate    float64 `json:"rate"`
}

type Branch struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Distance  string  `json:"distance"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
}
```

> **Note:** `CountryCode` uses camelCase JSON tag to match the api.txt spec (`countryCode`), while the DB column uses `country_code`.

---

## Phase 3 — Repositories

> **Pattern:** Follow `internal/repository/menu.go` — struct with `*sql.DB`, `context.WithTimeout(ctx, 5*time.Second)`, `rows.Scan`.

### Files to Create

**`internal/repository/exchange_rate.go`**

```go
type ExchangeRateRepository struct { DB *sql.DB }

func (r *ExchangeRateRepository) GetAll(ctx context.Context) ([]models.ExchangeRate, error)
    // SELECT id, country, currency, country_code, buy, sell FROM exchange_rate ORDER BY country
```

**`internal/repository/interest_rate.go`**

```go
type InterestRateRepository struct { DB *sql.DB }

func (r *InterestRateRepository) GetAll(ctx context.Context) ([]models.InterestRate, error)
    // SELECT id, kind, deposit, rate FROM interest_rate ORDER BY kind, deposit
```

**`internal/repository/branch.go`**

```go
type BranchRepository struct { DB *sql.DB }

func (r *BranchRepository) GetAll(ctx context.Context) ([]models.Branch, error)
    // SELECT id, name, distance, latitude, longitude FROM branch

func (r *BranchRepository) SearchByName(ctx context.Context, q string) ([]models.Branch, error)
    // SELECT ... FROM branch WHERE name ILIKE '%' || $1 || '%'
```

---

## Phase 4 — HTTP Handlers

> **Pattern:** Follow `internal/handlers/menu.go` — handler struct, `writeJSON` / `writeError` helpers, return raw array.

### File to Create

**`internal/handlers/search.go`**

```go
type SearchHandler struct {
    ExchangeRateRepo *repository.ExchangeRateRepository
    InterestRateRepo *repository.InterestRateRepository
    BranchRepo       *repository.BranchRepository
}

func (h *SearchHandler) GetExchangeRates(w http.ResponseWriter, r *http.Request)
    // → writeJSON(w, 200, []ExchangeRate)

func (h *SearchHandler) GetInterestRates(w http.ResponseWriter, r *http.Request)
    // → writeJSON(w, 200, []InterestRate)

func (h *SearchHandler) GetBranches(w http.ResponseWriter, r *http.Request)
    // q := r.URL.Query().Get("q")
    // if q == "" → repo.GetAll
    // else       → repo.SearchByName(ctx, q)
    // → writeJSON(w, 200, []Branch)
```

---

## Phase 5 — Proto + Protogen

> **Pattern:** Hand-written protogen (no protoc). Follow the existing structs in `protogen/user-profile-service/`.

### 5a. Update `proto/user_profile_api.proto`

Add 3 RPCs to the `UserProfileService`:

```proto
rpc GetExchangeRates(GetExchangeRatesRequest) returns (ExchangeRateListResponse);
rpc GetInterestRates(GetInterestRatesRequest) returns (InterestRateListResponse);
rpc GetBranches(GetBranchesRequest)           returns (BranchListResponse);
```

### 5b. Update `proto/user_profile_payload.proto`

Add new messages:

```proto
// Exchange Rate
message GetExchangeRatesRequest {}
message ExchangeRateItem {
    string id           = 1;
    string country      = 2;
    string currency     = 3;
    string country_code = 4;
    double buy          = 5;
    double sell         = 6;
}
message ExchangeRateListResponse {
    repeated ExchangeRateItem exchange_rates = 1;
}

// Interest Rate
message GetInterestRatesRequest {}
message InterestRateItem {
    string id      = 1;
    string kind    = 2;
    string deposit = 3;
    double rate    = 4;
}
message InterestRateListResponse {
    repeated InterestRateItem interest_rates = 1;
}

// Branch
message GetBranchesRequest {
    string query = 1; // optional search query
}
message BranchItem {
    string id        = 1;
    string name      = 2;
    string distance  = 3;
    double latitude  = 4;
    double longitude = 5;
}
message BranchListResponse {
    repeated BranchItem branches = 1;
}
```

### 5c. Update `protogen/user-profile-service/user_profile_payload.pb.go`

Add hand-written Go structs + getter methods for all 6 new messages above, following the existing struct pattern in the file.

### 5d. Update `protogen/user-profile-service/user_profile_api_grpc.pb.go`

1. Add 3 new method signatures to `UserProfileServiceServer` interface
2. Add 3 unimplemented stubs to `UnimplementedUserProfileServiceServer`
3. Add 3 handler functions (`_UserProfileService_GetExchangeRates_Handler`, etc.)
4. Register them in `UserProfileService_ServiceDesc.Methods`

---

## Phase 6 — gRPC Handlers

> **Pattern:** Follow `internal/grpchandler/menu.go` — receiver on `*GrpcServer`, call repo, convert model → proto.

### File to Create

**`internal/grpchandler/search.go`**

```go
func (s *GrpcServer) GetExchangeRates(ctx context.Context, _ *pb.GetExchangeRatesRequest) (*pb.ExchangeRateListResponse, error)

func (s *GrpcServer) GetInterestRates(ctx context.Context, _ *pb.GetInterestRatesRequest) (*pb.InterestRateListResponse, error)

func (s *GrpcServer) GetBranches(ctx context.Context, req *pb.GetBranchesRequest) (*pb.BranchListResponse, error)
    // if req.GetQuery() == "" → repo.GetAll else → repo.SearchByName
```

### File to Update

**`internal/grpchandler/profile.go`** — add 3 new repository fields to `GrpcServer` struct:

```go
type GrpcServer struct {
    pb.UnimplementedUserProfileServiceServer
    ProfileRepo      *repository.ProfileRepository
    MenuRepo         *repository.MenuRepository
    ExchangeRateRepo *repository.ExchangeRateRepository  // new
    InterestRateRepo *repository.InterestRateRepository  // new
    BranchRepo       *repository.BranchRepository        // new
}
```

**`internal/grpchandler/converter.go`** — add 3 conversion helpers:

```go
func exchangeRatesToProto(items []models.ExchangeRate) []*pb.ExchangeRateItem
func interestRatesToProto(items []models.InterestRate) []*pb.InterestRateItem
func branchesToProto(items []models.Branch) []*pb.BranchItem
```

---

## Phase 7 — Wiring

### Update `internal/server/server.go`

1. Add repo fields to `Server` struct:
   ```go
   exchangeRateRepo *repository.ExchangeRateRepository
   interestRateRepo *repository.InterestRateRepository
   branchRepo       *repository.BranchRepository
   ```

2. Instantiate in `NewServer`:
   ```go
   exchangeRateRepo := &repository.ExchangeRateRepository{DB: db}
   interestRateRepo := &repository.InterestRateRepository{DB: db}
   branchRepo       := &repository.BranchRepository{DB: db}

   searchHandler := &handlers.SearchHandler{
       ExchangeRateRepo: exchangeRateRepo,
       InterestRateRepo: interestRateRepo,
       BranchRepo:       branchRepo,
   }
   ```

3. Pass to `setupRoutes`:
   ```go
   s.Router = setupRoutes(profileHandler, menuHandler, uploadHandler, searchHandler)
   ```

4. Update `GrpcServer` initialization in `StartGRPC`:
   ```go
   grpcHandler := &grpchandler.GrpcServer{
       ProfileRepo:      s.profileRepo,
       MenuRepo:         s.menuRepo,
       ExchangeRateRepo: s.exchangeRateRepo,
       InterestRateRepo: s.interestRateRepo,
       BranchRepo:       s.branchRepo,
   }
   ```

### Update `internal/server/router.go`

1. Add `searchHandler *handlers.SearchHandler` parameter to `setupRoutes`
2. Register new routes:
   ```go
   // Search / rates / branches routes
   r.Get("/api/exchange-rates", searchHandler.GetExchangeRates)
   r.Get("/api/interest-rates", searchHandler.GetInterestRates)
   r.Get("/api/branches",       searchHandler.GetBranches)
   ```

---

## File Change Summary

| File | Action | Notes |
|---|---|---|
| `internal/db/migrations/004_add_exchange_rates.sql` | **Create** | DDL for `exchange_rate` |
| `internal/db/migrations/005_add_interest_rates.sql` | **Create** | DDL for `interest_rate` |
| `internal/db/migrations/006_add_branches.sql` | **Create** | DDL for `branch` |
| `seed.sql` | **Update** | Add INSERT for 3 new tables |
| `internal/models/search.go` | **Create** | `ExchangeRate`, `InterestRate`, `Branch` structs |
| `internal/repository/exchange_rate.go` | **Create** | `GetAll` |
| `internal/repository/interest_rate.go` | **Create** | `GetAll` |
| `internal/repository/branch.go` | **Create** | `GetAll`, `SearchByName` |
| `internal/handlers/search.go` | **Create** | `SearchHandler` with 3 methods |
| `proto/user_profile_api.proto` | **Update** | Add 3 RPCs |
| `proto/user_profile_payload.proto` | **Update** | Add 6 new messages |
| `protogen/user-profile-service/user_profile_payload.pb.go` | **Update** | Add structs + getters |
| `protogen/user-profile-service/user_profile_api_grpc.pb.go` | **Update** | Add interface methods + ServiceDesc |
| `internal/grpchandler/profile.go` | **Update** | Add 3 repo fields to `GrpcServer` struct |
| `internal/grpchandler/converter.go` | **Update** | Add 3 conversion helpers |
| `internal/grpchandler/search.go` | **Create** | Implement 3 gRPC methods |
| `internal/server/server.go` | **Update** | Wire new repos + `SearchHandler` into `GrpcServer` |
| `internal/server/router.go` | **Update** | Add `searchHandler` param + 3 new routes |

**Total:** 8 new files, 10 modified files

---

## Verification Checklist

```bash
# 1. Build check — must compile with no errors
cd user-profile-service
go build ./...

# 2. Start stack
docker compose up --build

# 3. REST: Exchange Rates
curl http://localhost:8080/api/exchange-rates
# Expected: JSON array of 4 items (VND, NIO, KRW, CNY)

# 4. REST: Interest Rates
curl http://localhost:8080/api/interest-rates
# Expected: JSON array of 4 items

# 5. REST: Branches — all
curl http://localhost:8080/api/branches
# Expected: JSON array of 5 branches

# 6. REST: Branches — filtered
curl "http://localhost:8080/api/branches?q=union"
# Expected: JSON array of 2 branches (both "Bank 1656 Union Street")

# 7. REST: Branches — case-insensitive
curl "http://localhost:8080/api/branches?q=UNION"
# Expected: same 2 branches

# 8. gRPC: smoke test via grpcurl
grpcurl -plaintext localhost:9302 userprofile.UserProfileService/GetExchangeRates
grpcurl -plaintext localhost:9302 userprofile.UserProfileService/GetInterestRates
grpcurl -plaintext -d '{"query":"union"}' localhost:9302 userprofile.UserProfileService/GetBranches
```

---

## Design Decisions

| Decision | Choice | Reason |
|---|---|---|
| Response format | Raw JSON arrays | Matches api.txt spec; no wrapper envelope needed |
| Handler grouping | One `SearchHandler` struct | All 3 are "search/info" features; keeps files cohesive |
| gRPC exposure | Yes, all 3 endpoints | BFF may call these via gRPC later |
| Branch search SQL | `ILIKE '%' \|\| $1 \|\| '%'` | Case-insensitive partial match per api.txt spec |
| Auth | None (public endpoints) | api.txt spec has no auth on these endpoints |
| Empty `q` param | Returns all branches | Matches api.txt: "Omit or pass empty string to get all" |
| Seed data | From api.txt samples | Gives predictable test data matching spec examples |
| gRPC `GetBranchesRequest` | Includes optional `query string` | Mirrors REST `?q=` param for BFF flexibility |
