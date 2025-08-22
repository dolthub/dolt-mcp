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

# Add required parameters
CMD_ARGS="$CMD_ARGS --dolt-host $DOLT_HOST"
CMD_ARGS="$CMD_ARGS --dolt-port $DOLT_PORT"
CMD_ARGS="$CMD_ARGS --dolt-user $DOLT_USER"

# Add password if provided
if [ -n "$DOLT_PASSWORD" ]; then
    CMD_ARGS="$CMD_ARGS --dolt-password $DOLT_PASSWORD"
fi

# Add database if provided
if [ -n "$DOLT_DATABASE" ]; then
    CMD_ARGS="$CMD_ARGS --dolt-database $DOLT_DATABASE"
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
