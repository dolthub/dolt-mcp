# Dolt MCP Server

A Model Context Protocol (MCP) server that provides AI assistants with direct access to Dolt databases. This server enables AI tools like Claude to interact with Dolt's version-controlled SQL databases, allowing for database operations, version control workflows, and data management tasks.

## Overview

The Dolt MCP Server acts as a bridge between AI assistants and Dolt databases, exposing a comprehensive set of tools for:

- **Database Management**: Create, drop, and manage databases
- **Table Operations**: Create, alter, drop, describe, and query tables
- **Version Control**: Branch management, commits, merges, and diffs
- **Data Operations**: Insert, update, delete, and query data
- **Remote Operations**: Clone, fetch, push, and pull from remote repositories

## Installation

### Prerequisites

- Go 1.24.4 or later
- A running Dolt SQL server instance

### Building from Source

```bash
git clone https://github.com/dolthub/dolt-mcp
cd dolt-mcp
go build -o dolt-mcp-server ./mcp/cmd/dolt-mcp-server
```

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

### Native Binary Usage

#### 1. Stdio Server (Recommended for AI Assistants)

The stdio server communicates over standard input/output, making it ideal for integration with AI assistants like Claude Desktop.

```bash
./dolt-mcp-server \
  --stdio \
  --dolt-host 0.0.0.0 \
  --dolt-port 3306 \
  --dolt-user root \
  --dolt-database mydb
```

#### Claude Desktop Configuration

Add this configuration to your Claude Desktop MCP settings:

```json
{
  "mcpServers": {
    "dolt-mcp": {
      "command": "/path/to/dolt-mcp-server",
      "args": [
        "--stdio",
        "--dolt-host", "0.0.0.0",
        "--dolt-port", "3306", 
        "--dolt-user", "root",
        "--dolt-database", "your_database_name"
      ],
      "env": {
        "DOLT_PASSWORD": "your_password_if_needed"
      }
    }
  }
}
```

#### 2. HTTP Server

The HTTP server exposes a REST API for MCP tool calls, useful for web applications and custom integrations.

```bash
./dolt-mcp-server \
  --http \
  --mcp-port 8080 \
  --dolt-host 0.0.0.0 \
  --dolt-port 3306 \
  --dolt-user root \
  --dolt-database mydb
```

## Configuration Options

### Required Parameters

- `--dolt-host`: Hostname of the Dolt SQL server
- `--dolt-user`: Username for Dolt server authentication  
- `--stdio` or `--http`: Server mode selection

### Optional Parameters

- `--dolt-database`: Name of the database to connect to
- `--dolt-port`: Dolt server port (default: 3306)
- `--dolt-password`: Password for authentication (can also use environment variable)
- `--mcp-port`: HTTP server port (default: 8080, HTTP mode only)

### Environment Variables

- `DOLT_PASSWORD`: Set the password for Dolt server authentication

### Docker Environment Variables

When using Docker, you can configure the server using environment variables:

#### Required
- `DOLT_HOST`: Hostname of the Dolt SQL server
- `DOLT_USER`: Username for Dolt server authentication

#### Optional
- `DOLT_DATABASE`: Name of the database to connect to
- `DOLT_PASSWORD`: Password for authentication
- `DOLT_PORT`: Dolt server port (default: 3306)
- `MCP_MODE`: Server mode: `http` or `stdio` (default: stdio)
- `MCP_PORT`: HTTP server port (default: 8080, HTTP mode only)

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
- `dolt_reset_table_soft`: Soft reset specific table
- `dolt_reset_all_tables_soft`: Soft reset all tables
- `dolt_reset_hard`: Hard reset to specific commit

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
./dolt-mcp-server --stdio --dolt-host localhost --dolt-user root --dolt-database testdb

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

