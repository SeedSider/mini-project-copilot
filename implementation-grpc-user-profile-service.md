# Implementation Plan: gRPC Layer untuk user-profile-service

## Ringkasan

Menambahkan **Proto files + hand-written protogen Go code + gRPC server (port 9302)** ke `user-profile-service`, mengikuti pattern `identity-service` yang menggunakan hand-written protogen (tanpa protoc toolchain). gRPC berjalan bersamaan dengan REST yang sudah ada (port 8080).

### 6 RPC Methods

| # | Method | Request | Response |
|---|--------|---------|----------|
| 1 | `CreateProfile` | `CreateProfileRequest` | `ProfileResponse` |
| 2 | `GetProfileByID` | `GetProfileByIDRequest` | `ProfileResponse` |
| 3 | `GetProfileByUserID` | `GetProfileByUserIDRequest` | `ProfileResponse` |
| 4 | `UpdateProfile` | `UpdateProfileRequest` | `StandardResponse` |
| 5 | `GetAllMenus` | `GetAllMenusRequest` | `MenuListResponse` |
| 6 | `GetMenusByAccountType` | `GetMenusByAccountTypeRequest` | `MenuListResponse` |

---

## Phase 1: Proto Source Files (Referensi)

Proto files berfungsi sebagai **dokumentasi kontrak** — Go code di-generate manual (hand-written) di folder `protogen/`.

### Step 1 — `user-profile-service/proto/user_profile_api.proto`

Service definition `UserProfileService` dengan 6 RPC methods.

```proto
syntax = "proto3";

package userprofile;

option go_package = "protogen/user-profile-service";

import "user_profile_payload.proto";

service UserProfileService {
    rpc CreateProfile(CreateProfileRequest) returns (ProfileResponse);
    rpc GetProfileByID(GetProfileByIDRequest) returns (ProfileResponse);
    rpc GetProfileByUserID(GetProfileByUserIDRequest) returns (ProfileResponse);
    rpc UpdateProfile(UpdateProfileRequest) returns (StandardResponse);
    rpc GetAllMenus(GetAllMenusRequest) returns (MenuListResponse);
    rpc GetMenusByAccountType(GetMenusByAccountTypeRequest) returns (MenuListResponse);
}
```

### Step 2 — `user-profile-service/proto/user_profile_payload.proto`

Semua request/response message definitions.

