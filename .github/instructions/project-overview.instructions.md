---
applyTo: "**"
---

# BankEase — Project Overview

**Backend microservices untuk mobile banking app** (BRI BRICaMS ecosystem, React Native/Expo).
Menggantikan MSW mock dengan backend nyata — data persisten, validasi, JWT auth, arsitektur scalable.

## Bisnis

- BRI POC AI-assisted development (GitHub Copilot)
- Tim: Usman & Sintong (backend), Arief/Fajri/Rizky/Maul (frontend)
- Standar: SonarQube approved, unit test ≥ 90%, code duplication < 3%

## 4 Services

| Service | Fungsi | HTTP | gRPC | Database |
|---|---|---:|---:|---|
| **identity-service** | Auth: SignUp, SignIn, GetMe (JWT HS256, bcrypt) | 3031 | 9301 | identity_db |
| **user-profile-service** | Profile CRUD + Menu + Upload image | 8080 | 9302 | bankease_db |
| **saving-service** | Exchange rates, interest rates, branch search | 8081 | 9303 | saving |
| **bff-service** | Single entry point, orchestrate via gRPC, JWT local verify | 3000 | 9090 | — |

## Arsitektur

Mobile app → BFF (REST) → downstream services (gRPC). BFF orchestrates (e.g. SignUp = identity + profile).

## Response Format

- user-profile-service: `{ "code": 200, "description": "Success" }`
- identity-service error: `{ "error": true, "code": 401, "message": "Unauthorized" }`
- saving-service: raw JSON arrays (no wrapper)

## Scope

**In**: 4 services, PostgreSQL per service, Docker Compose, JWT auth, gRPC inter-service
**Out**: Multi-tenancy, Kubernetes, email verification, refresh token
