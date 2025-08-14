package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateDatabaseToolName                            = "create_database"
	CreateDatabaseToolDatabaseArgumentDescription     = "The name of the database to create."
	CreateDatabaseToolSQLQueryFormatString            = "CREATE DATABASE `%s`;"
	CreateDatabaseIfNotExistsToolSQLQueryFormatString = "CREATE DATABASE IF NOT EXISTS `%s`;"
	CreateDatabaseToolDescription                     = "Creates a database in the Dolt server."
	CreateDatabaseToolCallSuccessFormatString         = "successfully created database: %s"
	CreateDatabaseToolIfNotExistsArgumentDescription  = "If true will only create the specified database if it does not exist in the Dolt server."
)

func NewCreateDatabaseTool() mcp.Tool {
    return mcp.NewTool(
        CreateDatabaseToolName,
        mcp.WithDescription(CreateDatabaseToolDescription),
        mcp.WithReadOnlyHintAnnotation(false),
        mcp.WithDestructiveHintAnnotation(true),
        mcp.WithIdempotentHintAnnotation(true),
        mcp.WithOpenWorldHintAnnotation(false),
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
}

func RegisterCreateDatabaseTool(server pkg.Server) {
    mcpServer := server.MCP()
    createDatabaseTool := NewCreateDatabaseTool()

	mcpServer.AddTool(createDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var databaseToCreate string
		databaseToCreate, err = GetRequiredStringArgumentFromCallToolRequest(request, DatabaseCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		ifNotExists := GetBooleanArgumentFromCallToolRequest(request, IfNotExistsCallToolArgumentName)

		var query string
		if ifNotExists {
			query = fmt.Sprintf(CreateDatabaseIfNotExistsToolSQLQueryFormatString, databaseToCreate)
		} else {
			query = fmt.Sprintf(CreateDatabaseToolSQLQueryFormatString, databaseToCreate)
		}

		config := server.DBConfig()
		var tx db.DatabaseTransaction
		tx, err = db.NewDatabaseTransaction(ctx, config)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		defer func() {
			rerr := CommitTransactionOrRollbackOnError(ctx, tx, err)
			if rerr != nil {
				result = mcp.NewToolResultError(rerr.Error())
			}
		}()

		err = tx.ExecContext(ctx, query)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(CreateDatabaseToolCallSuccessFormatString, databaseToCreate))
		return
	})
}
