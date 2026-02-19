FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/aiox-api ./cmd/api

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates curl \
    && addgroup -S aiox && adduser -S aiox -G aiox

COPY --from=builder /bin/aiox-api /usr/local/bin/aiox-api
COPY --from=builder /app/migrations /app/migrations

USER aiox
WORKDIR /app

EXPOSE 8080 50051

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health/live || exit 1

ENTRYPOINT ["aiox-api"]
