#!/usr/bin/env bash
set -euo pipefail
migrate -path migrations -database "$DATABASE_URL" down 1
