package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	AbortDoltMergeToolName               = "abort_dolt_merge"
	AbortDoltMergeToolDescription        = "Aborts the current merge."
	AbortDoltMergeToolSQLQuery           = "CALL DOLT_MERGE('--abort');"
	AbortDoltMergeToolCallSuccessMessage = "successfully aborted merge"
)

func RegisterAbortDoltMergeTool(server pkg.Server) {
	mcpServer := server.MCP()

	abortDoltMergeTool := mcp.NewTool(
		AbortDoltMergeToolName,
		mcp.WithDescription(AbortDoltMergeToolDescription),
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

	mcpServer.AddTool(abortDoltMergeTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		err = tx.ExecContext(ctx, AbortDoltMergeToolSQLQuery)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(AbortDoltMergeToolCallSuccessMessage)
		return
	})
}
