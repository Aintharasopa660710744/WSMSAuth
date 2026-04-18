# ---- Build Stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /auth-service ./cmd/server

# ---- Run Stage ----
FROM scratch

COPY --from=builder /auth-service /auth-service
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8081
ENTRYPOINT ["/auth-service"]
