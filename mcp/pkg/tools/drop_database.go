package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DropDatabaseToolName                         = "drop_database"
	DropDatabaseToolDatabaseArgumentDescription  = "The name of the database to drop."
	DropDatabaseToolSQLQueryFormatString         = "DROP DATABASE `%s`;"
	DropDatabaseIfExistsToolSQLQueryFormatString = "DROP DATABASE IF EXISTS `%s`;"
	DropDatabaseToolDescription                  = "Drops a database in the Dolt server."
	DropDatabaseToolCallSuccessFormatString      = "successfully dropped database: %s"
	DropDatabaseToolIfExistsArgumentDescription  = "If true will only drop the specified database if it exists in the Dolt server."
)

func RegisterDropDatabaseTool(server pkg.Server) {
	mcpServer := server.MCP()

	dropDatabaseTool := mcp.NewTool(
		DropDatabaseToolName,
		mcp.WithDescription(DropDatabaseToolDescription),
		mcp.WithString(
			DatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DropDatabaseToolDatabaseArgumentDescription),
		),
		mcp.WithBoolean(
			IfExistsCallToolArgumentName,
			mcp.Description(DropDatabaseToolIfExistsArgumentDescription),
		),
	)

	mcpServer.AddTool(dropDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		databaseToDrop, err := GetRequiredStringArgumentFromCallToolRequest(request, DatabaseCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		ifExists := GetBooleanArgumentFromCallToolRequest(request, IfExistsCallToolArgumentName)

		var query string
		if ifExists {
			query = fmt.Sprintf(DropDatabaseIfExistsToolSQLQueryFormatString, databaseToDrop)
		} else {
			query = fmt.Sprintf(DropDatabaseToolSQLQueryFormatString, databaseToDrop)
		}

		database := server.DB()
		err = database.ExecContext(ctx, query)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(DropDatabaseToolCallSuccessFormatString, databaseToDrop)), nil
	})
}
