---
applyTo: "payment-service/**"
---

# payment-service

## Overview

Payment service for BankEase — internet bill providers, internet bill detail (JWT-protected), and currency exchange rates.
Module: `github.com/bankease/payment-service` | HTTP: 8082 | gRPC: 9304 | PostgreSQL: 5435 (DB: `payment`)

## Folder Structure

```
payment-service/
├── server/
│   ├── main.go                         # Entry point + CLI + HTTP/gRPC + cors + jwtMiddleware
│   ├── core_config.go                  # Config struct + initConfig (godotenv + GetEnv)
│   ├── core_db.go                      # DB connection + embed.FS migration
│   ├── api/
│   │   ├── api.go                      # Server struct + New() + compile-time check
│   │   ├── payment_api.go              # HTTP handlers (3 endpoints)
│   │   ├── payment_grpc_api.go         # gRPC handlers (3 RPCs)
│   │   ├── payment_authInterceptor.go  # JWT interceptor (skip public)
│   │   ├── payment_interceptor.go      # Chain: ProcessId → Logging → Errors → Auth
│   │   ├── converter.go               # DB model → proto
│   │   └── error.go                    # writeJSON, writeError
│   ├── db/
│   │   ├── provider.go                 # DB wrapper constructor
│   │   ├── bill_provider.go            # GetAllProviders, GetInternetBillByUserID
│   │   ├── currency_provider.go        # GetAllCurrencies
│   │   └── error.go                    # NotFoundErr
│   ├── jwt/manager.go                  # JWT Verify only (HS256)
│   ├── lib/database/                   # DB wrapper (identity-service pattern)
│   ├── lib/logger/                     # Zap + FluentBit
│   ├── constant/                       # ResponseCodeSuccess, ProcessIdCtx
│   └── utils/                          # GetEnv, GetProcessIdFromCtx
├── migrations/                         # 3 SQL + embed.go
├── proto/                              # payment_api.proto, payment_payload.proto
├── protogen/payment-service/           # codec.go + hand-written stubs
├── docs/                               # Swagger (docs.go, swagger.json, swagger.yaml)
├── seed.sql                            # 6 providers + 1 bill + 10 currencies
├── Dockerfile, docker-compose.yml, Makefile, sonar-project.properties
```

## DB Tables

- `provider` — id (VARCHAR PK), name
- `internet_bill` — id (UUID PK), user_id (UUID), customer_id, name, address, phone_number, code, bill_from, bill_to, internet_fee, tax, total
- `currency` — id (UUID PK), code, label, rate (NUMERIC)

## Auth Pattern

- `GET /api/pay-the-bill/providers` — **public**
- `GET /api/pay-the-bill/internet-bill` — **protected** (HTTP: jwtMiddleware in main.go; gRPC: authInterceptor)
- `GET /api/currency-list` — **public**

HTTP JWT middleware: main.go `jwtMiddleware()` → extracts Bearer token → `jwtMgr.Verify()` → `context.WithValue("user_claims")` → handler reads claims from ctx.

gRPC JWT interceptor: `payment_authInterceptor.go` → `accessibleRoles` map → only `GetInternetBill` restricted → `claimsToken()` extracts from metadata.

## Dependencies

Same as identity-service: `dgrijalva/jwt-go`, `urfave/cli`, `joho/godotenv`, `lib/pq`, `google.golang.org/grpc`, `go.uber.org/zap`, `swaggo/http-swagger`, `grpc-ecosystem/go-grpc-middleware`, `google/uuid`, `sirupsen/logrus`, `fluent/fluent-logger-golang`

## Key Patterns

- Domain type `ServiceProvider` (not `Provider`) to avoid conflict with `db.Provider` struct
- Response format: raw JSON (no envelope wrapper) — matches api.txt spec
- `internet_bill.user_id` is UUID cross-service link (not FK) — matched via JWT claims
- `codec.go` with `init()` RegisterCodec — MANDATORY for hand-written protogen
