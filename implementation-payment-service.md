# Implementation Plan — Payment Service

## Overview

Buat `payment-service` baru berdasarkan 3 endpoint dari `api.txt` (providers, internet-bill, currency-list), dengan folder structure mengikuti `identity-service`. Service punya JWT auth untuk endpoint internet-bill (data customer-spesifik). Termasuk integrasi ke BFF-service sebagai 4th downstream service (gRPC port 9304).

## Architecture

```
┌──────────────┐         REST (HTTP)          ┌─────────────────────────┐
│  Mobile App  │ ──────────────────────────▶  │   BFF Service           │
│  (React      │  GET /api/pay-the-bill/...   │   Port 3000 (HTTP)      │
│   Native)    │  GET /api/currency-list      │   Port 9090 (gRPC)      │
└──────────────┘                              └──────────┬──────────────┘
                                                         │ gRPC
                                                         ▼
                                              ┌──────────────────────┐
                                              │  payment-service     │
                                              │  Port 8082 (HTTP)    │
                                              │  Port 9304 (gRPC)    │
                                              └──────────┬───────────┘
                                                         ▼
                                              ┌──────────────────────┐
                                              │  PostgreSQL          │
                                              │  Port 5435           │
                                              │  DB: payment         │
                                              └──────────────────────┘
```

- **Module path**: `github.com/bankease/payment-service`
- **HTTP**: 8082 | **gRPC**: 9304 | **PostgreSQL**: 5435 (DB name: `payment`)
- **Auth**: JWT verify lokal (verify only, tidak generate)
  - `GET /api/pay-the-bill/providers` — **public**
  - `GET /api/pay-the-bill/internet-bill` — **protected** (JWT required, return data berdasarkan user)
  - `GET /api/currency-list` — **public**
- **Response format**: raw JSON array/object (tidak di-wrap envelope), sesuai api.txt spec

---

## API Endpoints

### 1. GET /api/pay-the-bill/providers

Returns a list of internet service providers.

**Auth**: Public (no JWT required)

**Response 200 OK** — `Provider[]`

| Field | Type   | Description       |
|-------|--------|-------------------|
| id    | string | Unique identifier |
| name  | string | Provider name     |

```json
[
  { "id": "1", "name": "Biznet" },
  { "id": "2", "name": "Indihome" },
  { "id": "3", "name": "MyRepublic" },
  { "id": "4", "name": "XL Home" },
  { "id": "5", "name": "CBN" },
  { "id": "6", "name": "First Media" }
]
```

### 2. GET /api/pay-the-bill/internet-bill

Returns the internet bill detail for the current customer (identified via JWT).

**Auth**: Protected (JWT Bearer token required — `user_id` extracted from claims)

**Response 200 OK** — `InternetBillDetail`

| Field       | Type   | Description                              |
|-------------|--------|------------------------------------------|
| customerId  | string | Customer account ID                      |
| name        | string | Customer full name                       |
| address     | string | Customer address                         |
| phoneNumber | string | Customer phone number                    |
| code        | string | Bill code                                |
| from        | string | Billing period start (DD/MM/YYYY)        |
| to          | string | Billing period end (DD/MM/YYYY)          |
| internetFee | string | Internet service charge (e.g. "$50")     |
| tax         | string | Tax amount                               |
| total       | string | Total amount due                         |

```json
{
  "customerId":  "#2345641ASS",
  "name":        "Jackson Maine",
  "address":     "403 East 4th Street, Santa Ana",
  "phoneNumber": "+8424599721",
  "code":        "#2345641",
  "from":        "01/09/2019",
  "to":          "01/10/2019",
  "internetFee": "$50",
  "tax":         "$0",
  "total":       "$50"
}
```

**Response 401 Unauthorized** — Missing or invalid JWT

```json
{ "error": true, "code": 401, "message": "Unauthorized" }
```

### 3. GET /api/currency-list

Returns a list of supported currencies with their exchange rates relative to USD.

