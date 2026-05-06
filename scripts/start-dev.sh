#!/bin/sh
set -eu

go run ./cmd/migrate up
go run ./cmd/seed
exec go run ./cmd/api
