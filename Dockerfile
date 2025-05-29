FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download \
 && apk add --no-cache git \
 && go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o person-api ./cmd/person-api

FROM alpine:3.18
RUN apk add --no-cache ca-certificates postgresql-client

WORKDIR /app

COPY --from=builder /go/bin/goose /usr/local/bin
COPY --from=builder /app/apply-migrations.sh .
COPY --from=builder /app/person-api .
COPY --from=builder /app/internal/storage/postgres/migrations ./internal/storage/postgres/migrations

ENV PATH="/usr/local/bin:${PATH}"
ENV DB_DSN="postgres://user:password@${DB_HOST:-db}:${DB_PORT:-5432}/persons?sslmode=disable"

EXPOSE 8080

ENTRYPOINT ["./person-api"]
