# Service Discoverer

A dynamic service discovery system for Nginx that automatically updates reverse proxy configurations based on JSON service definitions.

## Features

- Dynamic service registration through JSON files
- Automatic Nginx configuration updates
- Real-time service discovery
- Docker-based deployment
- Support for multiple backend services

## Quick Start

```bash
# Clone the repository
git clone https://github.com/will-wright-eng/discoverer.git
cd discoverer

# Start the system
docker-compose up --build
```

## Service Definition

Add service definitions to the `services` directory:

```json
{
    "name": "api",
    "host": "test-api",
    "port": 8080,
    "path": "/api",
    "protocol": "http"
}
```

## Testing

1. Start test services:
```bash
docker-compose up --build
```

2. Test endpoints:
```bash
# Test API service
curl http://localhost/api

# Test web app
curl http://localhost/app
```

3. Add new service:
```bash
# Add new service definition
cat > services/test.json <<EOF
{
    "name": "test",
    "host": "test-api",
    "port": 8080,
    "path": "/test",
    "protocol": "http"
}
EOF

# Test new endpoint
curl http://localhost/test
```

## Project Structure

```
.
├── cmd
│   ├── discoverer        # Main service
│   └── test-service      # Test backend
├── internal
│   └── service          # Core logic
├── services             # Service definitions
└── docker-compose.yml   # Deployment configuration
```

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Build binary
make build
```
