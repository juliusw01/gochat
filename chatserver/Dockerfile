# syntax=docker/dockerfile:1

# =========================
# Stage 1: Builder
# =========================
FROM golang:1.23 AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/chatserver

# =========================
# Stage 2: Runtime
# =========================
FROM alpine:3.19

# RUN apk --no-cache add ca-certificates curl

WORKDIR /app
COPY --from=builder /app/bin/chatserver .

RUN adduser -D appuser
USER appuser

ENTRYPOINT ["./chatserver"]

# Healthcheck (assumes app exposes HTTP health endpoint) --> Implement health endpoint
# HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
#   CMD curl -f http://localhost:8080/health || exit 1
