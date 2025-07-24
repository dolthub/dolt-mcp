package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateDatabaseToolName                            = "create_database"
	CreateDatabaseToolDatabaseArgumentDescription     = "The name of the database to create."
	CreateDatabaseToolSQLQueryFormatString            = "CREATE DATABASE %s;"
	CreateDatabaseIfNotExistsToolSQLQueryFormatString = "CREATE DATABASE IF NOT EXISTS %s;"
	CreateDatabaseToolDescription                     = "Creates a database in the Dolt server."
	CreateDatabaseToolCallSuccessFormatString         = "successfully created database: %s"
	CreateDatabaseToolIfNotExistsArgumentDescription  = "If true will only create the specified database if it does not exist in the Dolt server."
)

func RegisterCreateDatabaseTool(server pkg.Server) {
	mcpServer := server.MCP()

	createDatabaseTool := mcp.NewTool(
		CreateDatabaseToolName,
		mcp.WithDescription(CreateDatabaseToolDescription),
		mcp.WithString(
			DatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CreateDatabaseToolDatabaseArgumentDescription),
		),
		mcp.WithBoolean(
			IfNotExistsCallToolArgumentName,
			mcp.Description(CreateDatabaseToolIfNotExistsArgumentDescription),
		),
	)

	mcpServer.AddTool(createDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		databaseToCreate, err := GetRequiredStringArgumentFromCallToolRequest(request, DatabaseCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		ifNotExists := GetBooleanArgumentFromCallToolRequest(request, IfNotExistsCallToolArgumentName)

		var query string
		if ifNotExists {
			query = fmt.Sprintf(CreateDatabaseIfNotExistsToolSQLQueryFormatString, databaseToCreate)
		} else {
			query = fmt.Sprintf(CreateDatabaseToolSQLQueryFormatString, databaseToCreate)
		}

		database := server.DB()
		err = database.ExecContext(ctx, query)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(CreateDatabaseToolCallSuccessFormatString, databaseToCreate)), nil
	})
}

