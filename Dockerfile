FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GODEBUG=http2client=0

COPY go.mod go.sum ./
RUN set -eux; \
    attempt=0; \
    until [ "$attempt" -ge 5 ]; do \
      go mod download && break; \
      attempt=$((attempt + 1)); \
      echo "go mod download failed, retrying ($attempt/5)..." >&2; \
      sleep $((attempt * 2)); \
    done; \
    [ "$attempt" -lt 5 ]

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/seed ./cmd/seed


FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates && adduser -D -H appuser

COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/migrate /app/migrate
COPY --from=builder /bin/seed /app/seed

COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/scripts/start-prod.sh /app/start-prod.sh

RUN chmod +x /app/start-prod.sh

USER appuser

EXPOSE 8080

CMD ["/app/start-prod.sh"]
