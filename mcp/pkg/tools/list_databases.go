package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListDatabasesToolName     = "list_databases"
	ListDatabasesToolSQLQuery = "SHOW DATABASES;"
	ListDatabasesToolDescription = "Lists all databases in the Dolt server."
)

func RegisterListDatabasesTool(server pkg.Server) {
	mcpServer := server.MCP()

	listDatabasesTool := mcp.NewTool(ListDatabasesToolName, mcp.WithDescription(ListDatabasesToolDescription))
	mcpServer.AddTool(listDatabasesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		database := server.DB()
		result, err := database.QueryContext(ctx, ListDatabasesToolSQLQuery, db.ResultFormatMarkdown)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(result), nil
	})
}

