package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	UseDatabaseToolName                        = "use_database"
	UseDatabaseToolDatabaseArgumentDescription = "The name of the database to use."
	UseDatabaseToolSQLQueryFormatString        = "use `%s`;"
	UseDatabaseToolDescription                 = "Specifies which database in the Dolt server to use."
	UseDatabaseToolCallSuccessFormatString     = "now using database: %s"
)

func RegisterUseDatabaseTool(server pkg.Server) {
	mcpServer := server.MCP()

	useDatabaseTool := mcp.NewTool(
		UseDatabaseToolName,
		mcp.WithDescription(UseDatabaseToolDescription),
		mcp.WithString(
			DatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(UseDatabaseToolDatabaseArgumentDescription),
		),
	)

	mcpServer.AddTool(useDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var databaseToUse string
		databaseToUse, err = GetRequiredStringArgumentFromCallToolRequest(request, DatabaseCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
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

		err = tx.ExecContext(ctx, fmt.Sprintf(UseDatabaseToolSQLQueryFormatString, databaseToUse))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(UseDatabaseToolCallSuccessFormatString, databaseToUse))
		return
	})
}

