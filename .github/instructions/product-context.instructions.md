---
applyTo: "**"
---

# Product Context

## Mengapa Project Ini Ada

Mobile app BankEase (React Native/Expo) sebelumnya menggunakan MSW (Mock Service Worker) untuk simulasi backend. Project ini membuat backend nyata dengan arsitektur microservices agar:

- Data persisten di database PostgreSQL
- Behavior lebih realistis (validasi, error handling, auth)
- Siap untuk development dan testing yang lebih serius
- Arsitektur scalable mengikuti pattern BRICaMS (gRPC + grpc-gateway)

## Problem yang Diselesaikan

1. **Mock terlalu sederhana** — MSW tidak mendukung state management, data hilang saat restart
2. **Tidak ada validasi backend** — Request langsung direspons tanpa pengecekan
3. **Tidak cocok untuk testing end-to-end** — Perlu backend real untuk testing flow lengkap
4. **Tidak ada autentikasi** — MSW tidak bisa mensimulasikan JWT auth flow
5. **Mobile app perlu single entry point** — Banyak service di belakang, client butuh 1 endpoint

## User Experience Goals

- **Developer mobile**: Bisa develop dan test mobile app dengan backend real di localhost
- **Response cepat**: API response time minimal karena single-instance local dev
- **Kontrak jelas**: Format response konsisten untuk semua endpoint
- **Auth flow realistis**: SignUp → SignIn → JWT → protected endpoints

## Fitur Utama

### 1. Identitas Pengguna (identity-service)

- Registrasi user baru (SignUp) — username, password, phone
- Login (SignIn) — mendapat JWT token
- Get current user (GetMe) — dari JWT claims
- Password hashing (bcrypt), JWT HS256

### 2. Profil Pengguna (user-profile-service)

- Melihat data profil (bank, branch, name, card info, balance, currency, accountType, image)
- Membuat profil baru (otomatis saat signup melalui BFF)
- Mengubah data profil (bank, branch, name, card number)
- Upload image profil ke Azure Blob Storage

### 3. Menu Homepage (user-profile-service)

- Mengambil semua menu yang tersedia
- Mengambil menu berdasarkan accountType:
  - `REGULAR` → hanya menu tipe REGULAR
  - `PREMIUM` → semua menu (REGULAR + PREMIUM)

### 4. Data Finansial (saving-service)

- Melihat kurs mata uang asing (exchange rates)
- Melihat suku bunga deposito (interest rates)
- Mencari lokasi cabang bank (branches) — pencarian case-insensitive

### 5. BFF — Backend for Frontend (bff-service)

- Single entry point untuk mobile app (REST)
- Orchestrate multi-service calls (contoh: SignUp → identity + profile)
- Proxy ke saving-service untuk data finansial
- JWT verification lokal
- Upload image langsung ke Azure Blob Storage

## Format Response Standard

### user-profile-service

```json
{
  "code": "<status code>",
  "description": "<deskripsi status>"
}
```

### identity-service (error)

```json
{
  "error": true,
  "code": 401,
  "message": "Unauthorized"
}
```
