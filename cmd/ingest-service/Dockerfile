# Build stage
FROM golang:1.25 AS builder

WORKDIR /app

# Copy module files from repo root
COPY go.mod go.sum ./
RUN go mod download

# Copy service source
COPY cmd/ingest-service/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ingest-service

# Runtime stage
FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/ingest-service .

CMD ["./ingest-service"]