**Auth**: Public (no JWT required)

**Response 200 OK** — `CurrencyEntry[]`

| Field | Type   | Description                              |
|-------|--------|------------------------------------------|
| code  | string | ISO 4217 currency code (e.g. "USD")      |
| label | string | Display label (e.g. "USD (United States Dollar)") |
| rate  | number | Exchange rate relative to USD (USD = 1)  |

```json
[
  { "code": "AUD", "label": "AUD (Australian Dollar)",      "rate": 1.53   },
  { "code": "CNY", "label": "CNY (Chinese Yuan)",           "rate": 7.24   },
  { "code": "EUR", "label": "EUR (Euro)",                   "rate": 0.92   },
  { "code": "GBP", "label": "GBP (British Pound Sterling)", "rate": 0.79   },
  { "code": "IDR", "label": "IDR (Indonesian Rupiah)",      "rate": 16350  },
  { "code": "JPY", "label": "JPY (Japanese Yen)",           "rate": 149.5  },
  { "code": "MYR", "label": "MYR (Malaysian Ringgit)",      "rate": 4.72   },
  { "code": "SAR", "label": "SAR (Saudi Riyal)",            "rate": 3.75   },
  { "code": "SGD", "label": "SGD (Singapore Dollar)",       "rate": 1.34   },
  { "code": "USD", "label": "USD (United States Dollar)",   "rate": 1      }
]
```

---

## Database Schema

### Table: `provider`

```sql
CREATE TABLE IF NOT EXISTS provider (
    id   VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
```

### Table: `internet_bill`

```sql
CREATE TABLE IF NOT EXISTS internet_bill (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    customer_id  VARCHAR(50) NOT NULL,
    name         VARCHAR(100) NOT NULL,
    address      TEXT NOT NULL,
    phone_number VARCHAR(50) NOT NULL,
    code         VARCHAR(50) NOT NULL,
    bill_from    VARCHAR(20) NOT NULL,
    bill_to      VARCHAR(20) NOT NULL,
    internet_fee VARCHAR(50) NOT NULL,
    tax          VARCHAR(50) NOT NULL,
    total        VARCHAR(50) NOT NULL
);
```

### Table: `currency`

```sql
CREATE TABLE IF NOT EXISTS currency (
    id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code  VARCHAR(10) NOT NULL,
    label VARCHAR(100) NOT NULL,
    rate  NUMERIC(15,4) NOT NULL
);
```

---

## Folder Structure

