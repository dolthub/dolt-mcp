package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DoltResetTableSoftToolName                     = "dolt_reset_table_soft"
	DoltResetTableSoftToolTableArgumentDescription = "The name of the table to soft reset."
	DoltResetTableSoftToolDescription              = "Soft resets the specified table."
	DoltResetTableSoftToolSQLQueryFormatString     = "CALL DOLT_RESET('%s');"
	DoltResetTableSoftToolCallSuccessFormatString  = "successfully soft reset table: %s"
)

func NewDoltResetTableSoftTool() mcp.Tool {
    return mcp.NewTool(
        DoltResetTableSoftToolName,
        mcp.WithDescription(DoltResetTableSoftToolDescription),
        mcp.WithReadOnlyHintAnnotation(false),
        mcp.WithDestructiveHintAnnotation(false),
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
            mcp.Description(DoltResetTableSoftToolTableArgumentDescription),
        ),
    )
}

func RegisterDoltResetTableSoftTool(server pkg.Server) {
    mcpServer := server.MCP()
    resetTableSoftTool := NewDoltResetTableSoftTool()

	mcpServer.AddTool(resetTableSoftTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var table string
		table, err = GetRequiredStringArgumentFromCallToolRequest(request, TableCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
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

		err = tx.ExecContext(ctx, fmt.Sprintf(DoltResetTableSoftToolSQLQueryFormatString, table))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(DoltResetTableSoftToolCallSuccessFormatString, table))
		return
	})
}
