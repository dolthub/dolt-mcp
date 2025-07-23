package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateDatabaseToolName                            = "create_database"
	CreateDatabaseIfNotExistsToolName                 = "create_database_if_not_exists"
	CreateDatabaseToolDatabaseArgumentDescription     = "The name of the database to create."
	CreateDatabaseToolSQLQueryFormatString            = "CREATE DATABASE %s;"
	CreateDatabaseIfNotExistsToolSQLQueryFormatString = "CREATE DATABASE IF NOT EXISTS %s;"
	CreateDatabaseToolDescription                     = "Creates a database in the Dolt server."
	CreateDatabaseIfNotExistsToolDescription          = "Creates a database in the Dolt server, if the database does not already exist."
	CreateDatabaseToolCallSuccessFormatString         = "successfully created database: %s"
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
	)

	mcpServer.AddTool(createDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		databaseToCreate, err := GetRequiredStringArgumentFromCallToolRequest(request, DatabaseCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		err = database.ExecContext(ctx, fmt.Sprintf(CreateDatabaseToolSQLQueryFormatString, databaseToCreate))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(CreateDatabaseToolCallSuccessFormatString, databaseToCreate)), nil
	})
}

func RegisterCreateDatabaseIfNotExistsTool(server pkg.Server) {
	mcpServer := server.MCP()

	createDatabaseIfNotExistsTool := mcp.NewTool(
		CreateDatabaseIfNotExistsToolName,
		mcp.WithDescription(CreateDatabaseIfNotExistsToolDescription),
		mcp.WithString(
			DatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CreateDatabaseToolDatabaseArgumentDescription),
		),
	)

	mcpServer.AddTool(createDatabaseIfNotExistsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		databaseToCreate, err := GetRequiredStringArgumentFromCallToolRequest(request, DatabaseCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		err = database.ExecContext(ctx, fmt.Sprintf(CreateDatabaseIfNotExistsToolSQLQueryFormatString, databaseToCreate))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(CreateDatabaseToolCallSuccessFormatString, databaseToCreate)), nil
	})
}