```
payment-service/
├── server/
│   ├── main.go                         # Entry point + CLI + HTTP/gRPC + routes + cors
│   ├── core_config.go                  # Config struct + initConfig (godotenv + GetEnv)
│   ├── core_db.go                      # DB connection lifecycle (database.InitConnectionDB + retry)
│   ├── api/
│   │   ├── api.go                      # Server struct + constructor + compile-time check
│   │   ├── payment_api.go              # HTTP: HandleGetProviders, HandleGetInternetBill, HandleGetCurrencyList
│   │   ├── payment_grpc_api.go         # gRPC: GetProviders, GetInternetBill, GetCurrencyList
│   │   ├── payment_authInterceptor.go  # JWT auth interceptor (skip public endpoints)
│   │   ├── payment_interceptor.go      # Chain: ProcessId → Logging → Errors → Auth
│   │   ├── converter.go               # DB model → proto conversion helpers
│   │   └── error.go                    # writeJSON, writeError, standardResponse
│   ├── db/
│   │   ├── provider.go                 # Provider struct + constructor (DB wrapper)
│   │   ├── bill_provider.go            # GetAllProviders, GetInternetBillByUserID + domain types
│   │   ├── currency_provider.go        # GetAllCurrencies + Currency type
│   │   └── error.go                    # NotFoundErr type
│   ├── jwt/
│   │   └── manager.go                  # JWT Verify only (HS256, local verification)
│   ├── lib/
│   │   ├── database/                   # DB wrapper (database.go, wrapper/, mock/)
│   │   └── logger/                     # Zap structured logger + FluentBit
│   ├── constant/
│   │   ├── constant.go                 # ResponseCodeSuccess, date formats
│   │   └── process_id.go              # ProcessIdKey
│   └── utils/
│       └── utils.go                    # GetEnv, GetProcessIdFromCtx, GenerateProcessId
├── migrations/
│   ├── 001_add_providers.sql
│   ├── 002_add_internet_bills.sql
│   ├── 003_add_currencies.sql
│   └── embed.go                        # embed.FS for SQL migration files
├── proto/
│   ├── payment_api.proto               # Service: PaymentService (3 RPCs)
│   └── payment_payload.proto           # Messages: Provider, InternetBill, Currency
├── protogen/
│   └── payment-service/
│       ├── codec.go                    # JSONCodec + init() RegisterCodec — WAJIB
│       ├── payment_api_grpc.pb.go      # Hand-written gRPC service interface + Register func
│       └── payment_payload.pb.go       # Hand-written message structs (JSON tags)
├── docs/
│   ├── docs.go                         # Swagger generated
│   ├── swagger.json
│   └── swagger.yaml
├── seed.sql                            # 6 providers + 1 internet_bill + 10 currencies
├── Dockerfile                          # Multi-stage: golang:1.24-alpine → alpine:3.20
├── docker-compose.yml                  # Prod-like
├── docker-compose.local.yml            # Dev: PostgreSQL 17 (port 5435) + payment-service
├── Makefile                            # build, run, unit-test, docker-build
├── sonar-project.properties
├── .env.example
└── go.mod
```

---

## Implementation Phases

### Phase 1 — Foundation

> *All steps in this phase can be done in parallel.*

| # | File | Detail |
|---|------|--------|
| 1 | `go.mod` | Module `github.com/bankease/payment-service`, Go 1.24, deps = saving-service + `dgrijalva/jwt-go` |
| 2 | `proto/payment_api.proto` | Service `PaymentService`: `GetProviders`, `GetInternetBill`, `GetCurrencyList` |
| 3 | `proto/payment_payload.proto` | Request/response messages untuk 3 RPCs |
| 4 | `protogen/payment-service/codec.go` | JSONCodec + `init()` RegisterCodec — copy dari identity-service |
| 5 | `protogen/payment-service/payment_api_grpc.pb.go` | Interface `PaymentServiceServer`, `RegisterPaymentServiceServer`, `UnimplementedPaymentServiceServer` |
| 6 | `protogen/payment-service/payment_payload.pb.go` | Request/Response structs dengan JSON tags |
| 7 | `migrations/001_add_providers.sql` | DDL: tabel `provider` |
| 8 | `migrations/002_add_internet_bills.sql` | DDL: tabel `internet_bill` |
| 9 | `migrations/003_add_currencies.sql` | DDL: tabel `currency` |
| 10 | `migrations/embed.go` | embed.FS untuk migration files |
| 11 | `server/lib/database/` | Copy dari identity-service (database.go, wrapper/, mock/) |
| 12 | `server/lib/logger/` | Copy dari identity-service (logger.go, fluentbit.go) |
| 13 | `server/constant/constant.go` | ResponseCodeSuccess, date formats |
| 14 | `server/constant/process_id.go` | ProcessIdKey |
| 15 | `server/utils/utils.go` | GetEnv, GetProcessIdFromCtx, GenerateProcessId |
| 16 | `seed.sql` | 6 providers + 1 sample internet_bill + 10 currencies |
| 17 | `.env.example` | DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, JWT_SECRET, etc. |

### Phase 2 — Core Infrastructure

> *Depends on Phase 1*

