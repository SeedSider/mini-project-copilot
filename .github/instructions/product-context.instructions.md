---
applyTo: '**'
---

# Product Context

## Mengapa Project Ini Ada
Mobile app BankEase (React Native/Expo) sebelumnya menggunakan MSW (Mock Service Worker) untuk simulasi backend. Project ini membuat backend nyata agar:
- Data persisten di database PostgreSQL
- Behavior lebih realistis (validasi, error handling)
- Siap untuk development dan testing yang lebih serius

## Problem yang Diselesaikan
1. **Mock terlalu sederhana** — MSW tidak mendukung state management, data hilang saat restart
2. **Tidak ada validasi backend** — Request langsung direspons tanpa pengecekan
3. **Tidak cocok untuk testing end-to-end** — Perlu backend real untuk testing flow lengkap

## User Experience Goals
- **Developer mobile**: Bisa develop dan test mobile app dengan backend real di localhost
- **Response cepat**: API response time minimal karena single-instance local dev
- **Kontrak jelas**: Format response konsisten untuk semua endpoint

## Fitur Utama
### 1. Profil Pengguna
- Melihat data profil (bank, branch, name, card info, balance, currency, accountType)
- Mengubah data profil (bank, branch, name, card number, card provider, currency)

### 2. Menu Homepage  
- Mengambil semua menu yang tersedia
- Mengambil menu berdasarkan accountType:
  - `REGULAR` → hanya menu tipe REGULAR
  - `PREMIUM` → semua menu (REGULAR + PREMIUM)

## Format Response Standard
```json
{
  "code": "<status code>",
  "description": "<deskripsi status>"
}
```