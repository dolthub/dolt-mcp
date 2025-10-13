package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
    DoltResetHardToolName                        = "dolt_reset_hard"
    DoltResetHardToolRevisionArgumentDescription = "The revision to reset to (working set, table name, branch, commit sha, or '.' for all tables)."
    DoltResetHardToolDescription                 = "Hard resets the working set to the specified revision."
    DoltResetHardToolSQLQueryFormatString        = "CALL DOLT_RESET('--hard', '%s');"
    DoltResetHardToolCallSuccessFormatString     = "successfully hard reset: %s"
)

func NewDoltResetHardTool() mcp.Tool {
    return mcp.NewTool(
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
            RevisionCallToolArgumentName,
            mcp.Required(),
            mcp.Description(DoltResetHardToolRevisionArgumentDescription),
        ),
    )
}

func RegisterDoltResetHardTool(server pkg.Server) {
    mcpServer := server.MCP()
    resetHardTool := NewDoltResetHardTool()

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

        var revision string
        revision, err = GetRequiredStringArgumentFromCallToolRequest(request, RevisionCallToolArgumentName)
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

        err = tx.ExecContext(ctx, fmt.Sprintf(DoltResetHardToolSQLQueryFormatString, revision))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

        result = mcp.NewToolResultText(fmt.Sprintf(DoltResetHardToolCallSuccessFormatString, revision))
		return
	})
}
