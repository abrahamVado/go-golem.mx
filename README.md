# Gin Multi-Tenant SaaS Backend Scaffold

Production-minded Go/Gin API scaffold for a Next.js frontend with PostgreSQL, GORM, SQL migrations, JWT access tokens, HTTP-only refresh cookies, tenant isolation, RBAC, audit logs, seed commands, and future admin CLI/TUI expansion.

## Run locally

```bash
cp .env.example .env
docker compose up -d postgres
go mod tidy
make migrate-up
make seed
make dev
```

Health check:

```bash
curl http://localhost:8080/api/v1/health
```

## Auth architecture

- Access token: short-lived JWT sent in `Authorization: Bearer <token>`.
- Refresh token: opaque token stored in `HttpOnly`, `Secure`, `SameSite` cookie.
- Refresh tokens are hashed in PostgreSQL.
- Refresh rotates tokens and revokes the previous token.
- Logout revokes the refresh token.

## Multi-tenancy

Every tenant-owned module carries `company_id`. Private requests resolve tenant context from the authenticated JWT. Repository methods should always scope queries with the company ID.

## RBAC

Permissions are global. Roles can be system-level or company-level. `user_roles` is company-scoped. Route middleware supports permission, any-permission, and role checks.
