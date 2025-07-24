package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	AlterTableToolName                     = "alter_table"
	AlterTableToolQueryArgumentDescription = "The ALTER TABLE statement to run."
	AlterTableToolDescription              = "Alters a table."
	AlterTableToolCallSuccessMessage       = "successfully altered table"
)

func RegisterAlterTableTool(server pkg.Server) {
	mcpServer := server.MCP()

	alterTableTool := mcp.NewTool(
		AlterTableToolName,
		mcp.WithDescription(AlterTableToolDescription),
		mcp.WithString(
			QueryCallToolArgumentName,
			mcp.Required(),
			mcp.Description(AlterTableToolQueryArgumentDescription),
		),
	)

	mcpServer.AddTool(alterTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		alterTableStatement, err := GetRequiredStringArgumentFromCallToolRequest(request, QueryCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		err = database.ExecContext(ctx, alterTableStatement)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(AlterTableToolCallSuccessMessage), nil
	})
}

