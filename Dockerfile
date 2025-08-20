# Build stage
FROM golang:1.23-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git build-base

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Copy the source code
COPY . .

# Build the application with embedded templates and migrations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final stage
FROM alpine:3.21.2

WORKDIR /app

# Install runtime dependencies including docker and docker compose plugin
RUN apk add --no-cache \
    ca-certificates \
    curl \
    docker-cli \
    docker-cli-compose

# Copy the binary from the builder stage
COPY --from=builder /app/app /app/app

# Ensure the binary is executable
RUN chmod +x /app/app
