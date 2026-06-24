# Build stage — includes go workspace + shared pkg
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY notification-service/ ./notification-service/

WORKDIR /build/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app ./cmd

# Final stage — scratch for minimal image size
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /app
COPY --from=builder /build/notification-service/config.yaml /config.yaml

ENTRYPOINT ["/app"]
