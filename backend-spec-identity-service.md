# Identity Service — Backend API Spec (Sign Up & Sign In)

## Tujuan

Menyediakan layanan registrasi (sign up) dan autentikasi (sign in) user berbasis gRPC, JWT, dan PostgreSQL, dengan struktur dan teknologi mengikuti arsitektur addons-issuance-lc-service.

## Stack & Infrastruktur

- Bahasa: Go (gRPC + grpc-gateway)
- Database: PostgreSQL
- Auth: JWT (email & password, hash bcrypt)
- ORM: GORM v1 & v2 (Provider pattern)
- Deployment: Dual server (gRPC 9090/9301, HTTP 3000/3031)

## Struktur Folder (mengacu LC service)

```
addons-identity-service/
├── proto/
│   ├── identity_api.proto         # API definitions (SignUp, SignIn)
│   ├── identity_payload.proto     # Request/response messages
├── protogen/identity-service/     # Generated proto Go files
├── server/
│   ├── main.go                    # Entry point + CLI
│   ├── core_config.go             # Config loader
│   ├── core_db.go                 # DB connection
│   ├── api/
│   │   ├── api.go
│   │   ├── identity_auth_api.go   # Handler SignUp, SignIn
│   │   ├── error.go
│   │   └── ...
│   ├── db/
│   │   ├── provider.go
│   │   ├── identity_provider.go   # DB logic user/profile
│   ├── jwt/manager.go             # JWT manager
│   ├── services/service.go        # External gRPC (future)
│   ├── utils/utils.go
│   └── ...
├── Makefile, Dockerfile, .env.example
```

## Database Schema

```sql
-- USERS
CREATE TABLE users (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   username VARCHAR(255) UNIQUE NOT NULL,
   password_hash TEXT NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- PROFILES
CREATE TABLE profiles (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
   full_name VARCHAR(255) NOT NULL,
   phone VARCHAR(50),
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Protobuf API (gRPC)

### identity_api.proto

```proto
service IdentityService {
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
}
```

### identity_payload.proto

```proto
message SignUpRequest {
   string email = 1;
   string password = 2;
   string full_name = 3;
   string phone = 4;
}
message SignUpResponse {
   string user_id = 1;
   string email = 2;
   string full_name = 3;
}
message SignInRequest {
   string email = 1;
   string password = 2;
}
message SignInResponse {
   string user_id = 1;
   string email = 2;
   string full_name = 3;
   string token = 4;
}
```

## Business Logic

### Sign Up (Registration)

1. Validasi input (email format, password ≥ 6, full_name wajib)
2. Cek email sudah terdaftar
3. Hash password (bcrypt)
4. Insert ke tabel users (tx)
5. Insert ke tabel profiles (tx)
6. Return response (user_id, email, full_name)

### Sign In (Login)

1. Cari user by email
2. Compare password (bcrypt)
3. Ambil profile
4. Generate JWT token (exp: 24h)
5. Return response (user_id, email, full_name, token)

### Password Hashing

```go
import "golang.org/x/crypto/bcrypt"
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
```

## Validation Rules

| Field     | Rule               |
| --------- | ------------------ |
| email     | Format email valid |
| password  | Minimal 6 karakter |
| full_name | Tidak boleh kosong |

## Error Handling

- 409: Email sudah terdaftar
- 401: Invalid credentials
- 422: Validation error

## Testing Checklist

- [ ] SignUp berhasil
- [ ] SignUp duplicate email → 409
- [ ] SignIn berhasil
- [ ] SignIn gagal → 401
- [ ] Profile otomatis terbuat saat signup
- [ ] Password tersimpan dalam bentuk hash

## Next Improvements (Future)

- Email verification
- Forgot password
- Refresh token
- Rate limiting login
- Redis session

---

**Catatan:**

- Semua endpoint diakses via HTTP (grpc-gateway) dan gRPC
- JWT token dikirim via header: `Authorization: Bearer <token>`
- Struktur, naming, dan pattern mengikuti LC service (provider, interceptor, error, logging, dsb)
