package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DoltResetSoftToolName                        = "dolt_reset_soft"
	DoltResetSoftToolRevisionArgumentDescription = "The revision to move the branch HEAD to (a branch name, commit sha, or ancestor spec such as 'HEAD~1'). Table names are not valid revisions."
	DoltResetSoftToolDescription                 = "Moves the branch HEAD to the specified revision, leaving the staging area and working set unchanged. To unstage tables instead, use the unstage_table or unstage_all_tables tools."
	DoltResetSoftToolCallSuccessFormatString     = "successfully soft reset: %s"
)

func NewDoltResetSoftTool() mcp.Tool {
	return mcp.NewTool(
		DoltResetSoftToolName,
		mcp.WithDescription(DoltResetSoftToolDescription),
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
			RevisionCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DoltResetSoftToolRevisionArgumentDescription),
		),
	)
}

func RegisterDoltResetSoftTool(server pkg.Server) {
	mcpServer := server.MCP()
	resetSoftTool := NewDoltResetSoftTool()

	mcpServer.AddTool(resetSoftTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var revision string
		revision, err = GetRequiredStringArgumentFromCallToolRequest(request, RevisionCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		dialect := server.Dialect()
		config := server.DBConfig()

		var tx db.DatabaseTransaction
		tx, err = NewDatabaseTransactionUsingDatabaseOnBranch(ctx, config, dialect, workingDatabase, workingBranch)
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

		err = tx.ExecContext(ctx, dialect.CallProcedure(db.DoltReset, "--soft", revision))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(DoltResetSoftToolCallSuccessFormatString, revision))
		return
	})
}