```proto
syntax = "proto3";

package userprofile;

option go_package = "protogen/user-profile-service";

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
    string image         = 10;
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

## Phase 2: Hand-Written Protogen Files

Mengikuti pattern `identity-service/protogen/identity-service/`:
- `identity_payload.pb.go` → Go structs dengan protobuf+json tags + getter methods
- `identity_api_grpc.pb.go` → Server/Client interface + ServiceDesc + registration

### Step 3 — `user-profile-service/protogen/user-profile-service/user_profile_payload.pb.go`

**Package**: `user_profile_service`

Go structs untuk semua 11 proto messages:

| Struct | Fields | Getter Methods |
|--------|--------|----------------|
| `CreateProfileRequest` | UserId, Bank, Branch, Name, CardNumber, CardProvider, Balance, Currency, AccountType, Image | `GetUserId()`, `GetBank()`, dll. |
| `GetProfileByIDRequest` | Id | `GetId()` |
| `GetProfileByUserIDRequest` | UserId | `GetUserId()` |
| `UpdateProfileRequest` | Id, Bank, Branch, Name, CardNumber | `GetId()`, `GetBank()`, dll. |
| `ProfileResponse` | Id, UserId, Bank, Branch, Name, CardNumber, CardProvider, Balance, Currency, AccountType, Image | `GetId()`, `GetUserId()`, dll. |
| `GetAllMenusRequest` | *(empty)* | — |
| `GetMenusByAccountTypeRequest` | AccountType | `GetAccountType()` |
| `MenuItem` | Id, Index, Type, Title, IconUrl, IsActive | `GetId()`, `GetIndex()`, dll. |
| `MenuListResponse` | Menus `[]*MenuItem` | `GetMenus()` |
| `StandardResponse` | Code, Description | `GetCode()`, `GetDescription()` |

**Tag format** (mengikuti identity-service pattern):

```go
type CreateProfileRequest struct {
    UserId       string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
    Bank         string `protobuf:"bytes,2,opt,name=bank,proto3" json:"bank,omitempty"`
    // ...
}
```

### Step 4 — `user-profile-service/protogen/user-profile-service/user_profile_api_grpc.pb.go`

Komponen:

| Komponen | Deskripsi |
|----------|-----------|
| `UserProfileServiceServer` | Interface server — 6 method signatures |
| `UnimplementedUserProfileServiceServer` | Struct dengan default `codes.Unimplemented` return |
| `UserProfileServiceClient` | Interface client — 6 method signatures |
| `userProfileServiceClient` (private) | Client implementation via `grpc.ClientConnInterface` |
| `NewUserProfileServiceClient(cc)` | Constructor client |
| `UserProfileService_ServiceDesc` | `grpc.ServiceDesc` — ServiceName: `"userprofile.UserProfileService"` |
| `RegisterUserProfileServiceServer(s, srv)` | Register server ke `grpc.ServiceRegistrar` |
| `_UserProfileService_*_Handler` | 6 individual handler functions per RPC |

**Service name constant:**

```go
const userProfileServiceName = "/userprofile.UserProfileService/"
```

**ServiceDesc:**

```go
var UserProfileService_ServiceDesc = grpc.ServiceDesc{
    ServiceName: "userprofile.UserProfileService",
    HandlerType: (*UserProfileServiceServer)(nil),
    Methods: []grpc.MethodDesc{
        {MethodName: "CreateProfile", Handler: _UserProfileService_CreateProfile_Handler},
        {MethodName: "GetProfileByID", Handler: _UserProfileService_GetProfileByID_Handler},
        {MethodName: "GetProfileByUserID", Handler: _UserProfileService_GetProfileByUserID_Handler},
        {MethodName: "UpdateProfile", Handler: _UserProfileService_UpdateProfile_Handler},
        {MethodName: "GetAllMenus", Handler: _UserProfileService_GetAllMenus_Handler},
        {MethodName: "GetMenusByAccountType", Handler: _UserProfileService_GetMenusByAccountType_Handler},
    },
    Streams:  []grpc.StreamDesc{},
    Metadata: "user_profile_api.proto",
}
```

---

## Phase 3: gRPC Handler Implementation

Layer baru `internal/grpchandler/` yang **reuse** existing `repository.ProfileRepository` dan `repository.MenuRepository`.

### Step 5 — `user-profile-service/internal/grpchandler/profile.go`

```
GrpcServer struct
├── ProfileRepo  *repository.ProfileRepository
├── MenuRepo     *repository.MenuRepository
└── embed UnimplementedUserProfileServiceServer
```

**4 Profile RPC implementations:**

| Method | Flow | Error Mapping |
|--------|------|---------------|
| `CreateProfile` | Validate `user_id` required → `createReqToModel()` → `ProfileRepo.CreateProfile()` → `profileToProto()` | validation → `InvalidArgument`, DB error → `Internal` |
| `GetProfileByID` | Validate `id` required → `ProfileRepo.GetProfileByID()` → `profileToProto()` | `sql.ErrNoRows` → `NotFound`, DB error → `Internal` |
| `GetProfileByUserID` | Validate `user_id` required → `ProfileRepo.GetProfileByUserID()` → `profileToProto()` | `sql.ErrNoRows` → `NotFound`, DB error → `Internal` |
| `UpdateProfile` | Validate `id` required → `updateReqToModel()` → `ProfileRepo.UpdateProfile()` → `StandardResponse{200, "Data pengguna berhasil diubah."}` | `sql.ErrNoRows` → `NotFound`, DB error → `Internal` |

### Step 6 — `user-profile-service/internal/grpchandler/menu.go`

Methods pada `GrpcServer` yang sama:

| Method | Flow | Error Mapping |
|--------|------|---------------|
| `GetAllMenus` | `MenuRepo.GetAllMenus()` → `menusToProto()` → `MenuListResponse` | DB error → `Internal` |
| `GetMenusByAccountType` | Validate `account_type` required → `MenuRepo.GetMenusByAccountType()` → `menusToProto()` | DB error → `Internal` |

### Step 7 — `user-profile-service/internal/grpchandler/converter.go`

Helper functions untuk konversi domain models ↔ proto messages:

```go
// profileToProto converts models.Profile to proto ProfileResponse
func profileToProto(p *models.Profile) *pb.ProfileResponse

// menusToProto converts []models.Menu to []*pb.MenuItem
func menusToProto(menus []models.Menu) []*pb.MenuItem

// createReqToModel converts proto CreateProfileRequest to models.CreateProfileRequest
func createReqToModel(req *pb.CreateProfileRequest) models.CreateProfileRequest

