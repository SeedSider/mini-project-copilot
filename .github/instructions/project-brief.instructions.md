---
applyTo: "**"
---

# Project Brief

## Nama Project

**BankEase** — Backend microservices untuk mobile banking app (BRI BRICaMS ecosystem).

## Deskripsi

BankEase adalah platform backend microservices yang menyediakan layanan identitas, profil pengguna, dan menu homepage untuk mobile app (React Native/Expo). Awalnya menggunakan MSW mock, sekarang sudah memiliki backend nyata dengan arsitektur microservices. Proyek ini merupakan bagian dari POC AI-assisted development menggunakan GitHub Copilot untuk BRI.

## Konteks Bisnis

- BRI memiliki backlog menumpuk dengan resource developer terbatas
- Divisi IT BRI mengadopsi AI (GitHub Copilot) untuk mempercepat development
- Tim backend: **Usman & Sintong** (Ecomindo), Tim frontend: **Arief, Fajri, Rizky, Maul**
- Timeline mulai: **April 2026**, Review oleh: **Tech Lead BRI**
- Standar kualitas: SonarQube approved, unit test coverage ≥ 90%, code duplication < 3%

## Tujuan Utama

1. Menyediakan backend microservices production-ready sebagai pengganti MSW mock
2. Mengelola identitas pengguna (registrasi, autentikasi, JWT)
3. Mengelola data profil pengguna bank (CRUD) dan menu homepage
4. Menyediakan BFF (Backend for Frontend) sebagai single entry point untuk mobile app
5. Menerapkan AI-assisted development untuk efisiensi

## Scope

- **In scope**:
  - **3 backend services**: identity-service, user-profile-service, bff-service
  - Identity: SignUp, SignIn, GetMe (JWT auth)
  - Profile: CRUD profil pengguna, upload image (Azure Blob)
  - Menu: Get all menus, filter by accountType (REGULAR/PREMIUM)
  - BFF: REST entry point → gRPC orchestration ke downstream services
  - Database PostgreSQL per service
  - Docker Compose untuk development
- **Out of scope**:
  - Multi-tenancy
  - Production deployment (Kubernetes)
  - Email verification, forgot password, refresh token (future)

## Services

### 1. identity-service (SELESAI)

- Autentikasi: SignUp, SignIn, GetMe
- JWT HS256, bcrypt password hashing
- PostgreSQL (tabel `users`)
- Asal: dikembangkan di project terpisah, kemudian disatukan ke monorepo ini
- Module path: `bitbucket.bri.co.id/scm/addons/addons-identity-service`

### 2. user-profile-service (SELESAI)

- Profil pengguna: CRUD + image upload
- Menu homepage: filter by accountType
- Search: exchange rates, interest rates, branches
- PostgreSQL (tabel `profile`, `menu`, `exchange_rate`, `interest_rate`, `branch`)
- REST API (chi router) + gRPC (port 9302)
- Folder structure: `server/` pattern (sama dengan identity-service)

### 3. bff-service (SELESAI)

- Single entry point untuk mobile app (REST via manual gateway)
- Orchestrate calls ke identity-service + user-profile-service via gRPC
- JWT verification lokal
- Upload image langsung ke Azure Blob Storage
- Docker Compose full stack (5 containers) running & verified

## Referensi Arsitektur

- **`addons-issuance-lc-service`** (BRI BRICaMS ecosystem) — pattern gRPC + grpc-gateway, interceptor chain, ServiceConnection, Provider DB, JWT auth
- Source repo: Bitbucket (`bitbucket.bri.co.id/scm/addons/`)
