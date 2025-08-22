package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DeleteDoltBranchToolName                      = "delete_dolt_branch"
	DeleteDoltBranchToolBranchArgumentDescription = "The name of the branch to delete."
	DeleteDoltBranchToolForceArgumentDescription  = "If true, will force the deletion of the specified branch even if it has uncommitted changes."
	DeleteDoltBranchToolDescription               = "Deletes a branch."
	DeleteDoltBranchToolCallSuccessFormatString   = "successfully deleted branch: %s"
	DeleteDoltBranchToolSQLQueryFormatString      = "CALL DOLT_BRANCH('-d', '%s');"
	DeleteDoltBranchToolForceSQLQueryFormatString = "CALL DOLT_BRANCH('-f', '-d', '%s');"
)

func NewDeleteDoltBranchTool() mcp.Tool {
    return mcp.NewTool(
        DeleteDoltBranchToolName,
        mcp.WithDescription(DeleteDoltBranchToolDescription),
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
            BranchCallToolArgumentName,
            mcp.Required(),
            mcp.Description(DeleteDoltBranchToolBranchArgumentDescription),
        ),
        mcp.WithBoolean(
            ForceCallToolArgumentName,
            mcp.Description(DeleteDoltBranchToolForceArgumentDescription),
        ),
    )
}

func RegisterDeleteDoltBranchTool(server pkg.Server) {
    mcpServer := server.MCP()
    deleteDoltBranchTool := NewDeleteDoltBranchTool()

	mcpServer.AddTool(deleteDoltBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error

		var workingDatabase string
		workingDatabase, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var workingBranch string
		workingBranch, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingBranchCallToolArgumentName)
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

		force := GetBooleanArgumentFromCallToolRequest(request, ForceCallToolArgumentName)

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

		if force {
			err = tx.ExecContext(ctx, fmt.Sprintf(DeleteDoltBranchToolForceSQLQueryFormatString, branch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, fmt.Sprintf(DeleteDoltBranchToolSQLQueryFormatString, branch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(fmt.Sprintf(DeleteDoltBranchToolCallSuccessFormatString, branch))
		return
	})
}