// updateReqToModel converts proto UpdateProfileRequest to (id, models.EditProfileRequest)
func updateReqToModel(req *pb.UpdateProfileRequest) (string, models.EditProfileRequest)
```

**Field mapping `profileToProto`:**

| models.Profile | pb.ProfileResponse |
|----------------|-------------------|
| `ID` | `Id` |
| `*UserID` (pointer) | `UserId` (string, "" if nil) |
| `Bank` | `Bank` |
| `Branch` | `Branch` |
| `Name` | `Name` |
| `CardNumber` | `CardNumber` |
| `CardProvider` | `CardProvider` |
| `Balance` (int64) | `Balance` (int64) |
| `Currency` | `Currency` |
| `AccountType` | `AccountType` |
| `Image` | `Image` |

**Field mapping `menusToProto`:**

| models.Menu | pb.MenuItem |
|-------------|-------------|
| `ID` (string) | `Id` (string) |
| `Index` (int) | `Index` (int32) |
| `Type` | `Type` |
| `Title` | `Title` |
| `IconURL` | `IconUrl` |
| `IsActive` (bool) | `IsActive` (bool) |

---

## Phase 4: Server Infrastructure Changes

### Step 8 — Modify `user-profile-service/internal/server/server.go`

**Perubahan pada `Server` struct:**

```go
type Server struct {
    DB       *sql.DB
    Router   chi.Router
    Port     string
    GRPCPort string          // ← NEW
}
```

**Perubahan pada `NewServer`:** Tambah parameter `grpcPort string`.

**Method baru `StartGRPC() error`:**

```
StartGRPC()
├── net.Listen("tcp", ":"+GRPCPort)
├── grpc.NewServer()
├── grpchandler.GrpcServer{ProfileRepo, MenuRepo}
├── pb.RegisterUserProfileServiceServer(grpcServer, grpcHandler)
├── grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
├── log.Printf("gRPC server started on :%s", GRPCPort)
└── grpcServer.Serve(listener)
```

**Import baru:**

```go
import (
    "net"
    pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
    "github.com/bankease/user-profile-service/internal/grpchandler"
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
)
```

### Step 9 — Modify `user-profile-service/cmd/server/main.go`

**Perubahan:**

```go
func main() {
    // ... existing env loading ...
    grpcPort := GetEnv("GRPC_PORT", "9302")    // ← NEW

    // ... existing DB + migration ...

    srv := server.NewServer(database, port, azureSASURL, azureContainer, jwtSecret, grpcPort)  // ← ADD grpcPort

    // Start gRPC server in background goroutine         // ← NEW
    go func() {
        if err := srv.StartGRPC(); err != nil {
            log.Fatalf("gRPC server failed: %v", err)
        }
    }()

    // Existing HTTP server (blocking)
    log.Printf("HTTP server started on :%s", port)
    if err := srv.Start(); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
```

---

## Phase 5: Dependencies & Docker

### Step 10 — Update `user-profile-service/go.mod`

Tambah dependency:

```
google.golang.org/grpc v1.48.0
google.golang.org/protobuf v1.28.1  (transitive)
golang.org/x/net (transitive, sudah ada)
```

Setelah edit, jalankan:

```bash
cd user-profile-service
go mod tidy
```

### Step 11 — Update `user-profile-service/Dockerfile`

```dockerfile
EXPOSE 8080
EXPOSE 9302
```

### Step 12 — Update `user-profile-service/docker-compose.yml`

```yaml
services:
  app:
    ports:
      - "8080:8080"
      - "9302:9302"    # gRPC
```

### Step 13 — Update `user-profile-service/.env.example`

```env
GRPC_PORT=9302
```

---

## File Summary

### File Baru (7 files)

| # | File Path | Deskripsi |
|---|-----------|-----------|
| 1 | `proto/user_profile_api.proto` | Service definition — 6 RPCs (referensi) |
| 2 | `proto/user_profile_payload.proto` | Message definitions (referensi) |
| 3 | `protogen/user-profile-service/user_profile_payload.pb.go` | 11 Go structs + getters |
| 4 | `protogen/user-profile-service/user_profile_api_grpc.pb.go` | Interface + ServiceDesc + registration |
| 5 | `internal/grpchandler/profile.go` | GrpcServer struct + 4 profile RPCs |
| 6 | `internal/grpchandler/menu.go` | 2 menu RPCs |
| 7 | `internal/grpchandler/converter.go` | Model ↔ proto converter helpers |

### File yang Dimodifikasi (6 files)

| # | File Path | Perubahan |
|---|-----------|-----------|
| 1 | `internal/server/server.go` | + `GRPCPort` field, + `StartGRPC()` method, + imports |
| 2 | `cmd/server/main.go` | + `GRPC_PORT` env, + pass `grpcPort`, + gRPC goroutine |
| 3 | `go.mod` | + `google.golang.org/grpc` dependency |
| 4 | `Dockerfile` | + `EXPOSE 9302` |
| 5 | `docker-compose.yml` | + port mapping `9302:9302` |
| 6 | `.env.example` | + `GRPC_PORT=9302` |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────┐
│            user-profile-service                  │
│                                                  │
│  ┌──────────────────┐  ┌──────────────────────┐ │
│  │   HTTP Server     │  │    gRPC Server        │ │
│  │   (chi router)    │  │    (google.golang.    │ │
│  │   Port 8080       │  │     org/grpc)         │ │
│  │                   │  │    Port 9302          │ │
│  │  /api/profile     │  │                       │ │
│  │  /api/menu        │  │  UserProfileService   │ │
│  │  /api/upload      │  │  ├ CreateProfile      │ │
│  │                   │  │  ├ GetProfileByID     │ │
│  └────────┬──────────┘  │  ├ GetProfileByUserID │ │
│           │              │  ├ UpdateProfile      │ │
│           │              │  ├ GetAllMenus        │ │
│           │              │  └ GetMenusByAccType  │ │
│           │              └──────────┬────────────┘ │
│           │                         │               │
│           ▼                         ▼               │
│  ┌──────────────────────────────────────────────┐  │
│  │              Repository Layer                 │  │
│  │  ProfileRepository    MenuRepository          │  │
│  └──────────────────────┬────────────────────────┘  │
│                         │                            │
│                         ▼                            │
│  ┌──────────────────────────────────────────────┐  │
│  │              PostgreSQL (bankease_db)          │  │
│  │   Tables: profile, menu                       │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

**Key insight**: HTTP handlers dan gRPC handlers **berbagi** repository layer yang sama. Tidak ada duplikasi logic database.

---

## Verification Checklist

### Build & Compile

- [ ] `cd user-profile-service && go build ./...` — compile pass tanpa error
- [ ] `docker build -t user-profile-service .` — build sukses

### Runtime

- [ ] `docker compose up --build` — kedua port 8080 + 9302 listening
- [ ] gRPC health check: `grpcurl -plaintext localhost:9302 grpc.health.v1.Health/Check` → `SERVING`

### REST Regression (port 8080) — Semua endpoint existing tetap berfungsi

- [ ] `GET /api/profile/{id}` → 200 + profile data
- [ ] `GET /api/profile/user/{user_id}` → 200 + profile data
- [ ] `POST /api/profile` → 201 + created profile
- [ ] `PUT /api/profile/{id}` → 200 + success message
- [ ] `GET /api/profile` (JWT) → 200 + my profile
- [ ] `GET /api/menu` → 200 + semua menus
- [ ] `GET /api/menu/REGULAR` → 200 + only REGULAR menus
- [ ] `GET /api/menu/PREMIUM` → 200 + semua menus
- [ ] `POST /api/upload/image` → 200 + upload URL

### gRPC Functional Test (port 9302)

- [ ] `CreateProfile` → profile created, `ProfileResponse` returned
- [ ] `GetProfileByID` → profile returned
- [ ] `GetProfileByUserID` → profile returned
- [ ] `UpdateProfile` → `StandardResponse{200, "Data pengguna berhasil diubah."}`
- [ ] `GetAllMenus` → 9 menus returned
- [ ] `GetMenusByAccountType("REGULAR")` → only REGULAR menus
- [ ] `GetMenusByAccountType("PREMIUM")` → semua menus (REGULAR + PREMIUM)

### gRPC Error Cases

- [ ] `GetProfileByID` dengan ID tidak ada → gRPC status `NotFound`
- [ ] `GetProfileByUserID` dengan user_id tidak ada → gRPC status `NotFound`
- [ ] `CreateProfile` tanpa `user_id` → gRPC status `InvalidArgument`
- [ ] `UpdateProfile` dengan ID tidak ada → gRPC status `NotFound`
- [ ] `GetMenusByAccountType` tanpa `account_type` → gRPC status `InvalidArgument`

---

## Keputusan Teknis

| Keputusan | Alasan |
|-----------|--------|
| **Tanpa protoc toolchain** | Mengikuti pattern identity-service — proto files sebagai referensi, Go code hand-written di `protogen/` |
| **Package `user_profile_service`** | Konsisten dengan identity-service yang pakai `identity_service` (underscore) |
| **gRPC port 9302** | Sesuai BFF spec dan system-patterns.instructions.md |
| **Tanpa interceptor** | Profile-service dipanggil secara internal oleh BFF; BFF yang handle auth. Profile-service trust internal calls |
| **REST tetap jalan** | gRPC bersifat additive, bukan replacement. Mobile app tetap bisa akses langsung via REST saat dev |
| **Upload TIDAK di gRPC** | BFF upload langsung ke Azure Blob Storage — tidak perlu forward file besar via gRPC |
| **Reuse repository layer** | gRPC handlers memanggil repository yang sama dengan HTTP handlers — zero duplikasi logic |

---

## Scope Excluded

- ❌ Unit tests (task terpisah — target ≥ 90% coverage)
- ❌ Perubahan ke identity-service
- ❌ Implementasi BFF service
- ❌ gRPC interceptor chain (tidak dibutuhkan untuk downstream service)
- ❌ gRPC-gateway (hanya BFF yang pakai grpc-gateway, profile-service murni gRPC server)
