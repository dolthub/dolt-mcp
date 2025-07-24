package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DropTableToolName                         = "drop_table"
	DropTableToolTableArgumentDescription     = "The name of the table to drop."
	DropTableToolSQLQueryFormatString         = "DROP TABLE `%s`;"
	DropTableIfExistsToolSQLQueryFormatString = "DROP TABLE IF EXISTS `%s`;"
	DropTableToolDescription                  = "Drops the specified table."
	DropTableToolCallSuccessFormatString      = "successfully dropped table: %s"
	DropTableToolIfExistsArgumentDescription  = "If true will only drop the specified table if it exists."
)

func RegisterDropTableTool(server pkg.Server) {
	mcpServer := server.MCP()

	dropTableTool := mcp.NewTool(
		DropTableToolName,
		mcp.WithDescription(DropTableToolDescription),
		mcp.WithString(
			TableCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DropTableToolTableArgumentDescription),
		),
		mcp.WithBoolean(
			IfExistsCallToolArgumentName,
			mcp.Description(DropTableToolIfExistsArgumentDescription),
		),
	)

	mcpServer.AddTool(dropTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tableToDrop, err := GetRequiredStringArgumentFromCallToolRequest(request, TableCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		ifExists := GetBooleanArgumentFromCallToolRequest(request, IfExistsCallToolArgumentName)

		var query string
		if ifExists {
			query = fmt.Sprintf(DropTableIfExistsToolSQLQueryFormatString, tableToDrop)
		} else {
			query = fmt.Sprintf(DropTableToolSQLQueryFormatString, tableToDrop)
		}

		database := server.DB()
		err = database.ExecContext(ctx, query)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(DropTableToolCallSuccessFormatString, tableToDrop)), nil
	})
}

