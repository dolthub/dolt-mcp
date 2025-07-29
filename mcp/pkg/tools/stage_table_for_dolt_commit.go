package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	StageTableForDoltCommitToolName                     = "stage_table_for_dolt_commit"
	StageTableForDoltCommitToolTableArgumentDescription = "The name of the table to stage."
	StageTableForDoltCommitToolDescription              = "Stages a table for a Dolt commit."
	StageTableForDoltCommitToolSQLQueryFormatString     = "CALL DOLT_ADD('%s');"
	StageTableForDoltCommitToolCallSuccessFormatString  = "successfully staged table: %s"
)

func RegisterStageTableForDoltCommitTool(server pkg.Server) {
	mcpServer := server.MCP()

	stageTableForDoltCommitTool := mcp.NewTool(
		StageTableForDoltCommitToolName,
		mcp.WithDescription(StageTableForDoltCommitToolDescription),
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
			mcp.Description(StageTableForDoltCommitToolTableArgumentDescription),
		),
	)

	mcpServer.AddTool(stageTableForDoltCommitTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		err = tx.ExecContext(ctx, fmt.Sprintf(StageTableForDoltCommitToolSQLQueryFormatString, table))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(StageTableForDoltCommitToolCallSuccessFormatString, table))
		return
	})
}

