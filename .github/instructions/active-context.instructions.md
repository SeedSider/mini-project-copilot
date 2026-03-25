---
applyTo: '**'
---

# Active Context

## Current Focus
Service `user-profile-service` sudah **selesai dan terverifikasi**. Docker Compose running, semua 4 endpoint tested dan working.

## Recent Changes
- Service `user-profile-service` dibuat (folder name pengganti `backend/`)
- Semua 14 file Go + SQL + config sudah dibuat
- Go terinstall, `go mod tidy` berhasil → `go.sum` tergenerate
- Docker setup: `Dockerfile` (multi-stage build), `docker-compose.yml`, `.dockerignore`
- `.env.example` diupdate dengan kredensial Docker Compose
- Docker Compose fix: DDL migration (`001_init.sql`) dimount sebagai `01-schema.sql` di init dir agar tables dibuat sebelum seed
- `docker compose up --build` berhasil — kedua container running (bankease-db healthy, bankease-app up)
- Semua endpoint ditest dan verified:
  - GET /api/profile/{id} → 200 + profile JSON
  - PUT /api/profile/{id} → 200 + `{code, description}`
  - GET /api/menu → 200 + 9 menu items
  - GET /api/menu/REGULAR → 200 + 5 REGULAR menus
  - GET /api/menu/PREMIUM → 200 + semua 9 menus (bisnis logic benar)

## What's Working
- Semua source code terverifikasi tanpa error
- Docker Compose: PostgreSQL 17 + app running di localhost:8080
- Seed data auto-loaded (1 profile + 9 menu items)
- Semua 4 endpoint API berfungsi dengan benar
- Business logic menu filter (PREMIUM → semua, REGULAR → hanya REGULAR) terverifikasi
- Profile update + persistence terverifikasi

## Next Steps
- Project backend selesai. Possible future enhancements:
  1. Frontend integration — hubungkan mobile app ke backend real
  2. Unit tests — tambahkan Go tests untuk handler dan repository
  3. Graceful shutdown — handle SIGINT untuk clean DB close

## Active Decisions
- Menggunakan `database/sql` (stdlib) bukan GORM — lebih ringan, sesuai scope project
- Chi router dipilih karena lightweight dan idiomatis Go
- Tidak ada auth layer — ini dev/internal service
- Docker Compose untuk development — PostgreSQL + app dalam satu command
- Seed data dijalankan otomatis via PostgreSQL init directory

## Important Patterns
- Semua response harus berformat `{code, description}`
- Menu filter: PREMIUM → semua menu, REGULAR → hanya menu REGULAR
- Balance dalam minor unit (cents/pence)