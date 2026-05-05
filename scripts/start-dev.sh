#!/bin/sh
set -eu

go run ./cmd/migrate up
exec go run ./cmd/api
