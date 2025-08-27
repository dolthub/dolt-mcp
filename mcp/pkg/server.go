package pkg

import (
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/server"
)

const (
	DoltMCPServerName = "dolt-mcp"
)

var DoltMCPServerVersion = "0.2.0"

type Server interface {
	MCP() *server.MCPServer
	DBConfig() db.Config
}

type Option func(Server)
