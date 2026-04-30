#!/usr/bin/env bash
set -euo pipefail
migrate -path migrations -database "$DATABASE_URL" down -all || true
migrate -path migrations -database "$DATABASE_URL" up
go run cmd/seed/main.go
