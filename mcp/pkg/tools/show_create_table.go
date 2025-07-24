package tools

import (
	"fmt"
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ShowCreateTableToolName                 = "show_create_table"
	ShowCreateTableToolSQLQueryFormatString = "SHOW CREATE TABLE `%s`;"
	ShowCreateTableToolDescription          = "Shows the schema of the specified table."
	ShowCreateTableTableArgumentDescription = "The name of the table."
)

func RegisterShowCreateTableTool(server pkg.Server) {
	mcpServer := server.MCP()

	showCreateTableTool := mcp.NewTool(
		ShowCreateTableToolName,
		mcp.WithDescription(ShowCreateTableToolDescription),
		mcp.WithString(
			TableCallToolArgumentName,
			mcp.Required(),
			mcp.Description(ShowCreateTableTableArgumentDescription),
		),
	)

	mcpServer.AddTool(showCreateTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		table, err := GetRequiredStringArgumentFromCallToolRequest(request, TableCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		result, err := database.QueryContext(ctx, fmt.Sprintf(ShowCreateTableToolSQLQueryFormatString, table), db.ResultFormatMarkdown)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(result), nil
	})
}

