package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DropDatabaseToolName                         = "drop_database"
	DropDatabaseIfExistsToolName                 = "create_database_if_exists"
	DropDatabaseToolDatabaseArgumentDescription  = "The name of the database to drop."
	DropDatabaseToolSQLQueryFormatString         = "DROP DATABASE %s;"
	DropDatabaseIfExistsToolSQLQueryFormatString = "DROP DATABASE IF EXISTS %s;"
	DropDatabaseToolDescription                  = "Drops a database in the Dolt server."
	DropDatabaseIfExistsToolDescription          = "Drops a database in the Dolt server, if the database exists."
	DropDatabaseToolCallSuccessFormatString      = "successfully dropped database: %s"
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
	)

	mcpServer.AddTool(dropDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		databaseToDrop, err := GetRequiredStringArgumentFromCallToolRequest(request, DatabaseCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		err = database.ExecContext(ctx, fmt.Sprintf(DropDatabaseToolSQLQueryFormatString, databaseToDrop))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(DropDatabaseToolCallSuccessFormatString, databaseToDrop)), nil
	})
}

func RegisterDropDatabaseIfExistsTool(server pkg.Server) {
	mcpServer := server.MCP()

	dropDatabaseIfExistsTool := mcp.NewTool(
		DropDatabaseIfExistsToolName,
		mcp.WithDescription(DropDatabaseIfExistsToolDescription),
		mcp.WithString(
			DatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DropDatabaseToolDatabaseArgumentDescription),
		),
	)

	mcpServer.AddTool(dropDatabaseIfExistsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		databaseToDrop, err := GetRequiredStringArgumentFromCallToolRequest(request, DatabaseCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		err = database.ExecContext(ctx, fmt.Sprintf(DropDatabaseIfExistsToolSQLQueryFormatString, databaseToDrop))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(DropDatabaseToolCallSuccessFormatString, databaseToDrop)), nil
	})
}

