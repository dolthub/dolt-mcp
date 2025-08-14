# Use official golang image as the build stage
FROM golang:1.24.4-alpine3.21 AS builder

# Set build arguments
ARG DOLT_MCP_VERSION=development

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X github.com/dolthub/dolt-mcp/mcp/pkg.DoltMCPServerVersion=${DOLT_MCP_VERSION}" \
    -a -installsuffix cgo \
    -o dolt-mcp-server \
    ./mcp/cmd/dolt-mcp-server

# Use a minimal base image for the final stage
FROM alpine:3.21

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -g 1001 doltmcp && \
    adduser -D -s /bin/sh -u 1001 -G doltmcp doltmcp

# Set the working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/dolt-mcp-server .

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Change ownership of the app directory
RUN chown -R doltmcp:doltmcp /app

# Switch to non-root user
USER doltmcp

# Expose the default HTTP port
EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD if [ "$MCP_MODE" = "http" ]; then \
        wget --quiet --tries=1 --spider http://localhost:${MCP_PORT:-8080}/health || exit 1; \
    else \
        pgrep dolt-mcp-server > /dev/null || exit 1; \
    fi

# Set default environment variables
ENV DOLT_HOST=127.0.0.1
ENV DOLT_PORT=3306
ENV DOLT_USER=root
ENV DOLT_DATABASE=
ENV MCP_MODE=stdio
ENV MCP_PORT=8080

# Create entrypoint script
COPY --chown=doltmcp:doltmcp docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Use entrypoint script
ENTRYPOINT ["/entrypoint.sh"]

# Default command (can be overridden)
CMD []
