package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
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

func NewDropTableTool() mcp.Tool {
	return mcp.NewTool(
		DropTableToolName,
		mcp.WithDescription(DropTableToolDescription),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString(
			WorkingDatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingDatabaseCallToolArgumentDescription),
		),
		mcp.WithString(
			WorkingBranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingBranchCallToolArgumentDescription),
		),
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
}

func RegisterDropTableTool(server pkg.Server) {
	mcpServer := server.MCP()
	dropTableTool := NewDropTableTool()

	mcpServer.AddTool(dropTableTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var workingBranch string
		workingBranch, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingBranchCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var workingDatabase string
		workingDatabase, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var tableToDrop string
		tableToDrop, err = GetRequiredStringArgumentFromCallToolRequest(request, TableCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		ifExists := GetBooleanArgumentFromCallToolRequest(request, IfExistsCallToolArgumentName)

		var query string
		if ifExists {
			query = fmt.Sprintf(DropTableIfExistsToolSQLQueryFormatString, tableToDrop)
		} else {
			query = fmt.Sprintf(DropTableToolSQLQueryFormatString, tableToDrop)
		}

		config := server.DBConfig()

		var tx db.DatabaseTransaction
		tx, err = NewDatabaseTransactionUsingDatabaseOnBranch(ctx, config, workingDatabase, workingBranch)
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

		result = mcp.NewToolResultText(fmt.Sprintf(DropTableToolCallSuccessFormatString, tableToDrop))
		return
	})
}
