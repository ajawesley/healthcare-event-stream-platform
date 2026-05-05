# ---------- Build Stage ----------
FROM golang:1.22 AS builder

WORKDIR /app

# Cache modules first
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire repo
COPY . .

# Build static binary for Linux (required for ECS Fargate)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o ingest-service ./cmd/ingest-service

# ---------- Runtime Stage ----------
FROM alpine:3.19

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/ingest-service .

# Expose the port your service listens on
EXPOSE 8080

# Run the service
CMD ["./ingest-service"]
