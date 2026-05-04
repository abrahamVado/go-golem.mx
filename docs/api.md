# API Documentation

This document defines the high-level API contract for the Go backend.

The API is designed for a SaaS platform with:

- Authentication
- Authorization
- RBAC
- Multi-company support
- Branch-level access
- Settings management
- Audit logging

---

# Base URL

All API routes use the following base prefix:

```txt
/api/v1