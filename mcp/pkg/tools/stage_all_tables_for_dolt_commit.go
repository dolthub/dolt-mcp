package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	StageAllTablesForDoltCommitToolName               = "stage_all_tables_for_dolt_commit"
	StageAllTablesForDoltCommitToolDescription        = "Stages a table for a Dolt commit."
	StageAllTablesForDoltCommitToolSQLQuery           = "CALL DOLT_ADD('-A');"
	StageAllTablesForDoltCommitToolCallSuccessMessage = "successfully staged tables"
)

func NewStageAllTablesForDoltCommitTool() mcp.Tool {
	return mcp.NewTool(
		StageAllTablesForDoltCommitToolName,
		mcp.WithDescription(StageAllTablesForDoltCommitToolDescription),
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
	)
}

func RegisterStageAllTablesForDoltCommitTool(server pkg.Server) {
	mcpServer := server.MCP()
	stageAllTablesForDoltCommitTool := NewStageAllTablesForDoltCommitTool()

	mcpServer.AddTool(stageAllTablesForDoltCommitTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		err = tx.ExecContext(ctx, StageAllTablesForDoltCommitToolSQLQuery)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(StageAllTablesForDoltCommitToolCallSuccessMessage)
		return
	})
}
