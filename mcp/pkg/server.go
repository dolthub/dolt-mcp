package pkg

import (
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/server"
)

const (
	DoltMCPServerName    = "dolt-mcp"
	DoltMCPServerVersion = "0.1.0"
)

type Server interface {
	MCP() *server.MCPServer
	DB() db.Database
}

type Option func(Server)

