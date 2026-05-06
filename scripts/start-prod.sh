#!/bin/sh
set -eu

/app/migrate up
/app/seed
exec /app/api
