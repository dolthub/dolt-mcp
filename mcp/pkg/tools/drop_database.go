package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
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

func NewDropDatabaseTool() mcp.Tool {
    return mcp.NewTool(
        DropDatabaseToolName,
        mcp.WithDescription(DropDatabaseToolDescription),
        mcp.WithReadOnlyHintAnnotation(false),
        mcp.WithDestructiveHintAnnotation(true),
        mcp.WithIdempotentHintAnnotation(true),
        mcp.WithOpenWorldHintAnnotation(false),
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
}

func RegisterDropDatabaseTool(server pkg.Server) {
    mcpServer := server.MCP()
    dropDatabaseTool := NewDropDatabaseTool()

	mcpServer.AddTool(dropDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var databaseToDrop string
		databaseToDrop, err = GetRequiredStringArgumentFromCallToolRequest(request, DatabaseCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		ifExists := GetBooleanArgumentFromCallToolRequest(request, IfExistsCallToolArgumentName)

		var query string
		if ifExists {
			query = fmt.Sprintf(DropDatabaseIfExistsToolSQLQueryFormatString, databaseToDrop)
		} else {
			query = fmt.Sprintf(DropDatabaseToolSQLQueryFormatString, databaseToDrop)
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

		result = mcp.NewToolResultText(fmt.Sprintf(DropDatabaseToolCallSuccessFormatString, databaseToDrop))
		return
	})
}

