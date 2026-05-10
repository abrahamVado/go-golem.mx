# Seeders

Go seeders live under `/seeders` and run with `go run cmd/seed/main.go` or `make seed`.

Production bootstrap:
- `DEFAULT_OWNER_EMAIL`, `DEFAULT_OWNER_PASSWORD`, and `DEFAULT_COMPANY_NAME` create or update the initial owner account and its company.
- The owner is assigned the `Owner` system role with full permissions.
- Demo/sample records are only created when `SEED_DEMO_DATA=true`.
