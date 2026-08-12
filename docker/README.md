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

## DoltLite Variant (Embedded, No Server Required)

The `-doltlite` image tags (`dolthub/dolt-mcp:latest-doltlite`, `dolthub/dolt-mcp:<version>-doltlite`) ship a build of the server with embedded [DoltLite](https://github.com/dolthub/doltlite) support: a version-controlled database stored in a single local file, with no Dolt or DoltgreSQL server needed. Mount a volume at `/data` so the database file persists across container restarts.

```bash
docker run -d \
  --name dolt-mcp-server-doltlite \
  -p 8080:8080 \
  -v dolt_mcp_data:/data \
  -e MCP_MODE=http \
  -e DOLT_DB_FILE=/data/mydb.db \
  -e DOLT_COMMIT_NAME="Your Name" \
  -e DOLT_COMMIT_EMAIL=you@example.com \
  dolthub/dolt-mcp:latest-doltlite
```

Stdio mode:

```bash
docker run -it --rm \
  -v dolt_mcp_data:/data \
  -e MCP_MODE=stdio \
  -e DOLT_DB_FILE=/data/mydb.db \
  -e DOLT_COMMIT_NAME="Your Name" \
  -e DOLT_COMMIT_EMAIL=you@example.com \
  dolthub/dolt-mcp:latest-doltlite
```

The DoltLite image defaults to `MCP_DIALECT=doltlite`, `DOLT_DB_FILE=/data/doltlite.db`, and `DOLTLITE_CREDS_DIR=/data/creds` (the database and credential directory are created on first use). `DOLT_HOST`, `DOLT_USER`, `DOLT_PORT`, and `DOLT_PASSWORD` are not used. Setting `DOLT_COMMIT_NAME`/`DOLT_COMMIT_EMAIL` is recommended; otherwise Dolt commits are authored as "doltlite". The server serializes tool calls through one connection to preserve session state; other DoltLite readers can still open the file, while file locking permits one writer at a time. A handful of server-oriented tools (database management, process management, and Dolt test tools) are hidden in this mode. The DoltLite image is published for `linux/amd64` and `linux/arm64`.

## Environment Variables

### Required
- `DOLT_HOST` - Hostname of the Dolt SQL server (not used with `doltlite`)
- `DOLT_USER` - Username for Dolt server authentication (not used with `doltlite`)
- `DOLT_DATABASE` - Name of the database to connect to (not used with `doltlite`)

### Optional
- `DOLT_PASSWORD` - Password for authentication (recommended to use Docker secrets in production)
- `DOLT_PORT` - Server port (default: 3306 for `dolt`, 5432 for `doltgres`)
- `MCP_DIALECT` - SQL dialect: `dolt` (MySQL-compatible), `doltgres` (PostgreSQL-compatible), or `doltlite` (embedded, requires the `-doltlite` image variant). Default: `dolt` (`doltlite` in the `-doltlite` images)
- `MCP_MODE` - Server mode: `http` or `stdio` (default: stdio)
- `MCP_PORT` - HTTP server port (default: 8080, HTTP mode only)

### DoltLite Only
- `DOLT_DB_FILE` - Path to the DoltLite database file inside the container (required with `doltlite`; default in the `-doltlite` images: `/data/doltlite.db`). The file is created if it does not exist.
- `DOLT_COMMIT_NAME` - Author name for Dolt commits (recommended)
- `DOLT_COMMIT_EMAIL` - Author email for Dolt commits (recommended)
- `DOLTLITE_CREDS_DIR` - Credential directory used by authenticated HTTP(S) remotes (default in the DoltLite image: `/data/creds`)
- `DOLTLITE_CREDS_KID` - Optional credential key ID when the directory contains more than one key
- `DOLTLITE_CA_FILE` - Optional CA bundle for a private HTTPS remote

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

Or with the DoltLite variant (no database server required):

```json
{
  "mcpServers": {
    "dolt-mcp-doltlite": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-v", "dolt_mcp_data:/data",
        "-e", "MCP_MODE=stdio",
        "-e", "DOLT_DB_FILE=/data/mydb.db",
        "-e", "DOLT_COMMIT_NAME=Your Name",
        "-e", "DOLT_COMMIT_EMAIL=you@example.com",
        "dolthub/dolt-mcp:latest-doltlite"
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
