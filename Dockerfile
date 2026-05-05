FROM golang:1.26-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/migrate ./cmd/migrate

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates wget && adduser -D -H appuser
COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/migrate /app/migrate
COPY --from=builder /app/scripts/start-prod.sh /app/start-prod.sh
RUN chmod +x /app/start-prod.sh
USER appuser
EXPOSE 8080
CMD ["/app/start-prod.sh"]