| # | File | Detail |
|---|------|--------|
| 18 | `server/core_config.go` | Config struct (DB + JWT_SECRET + logger + CORS), `initConfig()`, `GetEnv()` — pattern dari identity-service, tambah `JWTSecret` field |
| 19 | `server/core_db.go` | `startDBConnection()`, `closeDBConnection()`, `initDBMain()` (database wrapper + retry), `runMigration()` via embed.FS |
| 20 | `server/jwt/manager.go` | JWT Verify only — `Verify(token) → claims, error` (copy pattern dari bff-service) |

### Phase 3 — Data Layer

> *Depends on Phase 2*

| # | File | Detail |
|---|------|--------|
| 21 | `server/db/provider.go` | `type Provider struct`, `func New(db, logger) *Provider` — database wrapper style |
| 22 | `server/db/error.go` | `type NotFoundErr struct`, `func (e *NotFoundErr) Error() string` |
| 23 | `server/db/bill_provider.go` | `type ServiceProvider struct { ID, Name }`, `type InternetBill struct { ... }`, `GetAllProviders()`, `GetInternetBillByUserID()` |
| 24 | `server/db/currency_provider.go` | `type Currency struct { Code, Label, Rate }`, `GetAllCurrencies()` |

### Phase 4 — API Layer

> *Depends on Phase 3*

| # | File | Detail |
|---|------|--------|
| 25 | `server/api/error.go` | `writeJSON`, `writeError`, `standardResponse`, `badRequestError`, `unauthorizedError`, `serverError` |
| 26 | `server/api/converter.go` | `providersToProto`, `internetBillToProto`, `currenciesToProto` — DB models → proto messages |
| 27 | `server/api/api.go` | `type Server struct { provider, manager, logger }`, `var _ pb.PaymentServiceServer = (*Server)(nil)`, `func New(...)` |
| 28 | `server/api/payment_api.go` | HTTP handlers: `HandleGetProviders` (public), `HandleGetInternetBill` (protected — extract userID dari JWT context), `HandleGetCurrencyList` (public) |
| 29 | `server/api/payment_grpc_api.go` | gRPC handlers: `GetProviders`, `GetInternetBill` (extract userID dari ctx), `GetCurrencyList` |
| 30 | `server/api/payment_authInterceptor.go` | JWT auth interceptor — skip: `GetProviders` + `GetCurrencyList`; require: `GetInternetBill` |
| 31 | `server/api/payment_interceptor.go` | Chain: ProcessId → Logging → Errors → Auth. `UnaryInterceptors()`, `StreamInterceptors()` |

### Phase 5 — Server Entry Point

> *Depends on Phase 4*

| # | File | Detail |
|---|------|--------|
| 32 | `server/main.go` | `init()` → initConfig + logger. `main()` → CLI app (`grpc-gw-server`). Routes: `GET /api/pay-the-bill/providers`, `GET /api/pay-the-bill/internet-bill` (JWT middleware), `GET /api/currency-list`, `GET /health`, `/swagger/`. `cors()`, `methodOnly()` helpers. HTTP JWT middleware untuk internet-bill endpoint |

### Phase 6 — Docker & Config

> *Parallel with Phase 5*

| # | File | Detail |
|---|------|--------|
| 33 | `Dockerfile` | Multi-stage: `golang:1.24-alpine` build → `alpine:3.20` runtime, `go build -o /server ./server` |
| 34 | `docker-compose.local.yml` | `payment-db`: postgres:17-alpine (port 5435, DB=payment, seed.sql). `payment-service`: command `./server grpc-gw-server --port1 9304 --port2 8082` |
| 35 | `docker-compose.yml` | Prod-like (sama structure, tanpa port mapping eksternal) |
| 36 | `Makefile` | Targets: `build`, `run`, `test`, `coverage`, `docker-build`, `docker-up` |
| 37 | `sonar-project.properties` | `sonar.projectKey=bricams-addons-payment-service:project` |

### Phase 7 — Swagger Docs

> *Depends on Phase 5*

