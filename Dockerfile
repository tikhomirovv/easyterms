# syntax=docker/dockerfile:1

# Build stage: compile the Telegram bot binary.
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/telegram ./cmd/telegram

# Runtime stage: minimal image with only the binary.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
    && adduser -D -H -u 10001 app

WORKDIR /app

COPY --from=builder /out/telegram /app/telegram

USER app

ENTRYPOINT ["/app/telegram"]
