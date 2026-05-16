# Rust CLI Requirements for `paladin`

This document describes the backend contract and client-side behavior needed for a Rust CLI that manages tenants, teams, projects, and tasks from the terminal.

## Goals

- Provide a `git`-style command experience.
- Authenticate from a terminal without relying on browser cookies.
- Persist a local session securely.
- Support workspace-local context like active team and active project.

## Recommended command shape

```text
paladin auth register
paladin auth login
paladin auth whoami

paladin team list
paladin team create
paladin team add-member

paladin project list
paladin project create
paladin project use

paladin task list
paladin task create
paladin task show
paladin task update
paladin task move
paladin task delete

paladin pull
paladin push
```

## Backend endpoints already suitable for CLI

### Public auth

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/cli/login`
- `POST /api/v1/auth/cli/refresh`
- `POST /api/v1/auth/recover`
- `POST /api/v1/auth/reset-password`

### Private auth

- `POST /api/v1/auth/cli/logout`
- `GET /api/v1/me`

### Team, project, and task APIs

- `GET /api/v1/teams`
- `POST /api/v1/teams`
- `PUT /api/v1/teams/:id`
- `DELETE /api/v1/teams/:id`
- `GET /api/v1/teams/:id/projects`
- `GET /api/v1/teams/:id/members`
- `POST /api/v1/teams/:id/members`
- `GET /api/v1/projects`
- `POST /api/v1/projects`
- `GET /api/v1/projects/:id`
- `PUT /api/v1/projects/:id`
- `DELETE /api/v1/projects/:id`
- `GET /api/v1/projects/:id/members`
- `PUT /api/v1/projects/:id/members`
- `GET /api/v1/projects/:id/board`
- `POST /api/v1/projects/:id/tasks`
- `GET /api/v1/tasks/:id`
- `PUT /api/v1/tasks/:id`
- `DELETE /api/v1/tasks/:id`
- `PATCH /api/v1/tasks/:id/move`

## CLI auth contract

### Login request

`POST /api/v1/auth/cli/login`

```json
{
  "email": "owner@example.com",
  "password": "ChangeMe123!",
  "company_slug": "demo-company"
}
```

### Login response

```json
{
  "success": true,
  "data": {
    "access_token": "jwt",
    "refresh_token": "opaque-token",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

### Refresh request

`POST /api/v1/auth/cli/refresh`

```json
{
  "refresh_token": "opaque-token"
}
```

### Logout request

`POST /api/v1/auth/cli/logout`

Headers:

- `Authorization: Bearer <access_token>`

Body:

```json
{
  "refresh_token": "opaque-token"
}
```

## Context handling in the CLI

The CLI should track two kinds of context:

- Global account/session config in the user config directory.
- Optional per-workspace config in `.paladin/config.toml`.

Recommended local config structure:

```toml
api_url = "http://localhost:8080"
company_slug = "demo-company"
team_id = "uuid"
project_id = "uuid"
```

Recommended global config structure:

```toml
default_profile = "local"

[profiles.local]
api_url = "http://localhost:8080"
email = "owner@example.com"
company_slug = "demo-company"
access_token = "..."
refresh_token = "..."
```

## Headers the CLI should send

- `Authorization: Bearer <access_token>` for private requests.
- `X-Team-ID: <uuid>` when listing or creating projects for a selected team.

## Suggested Rust crates

- `clap` for command parsing
- `reqwest` for HTTP
- `serde` and `serde_json` for payloads
- `tokio` for async runtime
- `keyring` for secure token storage when available
- `toml` for local and global config files
- `thiserror` for structured CLI errors
- `anyhow` for top-level command orchestration

## `pull` and `push` proposal

### `paladin pull`

- Fetch active team/project context from local config.
- Call remote APIs and write a normalized local cache file like `.paladin/state.json`.
- Useful for offline inspection, scripts, and future sync workflows.

### `paladin push`

- Read `.paladin/state.json` or staged command output.
- Apply pending task or project changes to the backend.
- Best introduced after the CRUD command set is stable.

## Recommended first milestone

1. `auth login`
2. `auth whoami`
3. `team list`
4. `project list`
5. `project create`
6. `task create`
7. `task show`
8. `task update`
9. `task move`

## Notes

- Some write operations in the backend require premium access and will return `402` when the account does not qualify.
- Browser auth still exists for frontend use. The `/auth/cli/*` endpoints are intended for terminal clients and automation.
- Password reset tokens are now opaque, one-time-use tokens stored hashed in the backend and revoked on first successful use.