| # | File | Detail |
|---|------|--------|
| 38 | HTTP handlers | Tambahkan Swagger annotations (`@Summary`, `@Tags`, `@Produce`, `@Success`, `@Failure`, `@Router`) |
| 39 | `docs/` | Generate via `swag init -g server/main.go -o docs/` → `docs.go`, `swagger.json`, `swagger.yaml` |

### Phase 8 — BFF Integration

> *Depends on Phase 5, parallel with Phase 7*

| # | File | Detail |
|---|------|--------|
| 40 | `bff-service/protogen/payment-service/payment_api_grpc.pb.go` | Hand-written **client** stubs: `PaymentServiceClient` interface + `NewPaymentServiceClient` |
| 41 | `bff-service/protogen/payment-service/payment_payload.pb.go` | Same message structs as payment-service protogen (copied) |
| 42 | `bff-service/server/core_config.go` | Add `PaymentServiceAddr` config (default: `localhost:9304`) |
| 43 | `bff-service/server/services/service.go` | Add `PaymentService` gRPC client connection: `paymentConn`, `GetPaymentServiceClient()` |
| 44 | `bff-service/server/api/bff_payment_api.go` | 3 proxy handlers: `GetProviders`, `GetInternetBill`, `GetCurrencyList` |
| 45 | `bff-service/server/http_routes.go` | 3 new routes: `/api/pay-the-bill/providers`, `/api/pay-the-bill/internet-bill`, `/api/currency-list` |
| 46 | `bff-service/docker-compose.yml` | Add `payment-db` + `payment-service` containers (total 7 containers) |
| 47 | `bff-service/proto/bff_api.proto` | Add 3 new RPCs: `GetProviders`, `GetInternetBill`, `GetCurrencyList` |
| 48 | `bff-service/proto/bff_payload.proto` | Add payment-related request/response messages |
| 49 | `bff-service/protogen/bff-service/bff_api_grpc.pb.go` | Add payment RPC methods to BFF server interface |
| 50 | `bff-service/protogen/bff-service/bff_payload.pb.go` | Add payment-related message structs |

### Phase 9 — Verification

| # | Check | Expected |
|---|-------|----------|
| 51 | `cd payment-service && go build ./server/` | Compile pass |
| 52 | `cd payment-service && go vet ./server/...` | No warnings |
| 53 | `docker compose -f docker-compose.local.yml up --build -d` | payment-db + payment-service healthy |
| 54 | `curl http://localhost:8082/health` | `{"status":"ok"}` |
| 55 | `curl http://localhost:8082/api/pay-the-bill/providers` | JSON array, 6 items |
| 56 | `curl http://localhost:8082/api/pay-the-bill/internet-bill` | 401 Unauthorized |
| 57 | `curl -H "Authorization: Bearer <jwt>" http://localhost:8082/api/pay-the-bill/internet-bill` | JSON object (internet bill detail) |
| 58 | `curl http://localhost:8082/api/currency-list` | JSON array, 10 items |
| 59 | `cd bff-service && go build ./server/` | Compile pass with new payment protogen |
| 60 | `cd bff-service && docker compose up --build` | All 7 containers UP (4 DBs + 4 services minus BFF stateless = 3 DBs + 4 services) |
| 61 | `curl http://localhost:3000/api/pay-the-bill/providers` | Proxied response from payment-service |
| 62 | `curl http://localhost:3000/api/currency-list` | Proxied response from payment-service |

---

## Proto Definitions

### payment_api.proto

```protobuf
syntax = "proto3";
package payment;
option go_package = "github.com/bankease/payment-service/protogen/payment-service";

import "payment_payload.proto";

service PaymentService {
  rpc GetProviders(GetProvidersRequest) returns (ProviderListResponse);
  rpc GetInternetBill(GetInternetBillRequest) returns (InternetBillResponse);
  rpc GetCurrencyList(GetCurrencyListRequest) returns (CurrencyListResponse);
}
```

### payment_payload.proto

