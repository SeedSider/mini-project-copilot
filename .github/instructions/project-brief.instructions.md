---
applyTo: '**'
---

# Project Brief

## Nama Project
**BankEase** — Backend API service untuk mobile banking app.

## Deskripsi
BankEase adalah backend service yang berfungsi sebagai pengganti MSW mock untuk mobile app (React Native/Expo). Service ini menyediakan REST API untuk manajemen profil pengguna dan daftar menu homepage berdasarkan tipe akun.

## Tujuan Utama
1. Menyediakan backend API yang production-ready sebagai pengganti mock server (MSW)
2. Mengelola data profil pengguna bank (CRUD)
3. Mengelola daftar menu homepage dengan filtering berdasarkan account type (REGULAR/PREMIUM)
4. Mendukung idempotency untuk operasi transaksional

## Scope
- **In scope**:
  - REST API endpoints: Profile (GET, PUT) dan Menu (GET, GET by accountType)
  - Database PostgreSQL untuk persistensi data
  - Single merchant / tenant
  - Local development environment (localhost:8080)
  - Idempotency support via UUID header
- **Out of scope**:
  - Autentikasi (internal/dev only)
  - Multi-tenancy
  - Production deployment
  - gRPC (menggunakan REST murni, berbeda dari service referensi)

## Referensi Arsitektur
Project ini menggunakan arsitektur yang terinspirasi dari `addons-issuance-lc-service` (BRI CAMS ecosystem) namun disederhanakan untuk kebutuhan REST API murni dengan Go stdlib + chi router.