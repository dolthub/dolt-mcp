package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateTableToolName                     = "create_table"
	CreateTableToolQueryArgumentDescription = "The CREATE TABLE statement to run."
	CreateTableToolDescription              = "Creates a table."
	CreateTableToolCallSuccessMessage       = "successfully created table"
)

func RegisterCreateTableTool(server pkg.Server) {
	mcpServer := server.MCP()

	createTableTool := mcp.NewTool(
		CreateTableToolName,
		mcp.WithDescription(CreateTableToolDescription),
		mcp.WithString(
			QueryCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CreateTableToolQueryArgumentDescription),
		),
	)

	mcpServer.AddTool(createTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		createTableStatement, err := GetRequiredStringArgumentFromCallToolRequest(request, QueryCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		err = database.ExecContext(ctx, createTableStatement)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(CreateTableToolCallSuccessMessage), nil
	})
}