```protobuf
syntax = "proto3";
package payment;
option go_package = "github.com/bankease/payment-service/protogen/payment-service";

message GetProvidersRequest {}
message GetInternetBillRequest {}
message GetCurrencyListRequest {}

message Provider {
  string id = 1;
  string name = 2;
}

message InternetBillDetail {
  string customer_id = 1;
  string name = 2;
  string address = 3;
  string phone_number = 4;
  string code = 5;
  string from = 6;
  string to = 7;
  string internet_fee = 8;
  string tax = 9;
  string total = 10;
}

message CurrencyEntry {
  string code = 1;
  string label = 2;
  double rate = 3;
}

message ProviderListResponse {
  repeated Provider providers = 1;
}

message InternetBillResponse {
  InternetBillDetail bill = 1;
}

message CurrencyListResponse {
  repeated CurrencyEntry currencies = 1;
}
```

---

## gRPC Service Definition

| RPC | Request | Response | Auth |
|-----|---------|----------|------|
| `GetProviders` | `GetProvidersRequest` (empty) | `ProviderListResponse` | Public |
| `GetInternetBill` | `GetInternetBillRequest` (empty — user_id from JWT context) | `InternetBillResponse` | Protected (JWT) |
| `GetCurrencyList` | `GetCurrencyListRequest` (empty) | `CurrencyListResponse` | Public |

---

## Seed Data

```sql
-- Providers (6 items)
INSERT INTO provider (id, name) VALUES
  ('1', 'Biznet'),
  ('2', 'Indihome'),
  ('3', 'MyRepublic'),
  ('4', 'XL Home'),
  ('5', 'CBN'),
  ('6', 'First Media');

-- Sample Internet Bill (1 item — linked to a test user_id)
INSERT INTO internet_bill (user_id, customer_id, name, address, phone_number, code, bill_from, bill_to, internet_fee, tax, total) VALUES
  ('00000000-0000-0000-0000-000000000001', '#2345641ASS', 'Jackson Maine', '403 East 4th Street, Santa Ana', '+8424599721', '#2345641', '01/09/2019', '01/10/2019', '$50', '$0', '$50');

-- Currencies (10 items)
INSERT INTO currency (code, label, rate) VALUES
  ('AUD', 'AUD (Australian Dollar)',        1.53),
  ('CNY', 'CNY (Chinese Yuan)',             7.24),
  ('EUR', 'EUR (Euro)',                     0.92),
  ('GBP', 'GBP (British Pound Sterling)',   0.79),
  ('IDR', 'IDR (Indonesian Rupiah)',        16350),
  ('JPY', 'JPY (Japanese Yen)',             149.5),
  ('MYR', 'MYR (Malaysian Ringgit)',        4.72),
  ('SAR', 'SAR (Saudi Riyal)',              3.75),
  ('SGD', 'SGD (Singapore Dollar)',         1.34),
  ('USD', 'USD (United States Dollar)',     1);
```

---

## Reference Files (Pattern Sources)

### Payment Service — New Files

