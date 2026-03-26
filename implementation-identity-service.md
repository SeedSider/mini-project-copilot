# Detail Implementasi — identity-service

## Ringkasan

Service ini menyediakan fitur **Sign Up**, **Sign In**, dan **Get Me** (profil user) untuk platform BRICaMS mobile. Dibangun menggunakan Go dengan arsitektur dan pattern yang mengacu pada `addons-issuance-lc-service`.

---

## Stack Teknologi

| Komponen      | Teknologi                             |
| ------------- | ------------------------------------- |
| Bahasa        | Go 1.24                               |
| HTTP Server   | `net/http` + CORS middleware          |
| Database      | PostgreSQL 15                         |
| Auth          | JWT (HS256) via `dgrijalva/jwt-go`    |
| Password Hash | bcrypt (`golang.org/x/crypto/bcrypt`) |
| Logger        | Zap + FluentBit                       |
| CLI           | `urfave/cli`                          |
| Config        | `joho/godotenv` + `os.LookupEnv`      |
| Container     | Docker + Docker Compose               |
| Code Quality  | SonarQube                             |

---

## Struktur Folder

```
identity-service/
├── migrations/
│   └── 001_init.sql                    # DDL tabel users & profiles
├── proto/
│   ├── identity_api.proto              # Definisi service (SignUp, SignIn, GetMe)
│   └── identity_payload.proto          # Request/response messages
├── server/
│   ├── main.go                         # Entry point + CLI + HTTP server + auto-migration
│   ├── core_config.go                  # Config loader (env vars + .env)
│   ├── core_db.go                      # DB connection lifecycle
│   ├── api/
│   │   ├── api.go                      # Server struct + constructor
│   │   ├── identity_auth_api.go        # Handler: SignUp, SignIn, GetMe + HTTP handlers
│   │   ├── identity_authInterceptor.go # JWT auth interceptor (untuk GetMe)
│   │   ├── identity_interceptor.go     # Interceptor chain: ProcessId → Logging → Errors → Auth
│   │   └── error.go                    # Helper error (badRequest, unauthorized, conflict, dll)
│   ├── db/
│   │   ├── provider.go                 # Provider struct + constructor
│   │   ├── identity_provider.go        # Query: CreateUserWithProfile, GetUserByEmail, dll
│   │   └── error.go                    # NotFoundErr type
│   ├── jwt/
│   │   └── manager.go                  # JWT Generate + Verify (HS256)
│   ├── lib/
│   │   ├── database/
│   │   │   ├── database.go            # DB wrapper (connect, retry, transaction)
│   │   │   ├── wrapper/               # Interface + wrapper untuk sql.DB
│   │   │   └── mock/                  # Mock DB untuk unit test
│   │   └── logger/
│   │       ├── logger.go              # Zap structured logger
│   │       └── fluentbit.go           # FluentBit hook
│   ├── utils/
│   │   └── utils.go                   # GetProcessIdFromCtx, GetEnv, GenerateProcessId
│   └── constant/
│       ├── constant.go                # Format date, response code
│       └── process_id.go             # ProcessIdCtx key
├── www/
│   └── swagger.json                   # Swagger API documentation
├── .env.example                       # Template environment variables
├── docker-compose.yml                 # PostgreSQL + identity-service
├── Dockerfile                         # Multi-stage build (golang → alpine)
├── Makefile                           # build, run, unit-test, docker-build
├── sonar-project.properties           # SonarQube config
├── go.mod / go.sum                    # Go module dependencies
└── backend-spec.md                    # Spesifikasi API
```

---

## Database Schema

### Tabel `users`

| Kolom         | Tipe         | Constraint                      |
| ------------- | ------------ | ------------------------------- |
| id            | UUID         | PRIMARY KEY, DEFAULT gen_random |
| email         | VARCHAR(255) | UNIQUE NOT NULL                 |
| password_hash | TEXT         | NOT NULL                        |
| created_at    | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()          |

### Tabel `profiles`

| Kolom      | Tipe         | Constraint                              |
| ---------- | ------------ | --------------------------------------- |
| id         | UUID         | PRIMARY KEY, DEFAULT gen_random         |
| user_id    | UUID         | UNIQUE NOT NULL, FK → users(id) CASCADE |
| full_name  | VARCHAR(255) | NOT NULL                                |
| phone      | VARCHAR(50)  | nullable                                |
| created_at | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()                  |

Migrasi dijalankan **otomatis** saat server startup (`CREATE TABLE IF NOT EXISTS`).

---

## API Endpoints

### 1. POST `/api/auth/signup`

Registrasi user baru dan otomatis membuat profile.

**Request:**

```json
{
  "email": "user@email.com",
  "password": "password123",
  "full_name": "John Doe",
  "phone": "08123456789"
}
```

**Response 201:**

```json
{
  "user_id": "uuid",
  "email": "user@email.com",
  "full_name": "John Doe"
}
```

**Error:**
| Status | Kondisi |
| ------ | -------------------- |
| 409 | Email sudah terdaftar |
| 422 | Validation error |

