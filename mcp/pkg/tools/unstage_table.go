package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	UnstageTableToolName                     = "unstage_table"
	UnstageTableToolTableArgumentDescription = "The name of the table to remove from the staging area."
	UnstageTableToolDescription              = "Removes a staged table from the staging area."
	UnstageTableToolSQLQueryFormatString     = "CALL DOLT_RESET('%s');"
	UnstageTableToolCallSuccessFormatString  = "successfully unstaged table: %s"
)

func RegisterUnstageTableTool(server pkg.Server) {
	mcpServer := server.MCP()

	unstageTableTool := mcp.NewTool(
		UnstageTableToolName,
		mcp.WithDescription(UnstageTableToolDescription),
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
			mcp.Description(UnstageTableToolTableArgumentDescription),
		),
	)

	mcpServer.AddTool(unstageTableTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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
		tx, err = NewDatabaseTransactionOnBranchUsingDatabase(ctx, config, workingBranch, workingDatabase)
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

		err = tx.ExecContext(ctx, fmt.Sprintf(UnstageTableToolSQLQueryFormatString, table))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(UnstageTableToolCallSuccessFormatString, table))
		return
	})
}

