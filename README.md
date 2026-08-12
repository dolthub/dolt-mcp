# Dolt MCP Server

A Model Context Protocol (MCP) server that provides AI assistants with direct access to Dolt and DoltgreSQL databases. This server enables AI tools like Claude to interact with Dolt's version-controlled SQL databases over either the MySQL or PostgreSQL wire protocol, allowing for database operations, version control workflows, and data management tasks.

## Overview

The Dolt MCP Server acts as a bridge between AI assistants and Dolt databases, exposing a comprehensive set of tools for:

- **Database Management**: Create, drop, and manage databases
- **Table Operations**: Create, alter, drop, describe, and query tables
- **Version Control**: Branch management, commits, merges, and diffs
- **Data Operations**: Insert, update, delete, and query data
- **Remote Operations**: Clone, fetch, push, and pull from remote repositories

Both [Dolt](https://github.com/dolthub/dolt) (MySQL-compatible) and [DoltgreSQL](https://github.com/dolthub/doltgresql) (PostgreSQL-compatible) backends are supported. The SQL dialect is selected at startup with the `--dolt` or `--doltgres` flag. An embedded [DoltLite](https://github.com/dolthub/doltlite) backend (`--doltlite`) is also available in specially built binaries, running against a local database file with no server at all — see [DoltLite Mode](#doltlite-mode-embedded).

## Installation

### Prerequisites

- Go 1.25 or later
- A running Dolt or DoltgreSQL SQL server instance (not needed for [DoltLite mode](#doltlite-mode-embedded))

### Building from Source

The default build is pure Go (no cgo) and cross-compiles freely:

```bash
git clone https://github.com/dolthub/dolt-mcp
cd dolt-mcp
go build -o dolt-mcp-server ./mcp/cmd/dolt-mcp-server
```

Binaries built this way do not include DoltLite support (`--doltlite` returns an error). Building with embedded DoltLite requires cgo and libdoltlite; see [Building with DoltLite Support](#building-with-doltlite-support).

### Docker Installation

Pull the official Docker image:

```bash
docker pull dolthub/dolt-mcp:latest
```

## Usage

The Dolt MCP Server can run in two modes and supports multiple deployment methods:

### Docker Usage (Recommended for Production)

#### HTTP Server with Docker

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

#### Stdio Server with Docker

```bash
docker run -it --rm \
  -e MCP_MODE=stdio \
  -e DOLT_HOST=your-dolt-host \
  -e DOLT_USER=root \
  -e DOLT_DATABASE=your_database \
  -e DOLT_PASSWORD=your_password \
  dolthub/dolt-mcp:latest
```

#### Connecting to DoltgreSQL from Docker

Set `MCP_DIALECT=doltgres` to point the container at a DoltgreSQL server. `DOLT_PORT` defaults to `5432` when the dialect is `doltgres`.

```bash
docker run -d \
  --name dolt-mcp-server \
  -p 8080:8080 \
  -e MCP_MODE=http \
  -e MCP_DIALECT=doltgres \
  -e DOLT_HOST=your-doltgres-host \
  -e DOLT_USER=postgres \
  -e DOLT_DATABASE=your_database \
  -e DOLT_PASSWORD=your_password \
  dolthub/dolt-mcp:latest
```

### Native Binary Usage

#### 1. Stdio Server (Recommended for AI Assistants)

The stdio server communicates over standard input/output, making it ideal for integration with AI assistants like Claude Desktop.

Against Dolt (MySQL dialect, the default):

```bash
./dolt-mcp-server \
  --stdio \
  --dolt \
  --host 0.0.0.0 \
  --port 3306 \
  --user root \
  --database mydb
```

Against DoltgreSQL (PostgreSQL dialect):

```bash
./dolt-mcp-server \
  --stdio \
  --doltgres \
  --host 0.0.0.0 \
  --port 5432 \
  --user postgres \
  --database mydb
```

If `--port` is omitted, it defaults to `3306` for Dolt and `5432` for DoltgreSQL.

#### Claude Desktop Configuration

Add this configuration to your Claude Desktop MCP settings:

```json
{
  "mcpServers": {
    "dolt-mcp": {
      "command": "/path/to/dolt-mcp-server",
      "args": [
        "--stdio",
        "--dolt",
        "--host", "0.0.0.0",
        "--port", "3306",
        "--user", "root",
        "--database", "your_database_name"
      ],
      "env": {
        "DOLT_PASSWORD": "your_password_if_needed"
      }
    }
  }
}
```

For a DoltgreSQL backend, swap `--dolt` for `--doltgres` and adjust the port/user to match your server.

#### HTTP Client Configuration

When connecting to a Dolt MCP server running in HTTP mode, you can configure Claude to use the HTTP transport. **Important**: HTTP connections require the `/mcp` endpoint to be appended to the server URL.

##### Using Claude CLI

```bash
claude mcp add --transport http dolt-mcp https://your-dolt-host:8080/mcp --header "Authorization: Bearer <token>"
```

##### Claude Desktop Configuration

Add this configuration to your Claude Desktop MCP settings:

```json
{
  "mcpServers": {
    "dolt-mcp": {
      "transport": "http",
      "url": "https://your-dolt-host:8080/mcp",
      "headers": {
        "Authorization": "Bearer <your-token>"
      }
    }
  }
}
```

**Note**: Replace `your-dolt-host`, the port, and `<your-token>` with your actual server details. The `/mcp` endpoint is required for HTTP connections.

#### 2. HTTP Server

The HTTP server exposes a REST API for MCP tool calls, useful for web applications and custom integrations.

```bash
./dolt-mcp-server \
  --http \
  --mcp-port 8080 \
  --dolt \
  --host 0.0.0.0 \
  --port 3306 \
  --user root \
  --database mydb
```

Pass `--doltgres` in place of `--dolt` to connect to a DoltgreSQL server.

## DoltLite Mode (Embedded)

[DoltLite](https://github.com/dolthub/doltlite) is a fork of SQLite that adds Dolt's version control features: an entire version-controlled database — branches, commits, diffs, merges, and remotes — lives in a single local file. In DoltLite mode the MCP server embeds the database engine directly, so there is no Dolt or DoltgreSQL server to install, configure, or run: point the server at a database file (created on first use if missing) and start working. This makes it ideal for local-first AI workflows on a laptop or in a container.

DoltLite mode requires a binary built with DoltLite support. Prebuilt archives named `dolt-mcp-server-doltlite-<platform>` are attached to [releases](https://github.com/dolthub/dolt-mcp/releases) for Linux and macOS on amd64/arm64, plus Windows x64, and the multi-architecture `dolthub/dolt-mcp:<version>-doltlite` Docker images include it. The default (pure Go) build does not: passing `--doltlite` to it errors with `this binary was built without DoltLite support; rebuild with -tags "doltlite libsqlite3" ...`.

### Quick Start with Docker

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

Mount a volume at `/data` so the database file persists across container restarts. See [docker/README.md](docker/README.md) for the full set of environment variables.

### Quick Start with a Native Binary

Download a `dolt-mcp-server-doltlite` archive from the [releases page](https://github.com/dolthub/dolt-mcp/releases) (or build it yourself, below), then:

```bash
./dolt-mcp-server-doltlite \
  --doltlite \
  --db-file /path/to/mydb.db \
  --commit-name "Your Name" \
  --commit-email you@example.com \
  --http --mcp-port 8080
# or --stdio in place of --http --mcp-port
```

No `--host`, `--port`, `--user`, `--password`, or TLS flags are needed — DoltLite runs in-process against the file.

### DoltLite Flags

- `--doltlite`: Use the embedded DoltLite dialect. Mutually exclusive with `--dolt` and `--doltgres`.
- `--db-file`: Path to the DoltLite database file (required with `--doltlite`). The file is created if it does not exist.
- `--commit-name` / `--commit-email`: The author name and email used for Dolt commits. Recommended — commits are authored as "doltlite" when unset.

### Claude Desktop Configuration (DoltLite)

```json
{
  "mcpServers": {
    "dolt-mcp-doltlite": {
      "command": "/path/to/dolt-mcp-server-doltlite",
      "args": [
        "--stdio",
        "--doltlite",
        "--db-file", "/path/to/mydb.db",
        "--commit-name", "Your Name",
        "--commit-email", "you@example.com"
      ]
    }
  }
}
```

### Building with DoltLite Support

DoltLite support requires cgo, the `doltlite` and `libsqlite3` build tags, and libdoltlite (with zlib and pthreads). Because of cgo, DoltLite binaries cannot be cross-compiled — build on the target platform.

Get libdoltlite either way:

1. **GitHub release zips**: download the `doltlite-lib-<platform>-<version>.zip` asset (e.g. `doltlite-lib-linux-x64-0.11.45.zip`) from a [dolthub/doltlite release](https://github.com/dolthub/doltlite/releases). It contains `doltlite.h` and `libdoltlite.a`. `mattn/go-sqlite3` built with the `libsqlite3` tag includes `<sqlite3.h>`, so copy the header: `cp doltlite.h sqlite3.h` inside the unpacked directory.
2. **Build from source**: clone [dolthub/doltlite](https://github.com/dolthub/doltlite) at the pinned tag, then `mkdir build && cd build && ../configure && make doltlite-lib` (requires a C toolchain, `tcl`, and zlib headers, e.g. `apt-get install build-essential tcl zlib1g-dev`). The build directory generates `sqlite3.h` natively.

Then build the server, pointing cgo at the directory containing the header and static library:

```bash
CGO_CFLAGS="-I/path/to/doltlite/build" \
CGO_LDFLAGS="/path/to/doltlite/build/libdoltlite.a -lz -lpthread" \
go build -tags "doltlite libsqlite3" -o dolt-mcp-server-doltlite ./mcp/cmd/dolt-mcp-server
```

On Linux, append `-lm -ldl` to `CGO_LDFLAGS`. The static link means the resulting binary has no runtime dependency on a doltlite shared library.

### Tools Unavailable in DoltLite Mode

DoltLite embeds a single database in a single file, so server- and multi-database-oriented tools are automatically hidden from clients in this mode:

- `list_databases`, `create_database`, `drop_database`, `clone_database`
- `show_processlist`, `kill_process`
- `get_dolt_merge_status`
- `run_dolt_tests`, `add_dolt_test`, `remove_dolt_test`

Everything else (35 tools) works, including remote operations against `file://` URLs and DoltLite-compatible HTTP(S) remotes. Authenticated remotes use DoltLite credentials; create one through the `exec` tool with `SELECT dolt_creds_new();`, then configure the returned key with the remote service. The engine reads credentials from `~/.doltlite/creds` by default or `DOLTLITE_CREDS_DIR` when set.

### Behavioral Notes

- **One database per file**: the `working_database` tool argument is accepted but ignored; there is only ever one database.
- **Commit author**: configure `--commit-name`/`--commit-email` or commits are authored as "doltlite".
- **Branch switching**: requires a clean working set — commit or reset changes first.
- **Concurrency**: the MCP server serializes its tool calls through one connection to preserve session state. Other DoltLite processes may read concurrently; file locking still permits only one writer at a time.
- **Remote compatibility**: remote storage must speak DoltLite's file or HTTP(S) protocol; a full Dolt repository and a DoltLite database use different storage formats.

## Configuration Options

### Required Parameters

- `--host`: Hostname of the Dolt or DoltgreSQL server (not used with `--doltlite`)
- `--user`: Username for server authentication (not used with `--doltlite`)
- `--stdio` or `--http`: Server mode selection

### Dialect Selection

- `--dolt`: Use the Dolt (MySQL-compatible) dialect. This is the default when no dialect flag is passed.
- `--doltgres`: Use the DoltgreSQL (PostgreSQL-compatible) dialect.
- `--doltlite`: Use the embedded DoltLite dialect. Requires `--db-file` and a binary built with DoltLite support (see [DoltLite Mode](#doltlite-mode-embedded)).

`--dolt`, `--doltgres`, and `--doltlite` are mutually exclusive.

### Optional Parameters

- `--database`: Name of the database to connect to
- `--port`: Server port. Defaults to `3306` for Dolt and `5432` for DoltgreSQL.
- `--password`: Password for authentication (can also use environment variable)
- `--tls`: TLS mode for the database connection: `true`, `false`, `skip-verify`, or `preferred`
- `--tls-ca`: Path to a CA certificate file for the database TLS connection
- `--mcp-port`: HTTP server port (default: 8080, HTTP mode only)
- `--db-file`: Path to the DoltLite database file, created if missing (required with `--doltlite`)
- `--commit-name`: Author name for Dolt commits (`--doltlite` only, recommended)
- `--commit-email`: Author email for Dolt commits (`--doltlite` only, recommended)

### Environment Variables

- `DOLT_PASSWORD`: Set the password for Dolt server authentication

### Docker Environment Variables

When using Docker, you can configure the server using environment variables:

#### Required
- `DOLT_HOST`: Hostname of the Dolt SQL server (not used with `doltlite`)
- `DOLT_USER`: Username for Dolt server authentication (not used with `doltlite`)

#### Optional
- `DOLT_DATABASE`: Name of the database to connect to
- `DOLT_PASSWORD`: Password for authentication
- `DOLT_PORT`: Server port (default: 3306 for `dolt`, 5432 for `doltgres`)
- `MCP_DIALECT`: SQL dialect: `dolt` (MySQL-compatible), `doltgres` (PostgreSQL-compatible), or `doltlite` (embedded, requires the `-doltlite` image variant). Default: `dolt` (`doltlite` in the `-doltlite` images)
- `MCP_MODE`: Server mode: `http` or `stdio` (default: stdio)
- `MCP_PORT`: HTTP server port (default: 8080, HTTP mode only)

#### DoltLite Only (`dolthub/dolt-mcp:<version>-doltlite` images)
- `DOLT_DB_FILE`: Path to the DoltLite database file inside the container (default: `/data/doltlite.db`); mount a volume at `/data` to persist it
- `DOLT_COMMIT_NAME`: Author name for Dolt commits (recommended)
- `DOLT_COMMIT_EMAIL`: Author email for Dolt commits (recommended)
- `DOLTLITE_CREDS_DIR`: Credential directory for authenticated HTTP(S) remotes (default: `/data/creds`)
- `DOLTLITE_CREDS_KID`: Optional credential key ID when more than one key is present
- `DOLTLITE_CA_FILE`: Optional CA bundle for a private HTTPS remote

### Docker Compose Example

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

## Available Tools

The Dolt MCP Server provides 40+ tools organized by functionality:

### Database Management
- `list_databases`: List all available databases
- `create_database`: Create a new database
- `drop_database`: Remove a database
- `select_version`: Get Dolt server version information

### Table Operations
- `show_tables`: List tables in current database
- `show_create_table`: Show table creation SQL
- `describe_table`: Show table schema and structure
- `create_table`: Create new tables
- `alter_table`: Modify table structure
- `drop_table`: Remove tables

### Data Operations
- `query`: Execute SELECT queries (read operations)
- `exec`: Execute INSERT, UPDATE, DELETE queries (write operations)

### Branch Management
- `list_dolt_branches`: List all branches
- `select_active_branch`: Show currently active branch
- `create_dolt_branch`: Create new branches
- `create_dolt_branch_from_head`: Create branch from current HEAD
- `delete_dolt_branch`: Remove branches
- `move_dolt_branch`: Rename branches

### Version Control
- `list_dolt_commits`: View commit history
- `create_dolt_commit`: Create commits with staged changes
- `stage_table_for_dolt_commit`: Stage specific tables
- `stage_all_tables_for_dolt_commit`: Stage all modified tables
- `unstage_table`: Remove tables from staging area
- `unstage_all_tables`: Clear staging area

### Diff and Status
- `list_dolt_diff_changes_in_working_set`: Show uncommitted changes
- `list_dolt_diff_changes_by_table_name`: Show changes for specific table
- `list_dolt_diff_changes_in_date_range`: Show changes within date range
- `get_dolt_merge_status`: Check merge conflicts and status

### Merge Operations
- `merge_dolt_branch`: Merge branches (fast-forward when possible)
- `merge_dolt_branch_no_fast_forward`: Force merge commit

### Reset Operations
- `dolt_reset_soft`: Soft reset to a revision (table, branch, commit, working set, or '.')
- `dolt_reset_hard`: Hard reset to a revision

### Remote Operations
- `list_dolt_remotes`: List configured remotes
- `add_dolt_remote`: Add new remote repositories
- `remove_dolt_remote`: Remove remote repositories
- `clone_database`: Clone remote databases
- `dolt_fetch_branch`: Fetch specific branch from remote
- `dolt_fetch_all_branches`: Fetch all branches from remote
- `dolt_push_branch`: Push branch to remote
- `dolt_pull_branch`: Pull branch from remote

## Example Workflows

### Basic Database Operations

```bash
# Start the MCP server
./dolt-mcp-server --stdio --dolt --host localhost --user root --database testdb

# Example AI interactions:
# "Show me all tables in the database"
# "Create a table called users with id, name, and email columns"  
# "Insert some sample data into the users table"
# "Show me the current branch and recent commits"
```

### Version Control Workflow

```bash
# Example AI workflow:
# "Create a new branch called 'feature-users'"
# "Switch to the feature-users branch" 
# "Create a users table with appropriate schema"
# "Stage and commit these changes"
# "Switch back to main and merge the feature branch"
```

### Data Analysis

```bash
# Example AI interactions:
# "Show me all data in the sales table"
# "Calculate total revenue by month from the orders table"
# "Show me what changed in the products table in the last week"
# "Create a branch to experiment with data transformations"
```

## Development

### Running Tests

```bash
go test ./...
```

### Integration Tests

The repository includes comprehensive integration tests that validate tool functionality against a real Dolt server instance.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project follows the same license as the main Dolt project.

## Support

For issues and questions:
- Create issues in this repository
- Join the [Dolt Discord](https://discord.gg/gqr7K4VNKe) community
- Check the [Dolt documentation](https://docs.dolthub.com/)
