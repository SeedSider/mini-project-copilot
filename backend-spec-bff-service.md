# Backend Spec — BFF Service (Backend for Frontend)

> **Tujuan dokumen ini**: Mendefinisikan spesifikasi lengkap BFF service yang menjadi **single entry point** untuk mobile app BankEase. BFF menerima request REST dari client dan meneruskan ke downstream services (`identity-service`, `user-profile-service`) via gRPC.

---

## Daftar Isi

- [Architecture Overview](#architecture-overview)
- [Stack Teknologi](#stack-teknologi)
- [Struktur Folder](#struktur-folder)
- [Downstream Services](#downstream-services)
- [Proto Definitions](#proto-definitions)
- [REST API Endpoints](#rest-api-endpoints)
- [Orchestration Flows](#orchestration-flows)
- [Error Handling](#error-handling)
- [Interceptor & Middleware](#interceptor--middleware)
- [ServiceConnection Pattern](#serviceconnection-pattern)
- [JWT Management](#jwt-management)
- [Upload Image (Azure Blob)](#upload-image-azure-blob)
- [Environment Variables](#environment-variables)
- [Docker Compose](#docker-compose)
- [Testing Checklist](#testing-checklist)
- [Referensi Arsitektur](#referensi-arsitektur)

---

## Architecture Overview

```
┌──────────────┐         REST (HTTP)          ┌──────────────────────┐
│  Mobile App  │ ──────────────────────────▶  │   BFF Service        │
│  (React      │  POST /api/auth/signup       │   (grpc-gateway)     │
│   Native)    │  POST /api/auth/signin       │                      │
│              │  GET  /api/profile            │   Port 3000 (HTTP)   │
│              │  GET  /api/menu/...           │   Port 9090 (gRPC)   │
└──────────────┘  POST /api/upload/image      └──────┬───────┬───────┘
                                                     │       │
                                          gRPC       │       │  gRPC
                                                     ▼       ▼
                                    ┌────────────────┐ ┌──────────────────┐
                                    │ identity-      │ │ user-profile-    │
                                    │ service        │ │ service          │
                                    │                │ │                  │
                                    │ Port 9301      │ │ Port 9302        │
                                    │ (gRPC)         │ │ (gRPC)           │
                                    └───────┬────────┘ └────────┬─────────┘
                                            │                   │
                                            ▼                   ▼
                                    ┌────────────────┐ ┌──────────────────┐
                                    │ PostgreSQL     │ │ PostgreSQL       │
                                    │ identity_db    │ │ bankease_db      │
                                    └────────────────┘ └──────────────────┘
```

**Prinsip utama:**

1. **Client → BFF**: Semua request dari mobile app masuk via REST (HTTP) ke BFF
2. **BFF → Downstream**: BFF memanggil `identity-service` dan `user-profile-service` via gRPC
3. **Orchestration**: BFF mengorkestrasi multi-service call (contoh: SignUp = identity + profile)
4. **Single Entry Point**: Client hanya perlu tahu 1 endpoint (BFF), tidak perlu tahu downstream services
5. **JWT Local Verification**: BFF memverifikasi JWT token secara lokal (secret key sama dengan identity-service)
6. **Upload Direct**: BFF handle upload image langsung ke Azure Blob Storage (tidak lewat downstream service)

---

## Stack Teknologi

| Komponen     | Pilihan                                                  | Referensi di `addons-issuance-lc-service` |
| ------------ | -------------------------------------------------------- | ----------------------------------------- |
| Language     | Go 1.24                                                  | Go 1.24                                   |
| gRPC Server  | `google.golang.org/grpc`                                 | `google.golang.org/grpc`                  |
| HTTP Gateway | `grpc-gateway/v2`                                        | `grpc-gateway/v2`                         |
| Protobuf     | `protoc` + `protoc-gen-go` + `protoc-gen-grpc-gateway`   | Sama                                      |
| Router       | grpc-gateway `runtime.ServeMux` + custom `http.ServeMux` | Sama                                      |
| Config       | `joho/godotenv` + `os.LookupEnv`                         | `viper` + `godotenv`                      |
| JWT          | `dgrijalva/jwt-go` (HS256)                               | `dgrijalva/jwt-go`                        |
| Logger       | Zap structured logger                                    | Zap + FluentBit                           |
| CLI          | `urfave/cli`                                             | `urfave/cli`                              |
| Container    | Docker + Docker Compose                                  | Docker                                    |

---

## Struktur Folder

```
bff-service/
├── proto/
│   ├── bff_api.proto                       # BFF service definition (11 RPC methods)
│   └── bff_payload.proto                   # Request/response messages
├── protogen/
│   ├── bff-service/                        # Generated BFF Go code (server stubs)
│   │   ├── bff_api_grpc.pb.go
│   │   ├── bff_api.pb.go
│   │   └── bff_api.pb.gw.go               # grpc-gateway HTTP handlers
│   ├── identity-service/                   # Generated identity-service Go code (client)
│   │   ├── identity_api_grpc.pb.go
│   │   └── identity_api.pb.go
│   └── user-profile-service/               # Generated user-profile-service Go code (client)
│       ├── user_profile_api_grpc.pb.go
│       └── user_profile_api.pb.go
├── server/
│   ├── main.go                             # Entry point + CLI (grpc-server, gw-server, grpc-gw-server)
│   ├── core_config.go                      # Config loader (env vars)
│   ├── gateway_http_handler.go             # Custom HTTP handler (upload, CORS, error mapping)
│   ├── api/
│   │   ├── api.go                          # Server struct + constructor (DI)
│   │   ├── bff_auth_api.go                 # Handler: SignUp, SignIn, GetMe
│   │   ├── bff_profile_api.go              # Handler: Profile CRUD
│   │   ├── bff_menu_api.go                 # Handler: Menu queries
│   │   ├── bff_interceptor.go              # Interceptor chain: ProcessId → Logging → Auth
│   │   ├── bff_authInterceptor.go          # JWT auth interceptor
│   │   └── error.go                        # Error helpers + gRPC → HTTP mapping
│   ├── services/
│   │   └── service.go                      # ServiceConnection: gRPC clients ke downstream
│   ├── jwt/
│   │   └── manager.go                      # JWT Verify (HS256) — local verification
│   ├── lib/
│   │   └── logger/
│   │       └── logger.go                   # Zap structured logger
│   ├── utils/
│   │   └── utils.go                        # GetProcessIdFromCtx, GetEnv, GenerateProcessId
│   └── constant/
│       ├── constant.go                     # Response codes, date format
│       └── process_id.go                   # ProcessIdCtx key
├── www/
│   ├── swagger.json                        # Swagger API documentation
│   └── swagger-ui/
│       └── index.html                      # Swagger UI
├── .env.example                            # Template environment variables
├── docker-compose.yml                      # Full stack: BFF + identity + profile + 2x PostgreSQL
├── Dockerfile                              # Multi-stage build (golang:1.24 → alpine)
├── Makefile                                # build, run, proto-gen, docker-build
├── sonar-project.properties                # SonarQube config
├── go.mod / go.sum                         # Go module dependencies
└── generate.sh / generate.bat              # Proto code generation script
```

---

## Downstream Services

BFF berkomunikasi dengan 2 downstream service via gRPC:

### 1. identity-service (existing)

| RPC Method | Request                         | Response                       |
| ---------- | ------------------------------- | ------------------------------ |
| `SignUp`   | `username`, `password`, `phone` | `user_id`, `username`          |
| `SignIn`   | `username`, `password`          | `user_id`, `username`, `token` |
| `GetMe`    | _(empty, from JWT context)_     | `user_id`, `username`          |

- **Port gRPC**: 9301
- **Proto**: `proto/identity_api.proto`, `proto/identity_payload.proto`

### 2. user-profile-service (perlu ditambah gRPC support)

> **Catatan penting**: `user-profile-service` saat ini hanya REST. Perlu ditambahkan gRPC layer agar bisa dipanggil oleh BFF via gRPC. Proto contract didefinisikan di bawah.

| RPC Method              | Request                                                                                                    | Response              |
| ----------------------- | ---------------------------------------------------------------------------------------------------------- | --------------------- |
| `CreateProfile`         | `user_id`, `bank`, `branch`, `name`, `card_number`, `card_provider`, `balance`, `currency`, `account_type` | `Profile`             |
| `GetProfileByID`        | `id` (UUID)                                                                                                | `Profile`             |
| `GetProfileByUserID`    | `user_id` (UUID)                                                                                           | `Profile`             |
| `UpdateProfile`         | `id`, `bank`, `branch`, `name`, `card_number`                                                              | `StandardResponse`    |
| `GetAllMenus`           | _(empty)_                                                                                                  | `MenuResponse` (list) |
| `GetMenusByAccountType` | `account_type` (`REGULAR` / `PREMIUM`)                                                                     | `MenuResponse` (list) |

- **Port gRPC**: 9302 (baru)
- **Proto**: `proto/user_profile_api.proto`, `proto/user_profile_payload.proto` (baru)

---

## Proto Definitions

### BFF Proto — `bff_api.proto`

```proto
syntax = "proto3";

package bff;

option go_package = "protogen/bff-service";

import "google/api/annotations.proto";

service BffService {
    // Auth endpoints (orchestrated via identity-service)
    rpc SignUp(SignUpRequest) returns (SignUpResponse) {
        option (google.api.http) = {
            post: "/api/auth/signup"
            body: "*"
        };
    }

    rpc SignIn(SignInRequest) returns (SignInResponse) {
        option (google.api.http) = {
            post: "/api/auth/signin"
            body: "*"
        };
    }

    rpc GetMe(GetMeRequest) returns (GetMeResponse) {
        option (google.api.http) = {
            get: "/api/auth/me"
        };
    }

    // Profile endpoints (proxied to user-profile-service)
    rpc GetMyProfile(GetMyProfileRequest) returns (ProfileResponse) {
        option (google.api.http) = {
            get: "/api/profile"
        };
    }

    rpc GetProfileByID(GetProfileByIDRequest) returns (ProfileResponse) {
        option (google.api.http) = {
            get: "/api/profile/{id}"
        };
    }

    rpc GetProfileByUserID(GetProfileByUserIDRequest) returns (ProfileResponse) {
        option (google.api.http) = {
            get: "/api/profile/user/{user_id}"
        };
    }

    rpc CreateProfile(CreateProfileRequest) returns (ProfileResponse) {
        option (google.api.http) = {
            post: "/api/profile"
            body: "*"
        };
    }

    rpc UpdateProfile(UpdateProfileRequest) returns (StandardResponse) {
        option (google.api.http) = {
            put: "/api/profile/{id}"
            body: "*"
        };
    }

    // Menu endpoints (proxied to user-profile-service)
    rpc GetAllMenus(GetAllMenusRequest) returns (MenuListResponse) {
        option (google.api.http) = {
            get: "/api/menu"
        };
    }

    rpc GetMenusByAccountType(GetMenusByAccountTypeRequest) returns (MenuListResponse) {
        option (google.api.http) = {
            get: "/api/menu/{account_type}"
        };
    }

    // Upload endpoint — handled directly by BFF (NOT via grpc-gateway)
    // Registered as custom HTTP handler in gateway_http_handler.go
    // POST /api/upload/image (multipart/form-data)
}
```

### BFF Proto — `bff_payload.proto`

```proto
syntax = "proto3";

package bff;

option go_package = "protogen/bff-service";

// ── Auth Messages ──

message SignUpRequest {
    string username = 1;
    string password = 2;
    string phone    = 3;
}

message SignUpResponse {
    string user_id  = 1;
    string username = 2;
}

message SignInRequest {
    string username = 1;
    string password = 2;
}

message SignInResponse {
    string user_id  = 1;
    string username = 2;
    string token    = 3;
}

message GetMeRequest {}

message GetMeResponse {
    string user_id  = 1;
    string username = 2;
}

// ── Profile Messages ──

message GetMyProfileRequest {}

message GetProfileByIDRequest {
    string id = 1;
}

message GetProfileByUserIDRequest {
    string user_id = 1;
}

message CreateProfileRequest {
    string user_id       = 1;
    string bank          = 2;
    string branch        = 3;
    string name          = 4;
    string card_number   = 5;
    string card_provider = 6;
    int64  balance       = 7;
    string currency      = 8;
    string account_type  = 9;
}

message UpdateProfileRequest {
    string id          = 1;
    string bank        = 2;
    string branch      = 3;
    string name        = 4;
    string card_number = 5;
}

message ProfileResponse {
    string id            = 1;
    string user_id       = 2;
    string bank          = 3;
    string branch        = 4;
    string name          = 5;
    string card_number   = 6;
    string card_provider = 7;
    int64  balance       = 8;
    string currency      = 9;
    string account_type  = 10;
    string image         = 11;
}

// ── Menu Messages ──

message GetAllMenusRequest {}

message GetMenusByAccountTypeRequest {
    string account_type = 1;
}

message MenuItem {
    string id        = 1;
    int32  index     = 2;
    string type      = 3;
    string title     = 4;
    string icon_url  = 5;
    bool   is_active = 6;
}

message MenuListResponse {
    repeated MenuItem menus = 1;
}

// ── Common Messages ──

message StandardResponse {
    int32  code        = 1;
    string description = 2;
}
```

### User-Profile Service Proto — `user_profile_api.proto` (BARU)

> Proto ini perlu ditambahkan ke `user-profile-service` agar service bisa menerima gRPC calls dari BFF.

```proto
syntax = "proto3";

package userprofile;

option go_package = "protogen/user-profile-service";

service UserProfileService {
    rpc CreateProfile(CreateProfileRequest) returns (ProfileResponse);
    rpc GetProfileByID(GetProfileByIDRequest) returns (ProfileResponse);
    rpc GetProfileByUserID(GetProfileByUserIDRequest) returns (ProfileResponse);
    rpc UpdateProfile(UpdateProfileRequest) returns (StandardResponse);
    rpc GetAllMenus(GetAllMenusRequest) returns (MenuListResponse);
    rpc GetMenusByAccountType(GetMenusByAccountTypeRequest) returns (MenuListResponse);
}

// ── Profile Messages ──

message CreateProfileRequest {
    string user_id       = 1;
    string bank          = 2;
    string branch        = 3;
    string name          = 4;
    string card_number   = 5;
    string card_provider = 6;
    int64  balance       = 7;
    string currency      = 8;
    string account_type  = 9;
}

message GetProfileByIDRequest {
    string id = 1;
}

message GetProfileByUserIDRequest {
    string user_id = 1;
}

message UpdateProfileRequest {
    string id          = 1;
    string bank        = 2;
    string branch      = 3;
    string name        = 4;
    string card_number = 5;
}

message ProfileResponse {
    string id            = 1;
    string user_id       = 2;
    string bank          = 3;
    string branch        = 4;
    string name          = 5;
    string card_number   = 6;
    string card_provider = 7;
    int64  balance       = 8;
    string currency      = 9;
    string account_type  = 10;
    string image         = 11;
}

// ── Menu Messages ──

message GetAllMenusRequest {}

message GetMenusByAccountTypeRequest {
    string account_type = 1;
}

message MenuItem {
    string id        = 1;
    int32  index     = 2;
    string type      = 3;
    string title     = 4;
    string icon_url  = 5;
    bool   is_active = 6;
}

message MenuListResponse {
    repeated MenuItem menus = 1;
}

// ── Common ──

message StandardResponse {
    int32  code        = 1;
    string description = 2;
}
```

---

## REST API Endpoints

### Base URL

| Environment | URL                     |
| ----------- | ----------------------- |
| Development | `http://localhost:3000` |
| gRPC        | `localhost:9090`        |

### Shared Headers

| Header          | Wajib         | Keterangan                                    |
| --------------- | ------------- | --------------------------------------------- |
| `Content-Type`  | Ya (POST/PUT) | `application/json` atau `multipart/form-data` |
| `Authorization` | Endpoint auth | `Bearer <JWT token>`                          |

---

### 1. `POST /api/auth/signup`

Registrasi user baru. **Orchestrated**: BFF memanggil `identity-service.SignUp` → `user-profile-service.CreateProfile`.

**Request Body:**

```json
{
  "username": "johndoe",
  "password": "password123",
  "phone": "08123456789"
}
```

| Field      | Type   | Wajib | Validasi           |
| ---------- | ------ | ----- | ------------------ |
| `username` | string | Ya    | Tidak boleh kosong |
| `password` | string | Ya    | Minimal 6 karakter |
| `phone`    | string | Tidak | Opsional           |

**Response 201:**

```json
{
  "user_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "username": "johndoe"
}
```

**Error Responses:**

| Status | Kondisi                  | Response Body                                                                       |
| ------ | ------------------------ | ----------------------------------------------------------------------------------- |
| 409    | Username sudah terdaftar | `{"error": true, "code": 409, "message": "Username already registered"}`            |
| 422    | Validation error         | `{"error": true, "code": 422, "message": "password must be at least 6 characters"}` |
| 500    | Internal server error    | `{"error": true, "code": 500, "message": "Internal error"}`                         |

---

### 2. `POST /api/auth/signin`

Login user dan mendapatkan JWT token. **Proxy**: BFF meneruskan ke `identity-service.SignIn`.

**Request Body:**

```json
{
  "username": "johndoe",
  "password": "password123"
}
```

| Field      | Type   | Wajib | Validasi           |
| ---------- | ------ | ----- | ------------------ |
| `username` | string | Ya    | Tidak boleh kosong |
| `password` | string | Ya    | Tidak boleh kosong |

**Response 200:**

```json
{
  "user_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "username": "johndoe",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Error Responses:**

| Status | Kondisi              | Response Body                                                     |
| ------ | -------------------- | ----------------------------------------------------------------- |
| 401    | Invalid credentials  | `{"error": true, "code": 401, "message": "Unauthorized"}`         |
| 400    | Request body invalid | `{"error": true, "code": 400, "message": "Invalid request body"}` |

---

### 3. `GET /api/auth/me`

Mengambil informasi user berdasarkan JWT token. **Proxy**: BFF meneruskan ke `identity-service.GetMe`.

**Header:**

```
Authorization: Bearer <token>
```

**Response 200:**

```json
{
  "user_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "username": "johndoe"
}
```

**Error Responses:**

| Status | Kondisi           | Response Body                                                                |
| ------ | ----------------- | ---------------------------------------------------------------------------- |
| 401    | Token tidak valid | `{"error": true, "code": 401, "message": "Invalid token"}`                   |
| 401    | Token tidak ada   | `{"error": true, "code": 401, "message": "Authorization token is required"}` |

---

### 4. `GET /api/profile`

Mengambil profil pengguna berdasarkan JWT token. **Orchestrated**: BFF verify JWT → extract `user_id` → `user-profile-service.GetProfileByUserID`.

**Header:**

```
Authorization: Bearer <token>
```

**Response 200:**

```json
{
  "id": "da08ecfe-de3b-42b1-b1ce-018e144198f5",
  "user_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "bank": "BRI",
  "branch": "Jakarta",
  "name": "johndoe",
  "card_number": "12355478990",
  "card_provider": "Mastercard Platinum",
  "balance": 5000000,
  "currency": "IDR",
  "accountType": "REGULAR",
  "image": ""
}
```

**Error Responses:**

| Status | Kondisi           |
| ------ | ----------------- |
| 401    | Token tidak valid |
| 404    | Profile not found |
| 500    | Internal error    |

---

### 5. `GET /api/profile/{id}`

Mengambil profil berdasarkan profile ID (UUID). **Proxy**: BFF meneruskan ke `user-profile-service.GetProfileByID`.

**Request:**

```
GET /api/profile/da08ecfe-de3b-42b1-b1ce-018e144198f5
```

**Response 200:**

```json
{
  "id": "da08ecfe-de3b-42b1-b1ce-018e144198f5",
  "user_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "bank": "BRI",
  "branch": "Jakarta",
  "name": "johndoe",
  "card_number": "12355478990",
  "card_provider": "Mastercard Platinum",
  "balance": 5000000,
  "currency": "IDR",
  "accountType": "REGULAR",
  "image": ""
}
```

**Error Responses:**

| Status | Kondisi           |
| ------ | ----------------- |
| 400    | ID kosong         |
| 404    | Profile not found |
| 500    | Internal error    |

---

### 6. `GET /api/profile/user/{user_id}`

Mengambil profil berdasarkan user ID (dari identity-service). **Proxy**: BFF meneruskan ke `user-profile-service.GetProfileByUserID`.

**Request:**

```
GET /api/profile/user/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**Response 200:** Sama seperti endpoint #5.

**Error Responses:**

| Status | Kondisi           |
| ------ | ----------------- |
| 400    | user_id kosong    |
| 404    | Profile not found |
| 500    | Internal error    |

---

### 7. `POST /api/profile`

Membuat profil baru. **Proxy**: BFF meneruskan ke `user-profile-service.CreateProfile`.

**Request Body:**

```json
{
  "user_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "bank": "BRI",
  "branch": "Jakarta",
  "name": "johndoe",
  "card_number": "",
  "card_provider": "",
  "balance": 0,
  "currency": "IDR",
  "accountType": "REGULAR"
}
```

| Field           | Type   | Wajib | Default     |
| --------------- | ------ | ----- | ----------- |
| `user_id`       | string | Ya    | —           |
| `bank`          | string | Tidak | —           |
| `branch`        | string | Tidak | —           |
| `name`          | string | Tidak | —           |
| `card_number`   | string | Tidak | `""`        |
| `card_provider` | string | Tidak | `""`        |
| `balance`       | int64  | Tidak | `0`         |
| `currency`      | string | Tidak | `"IDR"`     |
| `accountType`   | string | Tidak | `"REGULAR"` |

**Response 201:** Profile object (sama seperti response endpoint #5).

**Error Responses:**

| Status | Kondisi               |
| ------ | --------------------- |
| 400    | user_id kosong        |
| 500    | Failed create profile |

---

### 8. `PUT /api/profile/{id}`

Mengubah data profil. **Proxy**: BFF meneruskan ke `user-profile-service.UpdateProfile`.

**Request:**

```
PUT /api/profile/da08ecfe-de3b-42b1-b1ce-018e144198f5
```

**Request Body:**

```json
{
  "bank": "Citibank",
  "branch": "Tangerang",
  "name": "Jane Doe",
  "card_number": "12355478990"
}
```

| Field         | Type   | Keterangan       |
| ------------- | ------ | ---------------- |
| `bank`        | string | Nama bank        |
| `branch`      | string | Nama cabang bank |
| `name`        | string | Nama pengguna    |
| `card_number` | string | Nomor kartu      |

**Response 200:**

```json
{
  "code": 200,
  "description": "Data pengguna berhasil diubah."
}
```

**Error Responses:**

| Status | Kondisi              |
| ------ | -------------------- |
| 400    | Invalid request body |
| 404    | Profile not found    |
| 500    | Internal error       |

---

### 9. `GET /api/menu`

Mengambil semua menu homepage yang aktif. **Proxy**: BFF meneruskan ke `user-profile-service.GetAllMenus`.

**Response 200:**

```json
{
  "menus": [
    {
      "id": "menu_001",
      "index": 1,
      "type": "REGULAR",
      "title": "Account and Card",
      "icon_url": "https://fonts.googleapis.com/...",
      "is_active": true
    },
    {
      "id": "menu_002",
      "index": 2,
      "type": "PREMIUM",
      "title": "Transfer",
      "icon_url": "https://fonts.googleapis.com/...",
      "is_active": true
    }
  ]
}
```

**Error Responses:**

| Status | Kondisi        |
| ------ | -------------- |
| 500    | Internal error |

---

### 10. `GET /api/menu/{accountType}`

Mengambil menu berdasarkan account type. **Proxy**: BFF meneruskan ke `user-profile-service.GetMenusByAccountType`.

- `PREMIUM` → return **semua** menu (REGULAR + PREMIUM)
- `REGULAR` → return **hanya** menu tipe REGULAR

**Request:**

```
GET /api/menu/REGULAR
```

**Response 200:**

```json
{
  "menus": [
    {
      "id": "menu_001",
      "index": 1,
      "type": "REGULAR",
      "title": "Account and Card",
      "icon_url": "https://fonts.googleapis.com/...",
      "is_active": true
    }
  ]
}
```

**Error Responses:**

| Status | Kondisi            |
| ------ | ------------------ |
| 400    | accountType kosong |
| 500    | Internal error     |

---

### 11. `POST /api/upload/image`

Upload image ke Azure Blob Storage. **Handled langsung oleh BFF** — tidak diteruskan ke downstream service.

> Endpoint ini didaftarkan sebagai **custom HTTP handler** (bukan melalui grpc-gateway), mirip pattern custom endpoint di `addons-issuance-lc-service/server/gateway_http_handler.go`.

**Request:**

```
POST /api/upload/image
Content-Type: multipart/form-data
```

| Field   | Type | Wajib | Keterangan                                       |
| ------- | ---- | ----- | ------------------------------------------------ |
| `image` | file | Ya    | File image (jpeg, png, gif, webp, svg — max 5MB) |

**Response 200:**

```json
{
  "code": 200,
  "description": "Image uploaded successfully",
  "url": "https://account.blob.core.windows.net/images/a1b2c3d4.jpg"
}
```

**Error Responses:**

| Status | Kondisi                           |
| ------ | --------------------------------- |
| 400    | Field 'image' kosong              |
| 400    | MIME type tidak didukung          |
| 413    | File lebih dari 5MB               |
| 500    | Gagal upload ke Azure             |
| 503    | Azure SAS URL belum dikonfigurasi |

**Business Logic Upload:**

1. Limit request body ke 5MB (`http.MaxBytesReader`)
2. Parse multipart form dan ambil file `image`
3. Baca content ke memory
4. Deteksi MIME type via `http.DetectContentType`
5. Validasi MIME: `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `image/svg+xml`
6. Generate random hex filename (16 bytes = 32 chars)
7. HTTP PUT ke Azure Blob Storage via SAS URL
8. Return URL publik

**Allowed MIME Types:**

| MIME Type       | Extension |
| --------------- | --------- |
| `image/jpeg`    | `.jpg`    |
| `image/png`     | `.png`    |
| `image/gif`     | `.gif`    |
| `image/webp`    | `.webp`   |
| `image/svg+xml` | `.svg`    |

---

## Orchestration Flows

### Flow 1: Sign Up (Orchestrated — Multi-Service)

```
Client                    BFF                    identity-service        user-profile-service
  │                        │                          │                          │
  │ POST /api/auth/signup  │                          │                          │
  │ {username, password,   │                          │                          │
  │  phone}                │                          │                          │
  │───────────────────────▶│                          │                          │
  │                        │                          │                          │
  │                        │ 1. Validate input        │                          │
  │                        │    (username wajib,      │                          │
  │                        │     password ≥ 6)        │                          │
  │                        │                          │                          │
  │                        │ 2. gRPC SignUp            │                          │
  │                        │─────────────────────────▶│                          │
  │                        │                          │ • Check username exists  │
  │                        │                          │ • bcrypt hash password   │
  │                        │                          │ • INSERT INTO users      │
  │                        │◀─────────────────────────│                          │
  │                        │   {user_id, username}    │                          │
  │                        │                          │                          │
  │                        │ 3. gRPC CreateProfile    │                          │
  │                        │──────────────────────────────────────────────────▶  │
  │                        │                          │  • INSERT INTO profile   │
  │                        │                          │    (user_id, default     │
  │                        │                          │     bank/branch/balance) │
  │                        │◀──────────────────────────────────────────────────  │
  │                        │   {profile object}       │                          │
  │                        │                          │                          │
  │ 201 {user_id, username}│                          │                          │
  │◀───────────────────────│                          │                          │
```

**Error Handling SignUp:**

- Jika `identity-service.SignUp` gagal → return error langsung (409 / 422 / 500)
- Jika `user-profile-service.CreateProfile` gagal → **log error, tetap return SignUp response** (best-effort, profile bisa dibuat ulang nanti)
- Alasan: mengikuti pattern existing di `identity-service/server/api/identity_auth_api.go` (`createBankingProfile` bersifat best-effort)

**Default Profile Data saat SignUp:**

```json
{
  "user_id": "<dari identity-service>",
  "bank": "BRI",
  "branch": "Jakarta",
  "name": "<username>",
  "card_number": "",
  "card_provider": "",
  "balance": 0,
  "currency": "IDR",
  "accountType": "REGULAR"
}
```

---

### Flow 2: Sign In (Proxy)

```
Client                    BFF                    identity-service
  │                        │                          │
  │ POST /api/auth/signin  │                          │
  │ {username, password}   │                          │
  │───────────────────────▶│                          │
  │                        │ gRPC SignIn               │
  │                        │─────────────────────────▶│
  │                        │                          │ • Find user by username
  │                        │                          │ • bcrypt compare
  │                        │                          │ • Generate JWT
  │                        │◀─────────────────────────│
  │                        │ {user_id, username, token}│
  │ 200 {user_id, username,│                          │
  │      token}            │                          │
  │◀───────────────────────│                          │
```

---

### Flow 3: Get Me (JWT Verify + Proxy)

```
Client                    BFF                    identity-service
  │                        │                          │
  │ GET /api/auth/me       │                          │
  │ Authorization: Bearer  │                          │
  │───────────────────────▶│                          │
  │                        │ 1. Verify JWT (lokal)    │
  │                        │    extract user_claims   │
  │                        │                          │
  │                        │ 2. gRPC GetMe            │
  │                        │    (forward claims ctx)  │
  │                        │─────────────────────────▶│
  │                        │◀─────────────────────────│
  │                        │ {user_id, username}      │
  │ 200 {user_id, username}│                          │
  │◀───────────────────────│                          │
```

---

### Flow 4: Get My Profile (JWT Verify + Profile Proxy)

```
Client                    BFF                          user-profile-service
  │                        │                                  │
  │ GET /api/profile       │                                  │
  │ Authorization: Bearer  │                                  │
  │───────────────────────▶│                                  │
  │                        │ 1. Verify JWT (lokal)            │
  │                        │    extract user_id               │
  │                        │                                  │
  │                        │ 2. gRPC GetProfileByUserID       │
  │                        │    (user_id dari JWT)            │
  │                        │─────────────────────────────────▶│
  │                        │◀─────────────────────────────────│
  │                        │ {profile object}                 │
  │ 200 {profile}          │                                  │
  │◀───────────────────────│                                  │
```

---

### Flow 5: Profile CRUD (Simple Proxy)

Endpoint `GET /api/profile/{id}`, `GET /api/profile/user/{user_id}`, `POST /api/profile`, `PUT /api/profile/{id}` — semuanya **simple proxy** tanpa orchestration:

```
Client                    BFF                          user-profile-service
  │                        │                                  │
  │ REST request           │                                  │
  │───────────────────────▶│                                  │
  │                        │ Convert REST → gRPC              │
  │                        │ (grpc-gateway otomatis)          │
  │                        │─────────────────────────────────▶│
  │                        │◀─────────────────────────────────│
  │                        │ Convert gRPC → REST              │
  │ REST response          │                                  │
  │◀───────────────────────│                                  │
```

---

### Flow 6: Menu Queries (Simple Proxy)

Endpoint `GET /api/menu`, `GET /api/menu/{accountType}` — **simple proxy** ke `user-profile-service`:

```
Client → BFF → gRPC GetAllMenus / GetMenusByAccountType → user-profile-service → response
```

**Business Logic Menu Filter** (di user-profile-service):

- `PREMIUM` → return semua menu aktif
- `REGULAR` → return hanya menu dengan `type = 'REGULAR'`

---

### Flow 7: Upload Image (BFF Direct)

```
Client                    BFF                          Azure Blob Storage
  │                        │                                  │
  │ POST /api/upload/image │                                  │
  │ (multipart/form-data)  │                                  │
  │───────────────────────▶│                                  │
  │                        │ 1. Validate file size (≤ 5MB)   │
  │                        │ 2. Detect MIME type              │
  │                        │ 3. Validate MIME                 │
  │                        │ 4. Generate random filename      │
  │                        │ 5. HTTP PUT to Azure Blob        │
  │                        │─────────────────────────────────▶│
  │                        │◀─────────────────────────────────│
  │                        │ 6. Return URL                    │
  │ 200 {code, desc, url}  │                                  │
  │◀───────────────────────│                                  │
```

---

## Error Handling

### gRPC → HTTP Status Code Mapping

BFF menggunakan pattern yang sama dengan `addons-issuance-lc-service` untuk mapping gRPC status code ke HTTP:

| gRPC Code          | HTTP Status | Keterangan               |
| ------------------ | ----------- | ------------------------ |
| `OK`               | 200         | Sukses                   |
| `InvalidArgument`  | 422         | Validation error         |
| `Unauthenticated`  | 401         | JWT invalid / missing    |
| `AlreadyExists`    | 409         | Duplicate (username)     |
| `NotFound`         | 404         | Resource tidak ditemukan |
| `PermissionDenied` | 403         | Tidak punya akses        |
| `Internal`         | 500         | Server error             |
| `Unavailable`      | 503         | Service downstream down  |

### Standard Error Response Format

```json
{
  "error": true,
  "code": 401,
  "message": "Unauthorized"
}
```

### Custom HTTP Error Handler

Implementasi di `gateway_http_handler.go` — override default grpc-gateway error handler:

```go
func CustomHTTPError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler,
    w http.ResponseWriter, r *http.Request, err error) {

    st, _ := status.FromError(err)
    httpCode := grpcToHTTPCode(st.Code())

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(httpCode)
    json.NewEncoder(w).Encode(&ErrorBodyResponse{
        Error:   true,
        Code:    int32(httpCode),
        Message: st.Message(),
    })
}
```

---

## Interceptor & Middleware

### gRPC Interceptor Chain

Urutan interceptor (mengikuti pattern `addons-issuance-lc-service`):

```
ProcessIdInterceptor → LoggingInterceptor → AuthInterceptor
```

#### 1. ProcessIdInterceptor

- Generate UUID unik untuk setiap request sebagai trace ID
- Simpan di context (`process_id`)
- Propagate ke downstream service via gRPC metadata

#### 2. LoggingInterceptor

- Log method name, process_id, dan duration setiap RPC call
- Format: `[process_id] [method] duration=Xms`

#### 3. AuthInterceptor

- Hanya diterapkan pada method yang butuh autentikasi:
  - `GetMe`
  - `GetMyProfile`
- Extract `Authorization` header → verify JWT → inject `user_claims` ke context
- Method publik yang **tidak** butuh auth: `SignUp`, `SignIn`, semua profile CRUD by ID, menu queries

### HTTP Middleware

#### CORS Middleware

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, Idempotency-Key
```

#### Security Headers

```
Strict-Transport-Security: max-age=31536000
Content-Security-Policy: object-src 'none'; child-src 'none'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: no-referrer
```

---

## ServiceConnection Pattern

Mengikuti pattern `addons-issuance-lc-service/server/services/service.go`:

```go
// ServiceConnection holds gRPC client connections to downstream services.
type ServiceConnection struct {
    IdentityService    *grpc.ClientConn
    UserProfileService *grpc.ClientConn
}

// InitServicesConn initializes all gRPC client connections.
func InitServicesConn(identityAddr, profileAddr string) *ServiceConnection {
    services := &ServiceConnection{}

    var err error
    opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

    services.IdentityService, err = grpc.Dial(identityAddr, opts...)
    if err != nil {
        log.Fatalf("Failed to connect to identity-service: %v", err)
    }

    services.UserProfileService, err = grpc.Dial(profileAddr, opts...)
    if err != nil {
        log.Fatalf("Failed to connect to user-profile-service: %v", err)
    }

    return services
}

// Client accessor methods
func (sc *ServiceConnection) IdentityClient() identityPB.IdentityServiceClient {
    return identityPB.NewIdentityServiceClient(sc.IdentityService)
}

func (sc *ServiceConnection) UserProfileClient() profilePB.UserProfileServiceClient {
    return profilePB.NewUserProfileServiceClient(sc.UserProfileService)
}
```

### Server Struct (Dependency Injection)

```go
type Server struct {
    manager *manager.JWTManager           // JWT verification (local)
    scvConn *svc.ServiceConnection        // gRPC clients ke downstream
    logger  *logger.Logger                // Structured logger

    pb.BffServiceServer                   // Generated gRPC server interface
}

func New(jwtSecret, jwtDuration string, svcConn *svc.ServiceConnection, logger *logger.Logger) *Server {
    return &Server{
        manager: manager.NewJWTManager(jwtSecret, jwtDuration),
        scvConn: svcConn,
        logger:  logger,
    }
}
```

> **Perbedaan dengan `addons-issuance-lc-service`**: BFF **tidak punya database sendiri** dan **tidak punya provider**. Semua data diakses via gRPC ke downstream services.

---

## JWT Management

BFF memverifikasi JWT secara **lokal** (tidak perlu call ke identity-service untuk verifikasi). Secret key harus sama dengan yang digunakan identity-service.

### JWTManager

```go
type JWTManager struct {
    secretKey     string
    tokenDuration time.Duration
}

type UserClaims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    jwt.StandardClaims
}

// Verify validates the token and returns claims.
// BFF hanya perlu Verify, tidak perlu Generate (identity-service yang generate).
func (m *JWTManager) Verify(accessToken string) (*UserClaims, error)
```

### Protected Endpoints

| Endpoint                        | Auth Required |
| ------------------------------- | ------------- |
| POST /api/auth/signup           | ❌            |
| POST /api/auth/signin           | ❌            |
| GET /api/auth/me                | ✅            |
| GET /api/profile                | ✅            |
| GET /api/profile/{id}           | ❌            |
| GET /api/profile/user/{user_id} | ❌            |
| POST /api/profile               | ❌            |
| PUT /api/profile/{id}           | ❌            |
| GET /api/menu                   | ❌            |
| GET /api/menu/{accountType}     | ❌            |
| POST /api/upload/image          | ❌            |

---

## Upload Image (Azure Blob)

BFF menangani upload image **secara langsung** ke Azure Blob Storage tanpa meneruskan ke downstream service.

### Konfigurasi

| Variable          | Contoh                                                  |
| ----------------- | ------------------------------------------------------- |
| `AZURE_SAS_URL`   | `https://account.blob.core.windows.net/?sv=...&sig=...` |
| `AZURE_CONTAINER` | `images`                                                |

### Implementasi

Logika upload di `server/gateway_http_handler.go` sebagai custom HTTP handler (tidak melalui grpc-gateway karena endpoint ini menerima `multipart/form-data`, bukan JSON):

```go
// Registrasi di HTTP mux (bukan grpc-gateway mux)
httpMux.HandleFunc("/api/upload/image", s.HandleUploadImage)
```

**Flow:**

1. `http.MaxBytesReader(w, r.Body, 5<<20)` — limit 5MB
2. `r.ParseMultipartForm(5 << 20)` — parse form
3. `r.FormFile("image")` — ambil file
4. `io.ReadAll(file)` — baca content
5. `http.DetectContentType(data)` — deteksi MIME
6. Validate: hanya `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `image/svg+xml`
7. Generate filename: `crypto/rand` → hex (32 chars) + extension
8. Build Azure URL: `{SAS_URL_BASE}/{container}/{filename}?{SAS_PARAMS}`
9. `http.NewRequest("PUT", azureURL, bytes.NewReader(data))` → set `x-ms-blob-type: BlockBlob`
10. Execute HTTP PUT → check status
11. Return `{code: 200, description: "Image uploaded successfully", url: publicURL}`

---

## Environment Variables

| Variable                | Default          | Deskripsi                                |
| ----------------------- | ---------------- | ---------------------------------------- |
| `BFF_GRPC_PORT`         | `9090`           | Port gRPC server BFF                     |
| `BFF_HTTP_PORT`         | `3000`           | Port HTTP gateway BFF                    |
| `IDENTITY_SERVICE_ADDR` | `localhost:9301` | Alamat gRPC identity-service             |
| `PROFILE_SERVICE_ADDR`  | `localhost:9302` | Alamat gRPC user-profile-service         |
| `JWT_SECRET`            | `secret`         | Secret key JWT (harus sama dgn identity) |
| `JWT_DURATION`          | `24h`            | Durasi token JWT                         |
| `AZURE_SAS_URL`         | —                | Azure Blob Storage SAS URL               |
| `AZURE_CONTAINER`       | `images`         | Nama container Azure Blob                |
| `ENV`                   | `DEV`            | Environment (DEV/PROD)                   |
| `APP_NAME`              | `bff-service`    | Nama aplikasi                            |
| `LOGGER_OUTPUT`         | `stdout`         | Output logger (stdout/elastic)           |
| `LOGGER_LEVEL`          | `debug`          | Level log                                |

---

## Docker Compose

Full stack development environment — 5 containers:

```yaml
version: "3.8"

services:
  # ── Databases ──

  identity-db:
    image: postgres:17-alpine
    environment:
      POSTGRES_DB: identity_db
      POSTGRES_USER: identity
      POSTGRES_PASSWORD: identity123
    ports:
      - "5432:5432"
    volumes:
      - identity_pgdata:/var/lib/postgresql/data
      - ./identity-service/migrations/001_init.sql:/docker-entrypoint-initdb.d/01-schema.sql
      - ./identity-service/migrations/002_rename_email_to_username.sql:/docker-entrypoint-initdb.d/02-rename.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U identity -d identity_db"]
      interval: 5s
      timeout: 3s
      retries: 5

  profile-db:
    image: postgres:17-alpine
    environment:
      POSTGRES_DB: bankease_db
      POSTGRES_USER: bankease
      POSTGRES_PASSWORD: bankease123
    ports:
      - "5433:5432"
    volumes:
      - profile_pgdata:/var/lib/postgresql/data
      - ./user-profile-service/internal/db/migrations/001_init.sql:/docker-entrypoint-initdb.d/01-schema.sql
      - ./user-profile-service/internal/db/migrations/002_add_image_to_profile.sql:/docker-entrypoint-initdb.d/02-image.sql
      - ./user-profile-service/internal/db/migrations/003_add_user_id_to_profile.sql:/docker-entrypoint-initdb.d/03-userid.sql
      - ./user-profile-service/seed.sql:/docker-entrypoint-initdb.d/04-seed.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U bankease -d bankease_db"]
      interval: 5s
      timeout: 3s
      retries: 5

  # ── Services ──

  identity-service:
    build:
      context: ./identity-service
      dockerfile: Dockerfile
    environment:
      DB_HOST: identity-db
      DB_PORT: 5432
      DB_USER: identity
      DB_PASSWORD: identity123
      DB_NAME: identity_db
      DB_SSLMODE: disable
      JWT_SECRET: bankease-secret-key
      JWT_DURATION: 24h
      PROFILE_SERVICE_URL: "" # Kosong — BFF yang handle orchestration
    ports:
      - "9301:9301"
      - "3031:3031"
    depends_on:
      identity-db:
        condition: service_healthy

  user-profile-service:
    build:
      context: ./user-profile-service
      dockerfile: Dockerfile
    environment:
      DATABASE_URL: postgres://bankease:bankease123@profile-db:5432/bankease_db?sslmode=disable
      PORT: 8080
      GRPC_PORT: 9302
      JWT_SECRET: bankease-secret-key
    ports:
      - "8080:8080"
      - "9302:9302"
    depends_on:
      profile-db:
        condition: service_healthy

  bff-service:
    build:
      context: ./bff-service
      dockerfile: Dockerfile
    environment:
      BFF_GRPC_PORT: 9090
      BFF_HTTP_PORT: 3000
      IDENTITY_SERVICE_ADDR: identity-service:9301
      PROFILE_SERVICE_ADDR: user-profile-service:9302
      JWT_SECRET: bankease-secret-key
      JWT_DURATION: 24h
      AZURE_SAS_URL: ""
      AZURE_CONTAINER: images
    ports:
      - "3000:3000"
      - "9090:9090"
    depends_on:
      - identity-service
      - user-profile-service

volumes:
  identity_pgdata:
  profile_pgdata:
```

### Cara Menjalankan

```bash
# Start semua services
docker compose up --build

# Stop
docker compose down

# Stop + reset semua data
docker compose down -v
```

### Port Mapping

| Service              | HTTP | gRPC |
| -------------------- | ---- | ---- |
| BFF Service          | 3000 | 9090 |
| identity-service     | 3031 | 9301 |
| user-profile-service | 8080 | 9302 |
| identity-db          | 5432 | —    |
| profile-db           | 5433 | —    |

> **Catatan**: Client (mobile app) hanya perlu tahu `http://localhost:3000` (BFF). Tidak perlu akses langsung ke identity-service atau user-profile-service.

---

## Testing Checklist

### Auth Endpoints

- [ ] POST /api/auth/signup berhasil → 201 + user_id
- [ ] POST /api/auth/signup → profile otomatis terbuat di user-profile-service
- [ ] POST /api/auth/signup duplicate username → 409
- [ ] POST /api/auth/signup password < 6 → 422
- [ ] POST /api/auth/signup username kosong → 422
- [ ] POST /api/auth/signin berhasil → 200 + token
- [ ] POST /api/auth/signin password salah → 401
- [ ] POST /api/auth/signin username tidak ada → 401
- [ ] GET /api/auth/me dengan token valid → 200
- [ ] GET /api/auth/me tanpa header → 401
- [ ] GET /api/auth/me token expired → 401

### Profile Endpoints

- [ ] GET /api/profile dengan token valid → 200 + profile
- [ ] GET /api/profile tanpa token → 401
- [ ] GET /api/profile token valid tapi profile belum ada → 404
- [ ] GET /api/profile/{id} → 200 + profile
- [ ] GET /api/profile/{id} tidak ditemukan → 404
- [ ] GET /api/profile/user/{user_id} → 200 + profile
- [ ] POST /api/profile → 201 + profile baru
- [ ] POST /api/profile tanpa user_id → 400
- [ ] PUT /api/profile/{id} → 200 + success message
- [ ] PUT /api/profile/{id} tidak ditemukan → 404

### Menu Endpoints

- [ ] GET /api/menu → 200 + semua menu aktif
- [ ] GET /api/menu/REGULAR → 200 + hanya menu REGULAR
- [ ] GET /api/menu/PREMIUM → 200 + semua menu (REGULAR + PREMIUM)

### Upload Endpoint

- [ ] POST /api/upload/image jpeg → 200 + URL
- [ ] POST /api/upload/image png → 200 + URL
- [ ] POST /api/upload/image file > 5MB → 413
- [ ] POST /api/upload/image MIME tidak didukung → 400
- [ ] POST /api/upload/image tanpa file → 400
- [ ] POST /api/upload/image Azure belum dikonfigurasi → 503

### End-to-End Flow

- [ ] SignUp → SignIn → GetMe → GetMyProfile → UpdateProfile → GetMyProfile (data terupdate)
- [ ] SignUp → UploadImage → UpdateProfile (simpan URL image)
- [ ] GetAllMenus → GetMenusByAccountType(REGULAR) → GetMenusByAccountType(PREMIUM)

### Error Handling

- [ ] identity-service down → BFF return 503
- [ ] user-profile-service down → BFF return 503
- [ ] gRPC error mapping benar (InvalidArgument → 422, Unauthenticated → 401, dll)

---

## Referensi Arsitektur

BFF service mengikuti pattern dari `addons-issuance-lc-service`:

| Komponen            | addons-issuance-lc-service                | BFF Service                         |
| ------------------- | ----------------------------------------- | ----------------------------------- |
| Entry point + CLI   | `server/main.go`                          | `server/main.go`                    |
| Config loading      | `server/core_config.go`                   | `server/core_config.go`             |
| HTTP Gateway        | `server/gateway_http_handler.go`          | `server/gateway_http_handler.go`    |
| API struct          | `server/api/api.go`                       | `server/api/api.go`                 |
| API handlers        | `server/api/issued_lc_data_api.go`        | `server/api/bff_auth_api.go` dll    |
| Auth interceptor    | `server/api/issued_lc_authInterceptor.go` | `server/api/bff_authInterceptor.go` |
| Interceptor chain   | `server/api/issued_lc_interceptor.go`     | `server/api/bff_interceptor.go`     |
| Error helpers       | `server/api/error.go`                     | `server/api/error.go`               |
| Service connections | `server/services/service.go`              | `server/services/service.go`        |
| JWT manager         | `server/jwt/manager.go`                   | `server/jwt/manager.go`             |
| Logger              | `server/lib/logger/`                      | `server/lib/logger/`                |
| Utils               | `server/utils/utils.go`                   | `server/utils/utils.go`             |
| Constants           | `server/constant/`                        | `server/constant/`                  |
| Proto definitions   | `proto/`                                  | `proto/`                            |
| Generated code      | `protogen/`                               | `protogen/`                         |

### Perbedaan Utama

| Aspek               | addons-issuance-lc-service | BFF Service                     |
| ------------------- | -------------------------- | ------------------------------- |
| Database sendiri    | Ya (GORM + PostgreSQL)     | **Tidak** — stateless           |
| Provider/DB layer   | Ya (`server/db/`)          | **Tidak ada**                   |
| MinIO integration   | Ya (file storage)          | **Tidak** — Azure Blob direct   |
| Signature           | Ya (response signing)      | **Tidak**                       |
| Downstream services | 15+ services               | 2 services (identity + profile) |
| Own gRPC service    | Ya (IssuedLcService)       | Ya (BffService)                 |

---

## Prerequisites — Perubahan yang Dibutuhkan

Sebelum BFF bisa diimplementasi, perubahan berikut perlu dilakukan:

### 1. user-profile-service: Tambah gRPC Support

Saat ini `user-profile-service` hanya REST. Perlu ditambahkan:

- [ ] Proto files: `proto/user_profile_api.proto` + `proto/user_profile_payload.proto`
- [ ] gRPC server di `cmd/server/main.go` (listen di port `GRPC_PORT`)
- [ ] gRPC handler layer: translate gRPC request → repository call → gRPC response
- [ ] Proto code generation script (`generate.sh` / `generate.bat`)
- [ ] Update `Dockerfile` untuk expose port gRPC

### 2. identity-service: Pastikan gRPC Server Running

Identity-service sudah memiliki proto definitions tapi saat ini hanya menjalankan HTTP server manual. Perlu dipastikan:

- [ ] gRPC server listen di port 9301
- [ ] Proto generated code tersedia untuk BFF sebagai client
- [ ] `SignUp` tidak lagi memanggil `createBankingProfile` secara internal (BFF yang handle orchestration)

### 3. Shared Proto Repository

- [ ] Proto files dari `identity-service` dan `user-profile-service` perlu di-copy atau di-reference oleh BFF untuk generate client code
- [ ] Pastikan `go_package` path konsisten di semua proto files
