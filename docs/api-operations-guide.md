# API Operations Guide

This guide explains how to operate the API for:

- registration
- confirmation
- authentication
- tenant/company setup
- team creation
- project creation
- task creation

## Base URL

All endpoints use this base prefix:

```txt
/api/v1
```

Local example:

```bash
http://localhost:8080/api/v1
```

## Response format

Successful responses use:

```json
{
  "success": true,
  "data": {}
}
```

Errors use:

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request"
  }
}
```

## Important notes

1. There is currently no separate account confirmation endpoint in the API.
2. `POST /auth/register` already creates the tenant/company and the first user in one step.
3. For API clients, the easiest login flow is `POST /auth/cli/login` because it returns both the access token and refresh token in JSON.
4. Private endpoints require `Authorization: Bearer <access_token>`.
5. Team, project, and task creation require a premium-capable account. If the account is not premium, the API returns HTTP `402` with code `PREMIUM_REQUIRED`.

## 1. Registration

Use registration to create:

- the company/tenant
- the first user
- the initial role and permissions

Endpoint:

```http
POST /api/v1/auth/register
```

Request body:

```json
{
  "company_name": "Acme Studio",
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "StrongPassword123!"
}
```

Example:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Acme Studio",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "StrongPassword123!"
  }'
```

Expected response:

```json
{
  "success": true,
  "data": {
    "message": "registered"
  }
}
```

## 2. Confirmation

There is no confirmation endpoint implemented right now.

Current behavior:

- registration immediately creates the company and user
- no email verification route exists in the current API
- no manual confirmation token route exists in the current API

If you need confirmation later, it will need to be added as a new feature.

## 3. Authentication

### Option A: CLI/API login

Recommended for Postman, scripts, CLI tools, and backend integrations.

Endpoint:

```http
POST /api/v1/auth/cli/login
```

Request body:

```json
{
  "email": "jane@example.com",
  "password": "StrongPassword123!",
  "company_slug": "acme-studio"
}
```

Example:

```bash
curl -X POST http://localhost:8080/api/v1/auth/cli/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "password": "StrongPassword123!",
    "company_slug": "acme-studio"
  }'
```

Expected response:

```json
{
  "success": true,
  "data": {
    "access_token": "<jwt>",
    "refresh_token": "<opaque_refresh_token>",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

Use the access token like this:

```http
Authorization: Bearer <access_token>
```

### Option B: Browser login

Use this if you want the refresh token stored in cookies.

Endpoint:

```http
POST /api/v1/auth/login
```

This route returns session metadata in JSON and sets refresh/access cookies.

## 4. Check current user

After login, validate the token and inspect the account.

Endpoint:

```http
GET /api/v1/me
```

Example:

```bash
curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer <access_token>"
```

This response includes values such as:

- `user_id`
- `company_id`
- `email`
- `role`
- `permission_names`
- `account_type`
- `is_premium`

## 5. Tenant / Company operations

There is no separate "create tenant" endpoint because registration already creates the tenant.

Available tenant/company endpoints:

### Get current company

```http
GET /api/v1/companies/current
```

Example:

```bash
curl http://localhost:8080/api/v1/companies/current \
  -H "Authorization: Bearer <access_token>"
```

### Update current company name

```http
PATCH /api/v1/companies/current
```

Request body:

```json
{
  "name": "Acme Studio MX"
}
```

Example:

```bash
curl -X PATCH http://localhost:8080/api/v1/companies/current \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Studio MX"
  }'
```

## 6. Create a team

Endpoint:

```http
POST /api/v1/teams
```

Request body:

```json
{
  "name": "Platform Team",
  "slug": "platform-team"
}
```

Example:

```bash
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Platform Team",
    "slug": "platform-team"
  }'
```

Expected response fields:

- `id`
- `name`
- `slug`

Save the returned team `id`. You will use it for team member queries and team project queries.

## 7. Create a project

Project creation requires the team ID in the `X-Team-ID` header.

Endpoint:

```http
POST /api/v1/projects
```

Headers:

```http
Authorization: Bearer <access_token>
X-Team-ID: <team_id>
Content-Type: application/json
```

Request body:

```json
{
  "name": "Customer Portal",
  "description": "Main delivery project",
  "icon": "rocket",
  "sprint_size": 14,
  "sprint_start_date": "2026-05-16",
  "members": []
}
```

Example:

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Team-ID: <team_id>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Customer Portal",
    "description": "Main delivery project",
    "icon": "rocket",
    "sprint_size": 14,
    "sprint_start_date": "2026-05-16",
    "members": []
  }'
```

Expected response fields:

- `id`
- `team_id`
- `name`
- `description`
- `icon`

Save the returned project `id`.

## 8. Read the project board

Each project uses a board for task columns.

Endpoint:

```http
GET /api/v1/projects/:id/board
```

Example:

```bash
curl http://localhost:8080/api/v1/projects/<project_id>/board \
  -H "Authorization: Bearer <access_token>"
```

The default task column key is `todo`.

## 9. Create a task

Endpoint:

```http
POST /api/v1/projects/:id/tasks
```

Minimal request body:

```json
{
  "title": "Create onboarding screen",
  "description": "Initial UI and API integration",
  "column_key": "todo",
  "priority": "high"
}
```

Full example:

```bash
curl -X POST http://localhost:8080/api/v1/projects/<project_id>/tasks \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Create onboarding screen",
    "description": "Initial UI and API integration",
    "column_key": "todo",
    "priority": "high",
    "due_date": "2026-05-30",
    "assignees": [],
    "watchers": [],
    "story_points": 3,
    "checklist": [
      { "text": "Design draft", "completed": false },
      { "text": "API contract review", "completed": false }
    ],
    "tags": ["frontend", "api"],
    "time_entries": [
      { "entry_date": "2026-05-16", "minutes_spent": 30 }
    ]
  }'
```

Expected response fields:

- `id`
- `title`
- `description`
- `priority`
- `column_key`
- `due_date`
- `checklist`
- `time_entries`

## 10. Recommended end-to-end order

Use the API in this order:

1. Register with `POST /auth/register`
2. Log in with `POST /auth/cli/login`
3. Verify the account with `GET /me`
4. Read the tenant with `GET /companies/current`
5. Create a team with `POST /teams`
6. Create a project with `POST /projects` and `X-Team-ID`
7. Read the board with `GET /projects/:id/board`
8. Create tasks with `POST /projects/:id/tasks`

## 11. Useful extra endpoints

List teams:

```http
GET /api/v1/teams
```

List projects for one team:

```http
GET /api/v1/teams/:id/projects
```

List projects using the selected team header:

```http
GET /api/v1/projects
```

Header:

```http
X-Team-ID: <team_id>
```

Get one task:

```http
GET /api/v1/tasks/:id
```

Move a task to another column:

```http
PATCH /api/v1/tasks/:id/move
```

Body:

```json
{
  "column_key": "in_progress"
}
```
