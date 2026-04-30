include .env
export

.PHONY: dev build test migrate-up migrate-down seed db-reset

dev:
	go run cmd/api/main.go

build:
	go build ./...

test:
	go test ./...

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

seed:
	go run cmd/seed/main.go

db-reset:
	migrate -path migrations -database "$(DATABASE_URL)" down -all || true
	migrate -path migrations -database "$(DATABASE_URL)" up
	go run cmd/seed/main.go