---

### 2. POST `/api/auth/signin`

Login dengan email & password, mendapat JWT token.

**Request:**

```json
{
  "email": "user@email.com",
  "password": "password123"
}
```

**Response 200:**

```json
{
  "user_id": "uuid",
  "email": "user@email.com",
  "full_name": "John Doe",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Error:**
| Status | Kondisi |
| ------ | -------------------- |
| 401 | Invalid credentials |

---

### 3. GET `/api/identity/me`

Mengambil profil user berdasarkan JWT token.

**Header:**

```
Authorization: Bearer <token>
```

**Response 200:**

```json
{
  "user_id": "uuid",
  "email": "user@email.com",
  "full_name": "John Doe",
  "phone": "08123456789"
}
```

**Error:**
| Status | Kondisi |
| ------ | -------------------- |
| 401 | Token tidak valid |

---

### 4. GET `/health`

Health check endpoint.

**Response 200:**

```json
{
  "status": "ok"
}
```

---

## Business Logic

### Sign Up Flow

```
1. Validasi input (email format, password ≥ 6 char, full_name wajib)
2. Cek apakah email sudah terdaftar di DB
3. Hash password menggunakan bcrypt (DefaultCost)
4. BEGIN transaction
   4a. INSERT ke tabel users → dapat user_id
   4b. INSERT ke tabel profiles (user_id, full_name, phone)
5. COMMIT transaction
6. Return response (user_id, email, full_name)
```

### Sign In Flow

```
1. Validasi input (email & password wajib)
2. Query user dari DB berdasarkan email
3. Compare password hash (bcrypt)
4. Query profile berdasarkan user_id
5. Generate JWT token (HS256, exp: 24h)
6. Return response (user_id, email, full_name, token)
```

### Get Me Flow

```
1. Ambil token dari header Authorization (Bearer <token>)
2. Verify & parse JWT → extract claims (user_id, email)
3. Query profile dari DB berdasarkan user_id
4. Return response (user_id, email, full_name, phone)
```

---

## Validation Rules

| Field     | Rule                      |
| --------- | ------------------------- |
| email     | Wajib, format email valid |
| password  | Wajib, minimal 6 karakter |
| full_name | Wajib, tidak boleh kosong |
| phone     | Opsional                  |

---

## Pattern yang Diadopsi dari LC Service

### 1. Provider Pattern (Database Layer)

Semua operasi database ada di `server/db/`, terbungkus dalam struct `Provider`:

```go
type Provider struct {
    dbSql *database.DbSql
}
```

Method: `CreateUserWithProfile`, `GetUserByEmail`, `GetProfileByUserID`, `CheckEmailExists`

### 2. Interceptor Chain

Urutan interceptor (dipertahankan dari LC service):

```
ProcessIdInterceptor → LoggingInterceptor → ErrorsInterceptor → AuthInterceptor
```

- **ProcessId**: Generate/propagate UUID untuk request tracing
- **Logging**: Log method, duration, dan error setiap request
- **Errors**: Mapping `NotFoundErr` ke gRPC `NotFound`
- **Auth**: Validasi JWT token untuk endpoint yang restricted

### 3. JWT Manager

```go
type JWTManager struct {
    secretKey     string
    tokenDuration time.Duration
}
```

- `Generate(userID, email, fullName)` → signed JWT string
- `Verify(accessToken)` → `*UserClaims`

### 4. Config Loading

Environment variables dimuat via `godotenv` + `os.LookupEnv` dengan fallback default.

### 5. Database Wrapper

Custom wrapper di `server/lib/database/` dengan:

- Connection string builder
- Retry logic (`MaxRetry`)
- Connection pooling (`SetMaxIdleConns`, `SetMaxOpenConns`)
- Interface-based untuk testability (mock tersedia)

### 6. Structured Logger

Zap-based logger dengan output ke stdout atau FluentBit (Elastic).

### 7. Error Handling

gRPC status codes di-mapping ke HTTP status codes:

| gRPC Code        | HTTP Status |
| ---------------- | ----------- |
| InvalidArgument  | 422         |
| Unauthenticated  | 401         |
| AlreadyExists    | 409         |
| NotFound         | 404         |
| PermissionDenied | 403         |
| Internal         | 500         |

### 8. Security Headers (CORS Middleware)

```
Strict-Transport-Security: max-age=31536000
Content-Security-Policy: object-src 'none'; child-src 'none'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: no-referrer
```

---

## Cara Menjalankan

### Development (Lokal)

```bash
# 1. Copy env
cp .env.example .env

# 2. Pastikan PostgreSQL jalan di localhost:5432

# 3. Jalankan service
make run
# atau
go run ./server/ grpc-gw-server --port1 9301 --port2 3031 --grpc-endpoint :9301
```

Server berjalan di `http://localhost:3031`

### Docker Compose

```bash
docker-compose up --build
```

Service berjalan di `http://localhost:3031`, PostgreSQL di port `5432`.

