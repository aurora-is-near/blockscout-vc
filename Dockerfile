# Build stage
FROM golang:1.23-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git build-base

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Final stage
FROM alpine:3.17

WORKDIR /app

# Install runtime dependencies including docker cli
RUN apk add --no-cache \
    ca-certificates \
    curl \
    docker-cli

# Copy the binary from the builder stage
COPY --from=builder /app/app /app/app

# Ensure the binary is executable
RUN chmod +x /app/app
