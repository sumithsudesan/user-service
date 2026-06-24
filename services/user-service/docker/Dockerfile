# Build stage — includes go workspace + shared pkg
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY user-service/ ./user-service/

WORKDIR /build/user-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app ./cmd

# Final stage — scratch for minimal image size
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /app
COPY --from=builder /build/user-service/config.yaml /config.yaml

EXPOSE 8080
ENTRYPOINT ["/app"]
