# Dolt MCP Server Docker Image

The official Docker image for the Dolt MCP (Model Context Protocol) Server, providing AI assistants with direct access to Dolt databases.

## Quick Start

### HTTP Mode (Recommended for Docker)

```bash
docker run -d \
  --name dolt-mcp-server \
  -p 8080:8080 \
  -e MCP_MODE=http \
  -e DOLT_HOST=your-dolt-host \
  -e DOLT_USER=root \
  -e DOLT_DATABASE=your_database \
  -e DOLT_PASSWORD=your_password \
  dolthub/dolt-mcp:latest
```

### Stdio Mode

```bash
docker run -it --rm \
  -e MCP_MODE=stdio \
  -e DOLT_HOST=your-dolt-host \
  -e DOLT_USER=root \
  -e DOLT_DATABASE=your_database \
  -e DOLT_PASSWORD=your_password \
  dolthub/dolt-mcp:latest
```

## Environment Variables

### Required
- `DOLT_HOST` - Hostname of the Dolt SQL server
- `DOLT_USER` - Username for Dolt server authentication
- `DOLT_DATABASE` - Name of the database to connect to

### Optional
- `DOLT_PASSWORD` - Password for authentication (recommended to use Docker secrets in production)
- `DOLT_PORT` - Dolt server port (default: 3306)
- `MCP_MODE` - Server mode: `http` or `stdio` (default: stdio)
- `MCP_PORT` - HTTP server port (default: 8080, HTTP mode only)

## Docker Compose Example

```yaml
version: '3.8'

services:
  dolt-mcp-server:
    image: dolthub/dolt-mcp:latest
    ports:
      - "8080:8080"
    environment:
      - MCP_MODE=http
      - DOLT_HOST=dolt-server
      - DOLT_PORT=3306
      - DOLT_USER=root
      - DOLT_DATABASE=myapp
      - DOLT_PASSWORD=secret
    depends_on:
      - dolt-server
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  dolt-server:
    image: dolthub/dolt-sql-server:latest
    ports:
      - "3306:3306"
    volumes:
      - dolt_data:/var/lib/dolt
    environment:
      - DOLT_ROOT_PATH=/var/lib/dolt
    restart: unless-stopped

volumes:
  dolt_data:
```

## Using with Claude Desktop

For stdio mode with Claude Desktop, you can run the container and connect to it:

```json
{
  "mcpServers": {
    "dolt-mcp": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "MCP_MODE=stdio",
        "-e", "DOLT_HOST=your-dolt-host",
        "-e", "DOLT_USER=root",
        "-e", "DOLT_DATABASE=your_database",
        "-e", "DOLT_PASSWORD=your_password",
        "dolthub/dolt-mcp:latest"
      ]
    }
  }
}
```

## Security Considerations

- The image runs as a non-root user (`doltmcp:1001`)
- Use Docker secrets or external secret management for passwords in production
- Consider running in a private network when connecting to Dolt servers
- Regular security updates are provided through new image releases

## Health Checks

The image includes health checks:
- **HTTP mode**: Checks if the HTTP endpoint is responding
- **Stdio mode**: Verifies the process is running

## Available Tools

This image provides 40+ MCP tools for:
- Database management (create, drop, list databases)
- Table operations (create, alter, drop, query tables)
- Version control (branches, commits, merges, diffs)
- Data operations (insert, update, delete, select)
- Remote operations (clone, fetch, push, pull)

## Support

- [GitHub Repository](https://github.com/dolthub/dolt-mcp)
- [Dolt Discord](https://discord.gg/gqr7K4VNKe)
- [Dolt Documentation](https://docs.dolthub.com/)

## License

This project follows the same license as the main Dolt project.
