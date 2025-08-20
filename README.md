# Blockscout VC Sidecar

## Overview

The **Blockscout VC Sidecar** service monitors database changes in Supabase and automatically updates Docker container configurations and restarts services when necessary. This ensures that your Blockscout services are always running with the latest configuration. Intention is to make sure that changes made in Cloud Console are reflected in the running containers.

## Configuration File

To configure the service, use a YAML file with the following structure:

```yaml
supabaseRealtimeUrl: "wss://your-project.supabase.co/realtime/v1/websocket"
supabaseAnonKey: "your-anon-key"
pathToDockerCompose: "./config/docker-compose.yaml"
frontendServiceName: "frontend"
frontendContainerName: "frontend"
backendServiceName: "backend"
backendContainerName: "backend"
statsServiceName: "stats"
statsContainerName: "stats"
proxyServiceName: "proxy"
proxyContainerName: "proxy"
table: "silos"
chainId: 10
```

## Authentication

The service includes basic authentication (username/password) to protect sensitive endpoints while maintaining public access to read-only operations.

### Authentication Configuration

Configure authentication in your config.yaml file:

```yaml
# config.yaml
auth:
  username: "admin"
  password: "your-secure-password"
```

### Protected vs Public Endpoints

#### 🔒 Protected Endpoints (Authentication Required)
- `GET /` - Token Management Dashboard
- `GET /api/v1/tokens/paginated` - List tokens with pagination
- `POST /api/v1/tokens` - Create/update tokens
- `GET /api/v1/blockscout/tokens` - Fetch Blockscout tokens
- `GET /api/v1/blockscout/tokens/:address` - Search specific token

#### 🌐 Public Endpoints (No Authentication Required)
- `GET /api/v1/chains/:chainId/token-infos/:tokenAddress` - Get token information

### Using Authentication

Include credentials in the `Authorization` header or use curl's `-u` flag:

```bash
# For protected endpoints
curl -u username:password \
     http://localhost:8080/api/v1/tokens/paginated

# Or with Authorization header
curl -H "Authorization: Basic $(echo -n 'username:password' | base64 -w0)" \
     http://localhost:8080/api/v1/tokens/paginated

# For public endpoints (no auth required)
curl http://localhost:8080/api/v1/chains/1313161575/token-infos/0x123...
```

### Development Mode

If no `auth.username` and `auth.password` are set in config, authentication is disabled and all endpoints are accessible (useful for development).

## Running the Service with Docker Compose

The service can be deployed using Docker Compose. Below is an example configuration:

```yaml
services:
  sidecar:
    command:
        - sh
        - -c
        - /app/app sidecar --config /app/config/local.yaml
    container_name: blockscout-vc-sidecar
    image: ghcr.io/aurora-is-near/blockscout-vc:latest
    restart: unless-stopped
    environment:
        # Authentication is configured via config.yaml file
        # No environment variables needed for auth
    volumes:
        - ./config:/app/config
        - /var/run/docker.sock:/var/run/docker.sock
        - ./docker-compose.yaml:/app/config/docker-compose.yaml
```

### Important Notes
- Configuration files should be mounted in the `/app/config` directory

### Basic Commands

Start the service:
```bash
docker compose up -d
```

Restart the service:
```bash
docker compose up -d --force-recreate
```

Stop the service:
```bash
docker compose down
```

View logs:
```bash
docker logs -f blockscout-vc-sidecar
```

## Features

- Monitors Supabase database changes in real-time
- Automatically updates Docker Compose environment variables
- Restarts affected services when configuration changes
- Handles multiple service updates efficiently
- Prevents duplicate container restarts
- Validates configuration changes before applying
- Tracks explorer URL changes and updates related environment variables
- Uses template-based environment variable management
- **Token Management Dashboard** with embedded templates
- **Basic Authentication** for protected endpoints
- **Public Token Info Access** for read-only operations

## Development

### Prerequisites

- Go 1.21 or later
- Docker
- Docker Compose

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/blockscout/blockscout-vc-sidecar.git
```

2. Build the binary:
```bash
go build -o blockscout-vc-sidecar
```

3. Run with configuration:
```bash
./blockscout-vc-sidecar --config config/local.yaml
```

### Project Structure

```
blockscout-vc/
├── cmd/
│   └── root.go
│   └── sidecar.go
├── internal/
│   ├── client/        # WebSocket client implementation
│   ├── config/        # Configuration handling
│   ├── docker/        # Docker operations
│   ├── env/           # Environment variable management
│   ├── handlers/      # Event handlers (name, coin, image, explorer)
│   ├── heartbeat/     # Heartbeat logic
│   └── subscription/  # Supabase subscription logic
│   └── worker/        # Worker implementation
├── config/
│   └── local.yaml     # Configuration file
└── main.go
```

## Configuration Options

| Parameter | Description | Required |
|-----------|-------------|----------|
| `supabaseRealtimeUrl` | Supabase Realtime WebSocket URL | Yes |
| `supabaseAnonKey` | Supabase Anonymous Key | Yes |
| `pathToDockerCompose` | Path to the Docker Compose file | Yes |
| `frontendServiceName` | Name of the frontend service | Yes |
| `frontendContainerName` | Name of the frontend container | Yes |
| `backendServiceName` | Name of the backend service | Yes |
| `backendContainerName` | Name of the backend container | Yes |
| `statsServiceName` | Name of the stats service | Yes |
| `statsContainerName` | Name of the stats container | Yes |
| `proxyServiceName` | Name of the proxy service | Yes |
| `proxyContainerName` | Name of the proxy container | Yes |
| `table` | Name of the table to listen to | Yes |
| `chainId` | Chain ID to listen to | Yes |
| `pathToEnvFile` | Path to the environment file | Yes |

## Event Handlers

The sidecar service includes several handlers that respond to database changes:

- **Name Handler**: Updates network name and featured networks configuration
- **Coin Handler**: Updates cryptocurrency symbol and related settings
- **Image Handler**: Updates logo and favicon URLs
- **Explorer Handler**: Updates explorer URL and related environment variables

### Explorer Handler

The Explorer Handler monitors changes to the `explorer_url` field in the database and automatically updates related environment variables:

- `BLOCKSCOUT_HOST`: The main explorer host
- `MICROSERVICE_VISUALIZE_SOL2UML_URL`: Visualization service URL
- `NEXT_PUBLIC_FEATURED_NETWORKS`: Featured networks configuration for the frontend
- `NEXT_PUBLIC_API_HOST`: Frontend API host
- `NEXT_PUBLIC_APP_HOST`: Frontend app host
- `NEXT_PUBLIC_STATS_API_HOST`: Stats API host
- `NEXT_PUBLIC_VISUALIZE_API_HOST`: Visualization API host
- `STATS__BLOCKSCOUT_API_URL`: Stats service API URL
- `EXPLORER_URL`: Explorer host for nginx configuration
- `BLOCKSCOUT_HTTP_PROTOCOL`: Protocol (http/https) for nginx configuration

When the explorer URL changes, all affected services (backend, frontend, stats, proxy) are automatically restarted.

## Debugging

Enable debug logging by setting the environment variable:
```bash
export LOG_LEVEL=debug
```

## Deployment Guide

### Release and Versioning

Releases are managed via GitHub with canonical versioning (e.g., `0.1.2`). Ensure the versioning follows semantic versioning guidelines.

To release a new version:
1. Create a release on GitHub, specifying the appropriate tag (following semantic versioning guidelines).
2. This will trigger the build and push workflows to create a new Docker image and store it in the GitHub registry.