### Build Binary

```bash
make build
# Output: ./app
```

### Unit Test

```bash
make unit-test
```

---

## Environment Variables

| Variable       | Default          | Deskripsi                      |
| -------------- | ---------------- | ------------------------------ |
| DB_HOST        | localhost        | Host database                  |
| DB_PORT        | 5432             | Port database                  |
| DB_USER        | postgres         | User database                  |
| DB_PASSWORD    | postgres         | Password database              |
| DB_NAME        | identity         | Nama database                  |
| DB_SSLMODE     | disable          | Mode SSL                       |
| DB_TIMEZONE    | Asia/Jakarta     | Timezone                       |
| DB_MAX_RETRY   | 3                | Max retry koneksi DB           |
| DB_TIMEOUT     | 300              | Timeout koneksi (detik)        |
| JWT_SECRET     | secret           | Secret key untuk JWT           |
| JWT_DURATION   | 24h              | Durasi token JWT               |
| ENV            | DEV              | Environment (DEV/PROD)         |
| APP_NAME       | identity-service | Nama aplikasi                  |
| PRODUCT_NAME   | BRICaMS          | Nama produk                    |
| LOGGER_TAG     | identity.dev     | Tag logger                     |
| LOGGER_OUTPUT  | stdout           | Output logger (stdout/elastic) |
| LOGGER_LEVEL   | debug            | Level log                      |
| FLUENTBIT_HOST | 0.0.0.0          | Host FluentBit                 |
| FLUENTBIT_PORT | 24223            | Port FluentBit                 |

---

## File-file Penting

| File                                     | Fungsi                                        |
| ---------------------------------------- | --------------------------------------------- |
| `server/main.go`                         | Entry point, CLI, HTTP server, auto-migration |
| `server/core_config.go`                  | Load semua config dari env vars               |
| `server/core_db.go`                      | Lifecycle koneksi PostgreSQL                  |
| `server/api/api.go`                      | Struct `Server` + constructor                 |
| `server/api/identity_auth_api.go`        | **Handler utama**: SignUp, SignIn, GetMe      |
| `server/api/identity_authInterceptor.go` | JWT auth interceptor                          |
| `server/api/identity_interceptor.go`     | Chain: ProcessId → Logging → Errors → Auth    |
| `server/api/error.go`                    | Helper error methods                          |
| `server/db/provider.go`                  | Provider struct                               |
| `server/db/identity_provider.go`         | Query DB: create user, get by email, dll      |
| `server/jwt/manager.go`                  | JWT generate & verify                         |
| `server/lib/database/database.go`        | DB wrapper dengan retry                       |
| `server/lib/logger/logger.go`            | Structured logger (Zap)                       |
| `server/utils/utils.go`                  | Utility: ProcessId, GetEnv                    |

---

## Testing Checklist

- [ ] SignUp berhasil → 201
- [ ] SignUp duplicate email → 409
- [ ] SignUp email invalid → 422
- [ ] SignUp password < 6 char → 422
- [ ] SignUp full_name kosong → 422
- [ ] SignIn berhasil → 200 + token
- [ ] SignIn email salah → 401
- [ ] SignIn password salah → 401
- [ ] GetMe dengan token valid → 200
- [ ] GetMe tanpa token → 401
- [ ] GetMe dengan token expired → 401
- [ ] Password tersimpan sebagai bcrypt hash
- [ ] Profile terbuat otomatis saat signup

---

## Referensi Arsitektur

Service ini mengikuti pattern dari `addons-issuance-lc-service`:

| Komponen          | LC Service                                | Identity Service                         |
| ----------------- | ----------------------------------------- | ---------------------------------------- |
| Entry point + CLI | `server/main.go`                          | `server/main.go`                         |
| Config loading    | `server/core_config.go`                   | `server/core_config.go`                  |
| DB connection     | `server/core_db.go`                       | `server/core_db.go`                      |
| API struct        | `server/api/api.go`                       | `server/api/api.go`                      |
| API handlers      | `server/api/issued_lc_data_api.go`        | `server/api/identity_auth_api.go`        |
| Auth interceptor  | `server/api/issued_lc_authInterceptor.go` | `server/api/identity_authInterceptor.go` |
| Interceptor chain | `server/api/issued_lc_interceptor.go`     | `server/api/identity_interceptor.go`     |
| Error helpers     | `server/api/error.go`                     | `server/api/error.go`                    |
| DB provider       | `server/db/provider.go`                   | `server/db/provider.go`                  |
| DB queries        | `server/db/issued_lc_provider.go`         | `server/db/identity_provider.go`         |
| JWT manager       | `server/jwt/manager.go`                   | `server/jwt/manager.go`                  |
| DB wrapper        | `server/lib/database/`                    | `server/lib/database/`                   |
| Logger            | `server/lib/logger/`                      | `server/lib/logger/`                     |
| Utils             | `server/utils/utils.go`                   | `server/utils/utils.go`                  |
| Constants         | `server/constant/`                        | `server/constant/`                       |
