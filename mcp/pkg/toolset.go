package pkg

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
)

const (
	ListDatabasesToolName = "list_databases"
	ListDatabasesSQLQuery = "SHOW DATABASES;"
)

type ToolSet interface {
	RegisterTools(server Server)
}

type PrimitiveToolSetV1 struct{}

func (v *PrimitiveToolSetV1) registerListDatabasesTool(server Server) {
	mcpServer := server.MCP()

	listDatabasesTool := mcp.NewTool(ListDatabasesToolName, mcp.WithDescription("List all databases in the Dolt server"))
	mcpServer.AddTool(listDatabasesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		database := server.DB()
		result, err := database.QueryContext(ctx, ListDatabasesSQLQuery, db.ResultFormatMarkdown)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(result), nil
	})
}

func (v *PrimitiveToolSetV1) RegisterTools(server Server) {
	v.registerListDatabasesTool(server)
}

