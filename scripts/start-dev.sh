#!/bin/sh
set -eu

export MIGRATIONS_PATH="${MIGRATIONS_PATH:-file:///app/migrations}"

migrate up
go run ./cmd/seed
exec go run ./cmd/api
