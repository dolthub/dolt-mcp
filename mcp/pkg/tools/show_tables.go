package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ShowTablesToolName     = "show_tables"
	ShowTablesToolSQLQuery = "SHOW TABLES;"
	ShowTablesToolDescription = "Show tables in the current database."
)

func RegisterShowTablesTool(server pkg.Server) {
	mcpServer := server.MCP()

	showTablesTool := mcp.NewTool(ShowTablesToolName, mcp.WithDescription(ShowTablesToolDescription))
	mcpServer.AddTool(showTablesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		database := server.DB()
		result, err := database.QueryContext(ctx, ShowTablesToolSQLQuery, db.ResultFormatMarkdown)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(result), nil
	})
}

