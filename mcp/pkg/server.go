package pkg

import (
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/server"
)

const (
	DoltMCPServerName    = "dolt-mcp"
	DoltMCPServerVersion = "0.0.1"
)

type Server interface {
	MCP() *server.MCPServer
	DBConfig() db.Config
}

type Option func(Server)

