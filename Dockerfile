FROM golang:1.23-alpine

WORKDIR /app

# Add debugging tools
RUN apk add --no-cache curl

COPY go.* ./
RUN go mod download

COPY . .

# Add error checking during build
RUN go build -v -o /discoverer cmd/discoverer/main.go

# Add debugging wrapper script
COPY <<'EOF' /entrypoint.sh
#!/bin/sh
set -e

echo "Starting discoverer..."
/discoverer 2>&1
EOF

RUN chmod +x /entrypoint.sh

CMD ["/entrypoint.sh"]
