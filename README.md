# golem.mx — Go API (Gin Multi-Tenant SaaS Backend)

Production-ready Go backend built with Gin for a multi-tenant SaaS platform.

## Features

- JWT authentication (access + refresh)
- Multi-tenant isolation
- RBAC (roles + permissions)
- PostgreSQL + GORM
- SQL migrations
- Seed system
- Audit logging
- Docker-first workflow
- CLI tools (migrate, seed, admin)

--------------------------------------------------

## Architecture

Client (Next.js)
   ->
Nginx Proxy
   ->
Go API
   ->
PostgreSQL

--------------------------------------------------

## Local Development

1. Setup environment

cp .env.example .env

2. Start database

docker compose up -d postgres

3. Install dependencies

go mod tidy

4. Run migrations

make migrate-up

5. Seed database

make seed

6. Run API

make dev

7. Test

curl http://localhost:8080/api/v1/health

--------------------------------------------------

## Production

docker compose up -d --build

Health check:

curl http://localhost/api/v1/health

--------------------------------------------------

## Authentication

Access Token:
- JWT
- Sent via Authorization header
- Short-lived

Refresh Token:
- HttpOnly cookie
- Stored hashed in DB
- Rotated on refresh

--------------------------------------------------

## Multi-Tenancy

All data must include company_id

Rule:
WHERE company_id = current_user.company_id

Never trust client-provided tenant IDs

--------------------------------------------------

## RBAC

Permissions:
users.read
users.create
users.update
users.delete

Roles:
owner, admin, manager, viewer

--------------------------------------------------

## API

Base prefix:
/api/v1

Public:
GET  /api/v1/health
POST /api/v1/auth/login
POST /api/v1/auth/register

Private:
GET /me
GET /users
GET /roles
GET /settings

--------------------------------------------------

## Database

PostgreSQL + GORM

Migrations:
make migrate-up
make migrate-down

Seeds:
make seed

--------------------------------------------------

## CLI

docker exec -it golem-api ./migrate up
docker exec -it golem-api ./seed

--------------------------------------------------

## Environment

PORT=8080
DATABASE_URL=postgres://user:pass@postgres:5432/db?sslmode=disable
JWT_SECRET=secret
REFRESH_SECRET=secret

--------------------------------------------------

## Production Checklist

- Use HTTPS
- Do not expose DB
- Use strong secrets
- Enable backups
- Add rate limiting

--------------------------------------------------


DEBUGGING

docker logs golem-go-api
