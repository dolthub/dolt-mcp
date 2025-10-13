package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	MergeDoltBranchToolName                            = "merge_dolt_branch"
	MergeDoltBranchToolBranchArgumentDescription       = "The name of the branch to merge into the currently checked out branch."
	MergeDoltBranchToolMessageArgumentDescription      = "The message for the Dolt commit resulting from a successful merge."
	MergeDoltBranchToolDescription                     = "Merges the specified branch into the currently checked out branch."
	MergeDoltBranchToolSQLQueryFormatString            = "CALL DOLT_MERGE('%s');"
	MergeDoltBranchToolWithMessageSQLQueryFormatString = "CALL DOLT_MERGE('%s', '-m', '%s');"
	MergeDoltBranchToolCallSuccessMessage              = "successfully merged branch"
)

func NewMergeDoltBranchTool() mcp.Tool {
	return mcp.NewTool(
		MergeDoltBranchToolName,
		mcp.WithDescription(MergeDoltBranchToolDescription),
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

func RegisterMergeDoltBranchTool(server pkg.Server) {
	mcpServer := server.MCP()
	mergeDoltBranchTool := NewMergeDoltBranchTool()

	mcpServer.AddTool(mergeDoltBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		if commitMessage != "" {
			err = tx.ExecContext(ctx, fmt.Sprintf(MergeDoltBranchToolWithMessageSQLQueryFormatString, branch, commitMessage))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, fmt.Sprintf(MergeDoltBranchToolSQLQueryFormatString, branch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(MergeDoltBranchToolCallSuccessMessage)
		return
	})
}
