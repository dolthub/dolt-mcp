package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DoltResetHardToolName                                 = "dolt_reset_hard"
	DoltResetHardToolBranchOrCommitSHAArgumentDescription = "The branch or commit sha to hard reset."
	DoltResetHardToolDescription                          = "Hard resets the specified branch."
	DoltResetHardToolSQLQueryFormatString                 = "CALL DOLT_RESET('--hard', '%s');"
	DoltResetHardToolCallSuccessFormatString              = "successfully hard reset: %s"
)

func RegisterDoltResetHardTool(server pkg.Server) {
	mcpServer := server.MCP()

	resetHardTool := mcp.NewTool(
		DoltResetHardToolName,
		mcp.WithDescription(DoltResetHardToolDescription),
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
			BranchOrCommitSHACallToolArgumentName,
			mcp.Required(),
			mcp.Description(DoltResetHardToolBranchOrCommitSHAArgumentDescription),
		),
	)

	mcpServer.AddTool(resetHardTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var branchOrCommitSHA string
		branchOrCommitSHA, err = GetRequiredStringArgumentFromCallToolRequest(request, BranchOrCommitSHACallToolArgumentName)
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

		err = tx.ExecContext(ctx, fmt.Sprintf(DoltResetHardToolSQLQueryFormatString, branchOrCommitSHA))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(DoltResetHardToolCallSuccessFormatString, branchOrCommitSHA))
		return
	})
}
