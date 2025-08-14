package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	MergeDoltBranchNoFastForwardToolName                            = "merge_dolt_branch_no_fast_forward"
	MergeDoltBranchNoFastForwardToolDescription                     = "Performs a non fast-foward merge of the specified branch into the currently checked out branch."
	MergeDoltBranchNoFastForwardToolSQLQueryFormatString            = "CALL DOLT_MERGE('%s', '--no-ff');"
	MergeDoltBranchNoFastForwardToolWithMessageSQLQueryFormatString = "CALL DOLT_MERGE('%s', '--no-ff', '-m', '%s');"
)

func NewMergeDoltBranchNoFastForwardTool() mcp.Tool {
    return mcp.NewTool(
        MergeDoltBranchNoFastForwardToolName,
        mcp.WithDescription(MergeDoltBranchNoFastForwardToolDescription),
        mcp.WithReadOnlyHintAnnotation(false),
        mcp.WithDestructiveHintAnnotation(true),
        mcp.WithIdempotentHintAnnotation(false),
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
            BranchCallToolArgumentName,
            mcp.Required(),
            mcp.Description(MergeDoltBranchToolBranchArgumentDescription),
        ),
        mcp.WithString(
            MessageCallToolArgumentName,
            mcp.Description(MergeDoltBranchToolMessageArgumentDescription),
        ),
    )
}

func RegisterMergeDoltBranchNoFastForwardTool(server pkg.Server) {
    mcpServer := server.MCP()
    mergeDoltBranchNoFastForwardTool := NewMergeDoltBranchNoFastForwardTool()

	mcpServer.AddTool(mergeDoltBranchNoFastForwardTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var branch string
		branch, err = GetRequiredStringArgumentFromCallToolRequest(request, BranchCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		commitMessage := GetStringArgumentFromCallToolRequest(request, MessageCallToolArgumentName)

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

		if commitMessage != "" {
			err = tx.ExecContext(ctx, fmt.Sprintf(MergeDoltBranchNoFastForwardToolWithMessageSQLQueryFormatString, branch, commitMessage))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, fmt.Sprintf(MergeDoltBranchNoFastForwardToolSQLQueryFormatString, branch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(MergeDoltBranchToolCallSuccessMessage)
		return
	})
}
