# Zoho Email Verification Setup

This project now supports email verification with SMTP mail delivery.

## What was added

- registration creates users as `pending_verification`
- registration sends a verification email when mail is enabled
- login is blocked until the email is verified
- verification tokens are stored in the database
- resend verification endpoint is available
- SMTP now supports both:
  - TLS on port `587`
  - implicit SSL on port `465`

## Recommended Zoho SMTP settings

For Zoho Mail free/personal style SMTP:

```env
MAIL_ENABLED=true
SMTP_HOST=smtp.zoho.com
SMTP_PORT=587
SMTP_USERNAME=your-zoho-address@example.com
SMTP_PASSWORD=your-zoho-password-or-app-password
MAIL_FROM=your-zoho-address@example.com
FRONTEND_URL=http://localhost:3000
EMAIL_VERIFICATION_TTL_MINUTES=1440
```

If you prefer implicit SSL:

```env
SMTP_PORT=465
```

## Important Zoho notes

- If Zoho account 2FA is enabled, use an app-specific password for `SMTP_PASSWORD`.
- The `MAIL_FROM` address should match the authenticated Zoho mailbox or alias.
- For free Zoho SMTP, `smtp.zoho.com` is the expected host.

## Verification endpoints

Verify email with token:

```http
GET /api/v1/auth/verify-email?token=<token>
```

Also supported:

```http
POST /api/v1/auth/verify-email
Content-Type: application/json

{
  "token": "<token>"
}
```

Resend verification:

```http
POST /api/v1/auth/resend-verification
Content-Type: application/json

{
  "email": "jane@example.com",
  "company_slug": "acme-studio"
}
```

## Registration and login flow

1. User registers with `POST /api/v1/auth/register`
2. API creates the company and user
3. API sends verification email through Zoho SMTP
4. User opens the verification link
5. API marks `email_verified_at`
6. User can log in with `POST /api/v1/auth/login` or `POST /api/v1/auth/cli/login`

## Database migration

Run:

```bash
make migrate-up
```

This adds:

- `users.email_verified_at`
- `email_verification_tokens`