| File | Pattern Source |
|------|---------------|
| `server/main.go` | `identity-service/server/main.go` (CLI + HTTP + gRPC + cors) |
| `server/core_config.go` | `identity-service/server/core_config.go` (Config + initConfig + JWTSecret) |
| `server/core_db.go` | `identity-service/server/core_db.go` (database wrapper + retry + embed.FS migration) |
| `server/api/api.go` | `identity-service/server/api/api.go` (Server struct + New() + compile-time check) |
| `server/api/payment_api.go` | `saving-service/server/api/saving_api.go` (HTTP GET handlers + writeJSON) |
| `server/api/payment_grpc_api.go` | `saving-service/server/api/saving_grpc_api.go` (gRPC handlers + model→proto) |
| `server/api/payment_authInterceptor.go` | `identity-service/server/api/identity_authInterceptor.go` (JWT interceptor) |
| `server/api/payment_interceptor.go` | `identity-service/server/api/identity_interceptor.go` (interceptor chain) |
| `server/api/error.go` | `saving-service/server/api/saving_api.go` (writeJSON, writeError) |
| `server/api/converter.go` | New — DB model → proto conversion |
| `server/db/provider.go` | `identity-service/server/db/provider.go` (database wrapper style) |
| `server/db/bill_provider.go` | New — queries for providers + internet_bill |
| `server/db/currency_provider.go` | New — queries for currency |
| `server/db/error.go` | `identity-service/server/db/error.go` (NotFoundErr) |
| `server/jwt/manager.go` | `bff-service/server/jwt/manager.go` (Verify-only) |
| `server/lib/` | Copy from `identity-service/server/lib/` |
| `server/constant/` | Copy from `identity-service/server/constant/` |
| `server/utils/` | Copy from `identity-service/server/utils/` |
| `protogen/payment-service/codec.go` | `identity-service/protogen/identity-service/codec.go` |
| `protogen/payment-service/*.pb.go` | `saving-service/protogen/saving-service/*.pb.go` (hand-written) |
| `Dockerfile` | `saving-service/Dockerfile` |
| `docker-compose.local.yml` | `saving-service/docker-compose.local.yml` (ports: 5435, 8082, 9304) |

### BFF Service — Modified Files

| File | Change |
|------|--------|
| `bff-service/protogen/payment-service/` | New directory — client stubs + payload structs |
| `bff-service/server/core_config.go` | Add `PaymentServiceAddr` config |
| `bff-service/server/services/service.go` | Add PaymentService gRPC client connection |
| `bff-service/server/api/bff_payment_api.go` | New — 3 proxy handlers |
| `bff-service/server/http_routes.go` | Add 3 new routes |
| `bff-service/docker-compose.yml` | Add payment-db + payment-service containers |
| `bff-service/proto/bff_api.proto` | Add 3 new RPCs |
| `bff-service/proto/bff_payload.proto` | Add payment-related messages |
| `bff-service/protogen/bff-service/bff_api_grpc.pb.go` | Add payment RPC methods |
| `bff-service/protogen/bff-service/bff_payload.pb.go` | Add payment message structs |

---

## Architecture Decisions

| Keputusan | Alasan |
|-----------|--------|
| Module path `github.com/bankease/payment-service` | Konsisten dengan saving-service naming convention |
| Ports: HTTP 8082, gRPC 9304, PostgreSQL 5435 | Next available dalam sequence (saving: 8081/9303/5434) |
| Auth hybrid: internet-bill protected, lainnya public | internet-bill mengembalikan data customer-spesifik berdasarkan user_id dari JWT |
| DB wrapper dari identity-service `lib/database/` | Consistency across services (retry logic, connection pooling) |
| Migration via embed.FS | Lebih maintainable daripada inline SQL (saving-service pattern) |
| Response format: raw JSON (no envelope) | Sesuai api.txt spec, sama dengan saving-service |
| `internet_bill.user_id` UUID (bukan FK) | Cross-service link ke identity-service `users.id`, matched via JWT claims |
| Domain type `ServiceProvider` (bukan `Provider`) | Menghindari naming conflict dengan `db.Provider` struct |
| BFF = 4th downstream | identity (9301) + user-profile (9302) + saving (9303) + payment (9304) |
| Unit tests excluded dari plan ini | Akan ditambahkan setelah implementasi mengikuti go-unit-test skill |

---

## Notes

1. **Internet bill data model**: Menyimpan formatted strings ("$50") sesuai api.txt spec. Jika di masa depan app perlu kalkulasi, pertimbangkan storing sebagai numeric + formatting di response layer. Untuk sekarang: keep as string (match MSW mock behavior).

2. **`provider` table vs `db.Provider` struct**: Nama tabel `provider` di database, tapi domain type di Go code dinamakan `ServiceProvider` untuk menghindari konflik. Alternatif: rename tabel ke `bill_provider`.

3. **JWT Secret**: Harus sama dengan identity-service dan BFF (HS256, shared secret). Set via environment variable `JWT_SECRET`.
