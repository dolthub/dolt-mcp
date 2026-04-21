#!/bin/sh

set -e

# If arguments are provided, pass them directly to the binary
if [ $# -gt 0 ]; then
    exec ./dolt-mcp-server "$@"
fi

# Build command based on environment variables
CMD_ARGS=""

# Required parameters
if [ -z "$DOLT_HOST" ]; then
    echo "Error: DOLT_HOST environment variable is required"
    exit 1
fi

if [ -z "$DOLT_USER" ]; then
    echo "Error: DOLT_USER environment variable is required"
    exit 1
fi

# Determine SQL dialect (default: dolt)
MCP_DIALECT="${MCP_DIALECT:-dolt}"
case "$MCP_DIALECT" in
    dolt)
        CMD_ARGS="$CMD_ARGS --dolt"
        DEFAULT_PORT=3306
        ;;
    doltgres)
        CMD_ARGS="$CMD_ARGS --doltgres"
        DEFAULT_PORT=5432
        ;;
    *)
        echo "Error: MCP_DIALECT must be either 'dolt' or 'doltgres' (got: $MCP_DIALECT)"
        exit 1
        ;;
esac

# Default port based on dialect if not explicitly provided
if [ -z "$DOLT_PORT" ]; then
    DOLT_PORT="$DEFAULT_PORT"
fi

# Add required parameters
CMD_ARGS="$CMD_ARGS --host $DOLT_HOST"
CMD_ARGS="$CMD_ARGS --port $DOLT_PORT"
CMD_ARGS="$CMD_ARGS --user $DOLT_USER"

# Add password if provided
if [ -n "$DOLT_PASSWORD" ]; then
    CMD_ARGS="$CMD_ARGS --password $DOLT_PASSWORD"
fi

# Add database if provided
if [ -n "$DOLT_DATABASE" ]; then
    CMD_ARGS="$CMD_ARGS --database $DOLT_DATABASE"
fi

# Determine server mode
if [ "$MCP_MODE" = "http" ]; then
    CMD_ARGS="$CMD_ARGS --http"
    if [ -n "$MCP_PORT" ]; then
        CMD_ARGS="$CMD_ARGS --mcp-port $MCP_PORT"
    fi
    echo "Starting Dolt MCP Server in HTTP mode on port $MCP_PORT"
elif [ "$MCP_MODE" = "stdio" ]; then
    CMD_ARGS="$CMD_ARGS --stdio"
    echo "Starting Dolt MCP Server in stdio mode"
else
    echo "Error: MCP_MODE must be either 'http' or 'stdio'"
    exit 1
fi

echo "Connecting to Dolt server at $DOLT_HOST:$DOLT_PORT"

if [ -n "$DOLT_DATABASE" ]; then
    echo "Database: $DOLT_DATABASE"
fi

echo "User: $DOLT_USER"

# Execute the command
exec ./dolt-mcp-server $CMD_ARGS
