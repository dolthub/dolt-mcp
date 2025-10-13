package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DoltPullBranchToolName                          = "dolt_pull_branch"
	DoltPullBranchToolRemoteNameArgumentDescription = "The name of the remote to pull the branch from."
	DoltPullBranchToolBranchArgumentDescription     = "The name of the remote branch to pull."
	DoltPullBranchToolForceArgumentDescription      = "If true, the specified branch is force pulled."
	DoltPullBranchToolDescription                   = "Pulls the specified branch from the remote."
	DoltPullBranchToolCallSuccessFormatString       = "successfully pulled branch: %s"
	DoltPullBranchToolSQLQueryFormatString          = "CALL DOLT_PULL('%s', '%s');"
	DoltPullBranchToolForceSQLQueryFormatString     = "CALL DOLT_PULL('%s', '%s', '--force');"
)

func NewDoltPullBranchTool() mcp.Tool {
	return mcp.NewTool(
		DoltPullBranchToolName,
		mcp.WithDescription(DoltPullBranchToolDescription),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithString(
			WorkingDatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingBranchCallToolArgumentDescription),
		),
		mcp.WithString(
			RemoteNameCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DoltPullBranchToolRemoteNameArgumentDescription),
		),
		mcp.WithString(
			BranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DoltPullBranchToolBranchArgumentDescription),
		),
		mcp.WithBoolean(
			ForceCallToolArgumentName,
			mcp.Description(DoltPullBranchToolForceArgumentDescription),
		),
	)
}

func RegisterDoltPullBranchTool(server pkg.Server) {
	mcpServer := server.MCP()
	doltPullBranchTool := NewDoltPullBranchTool()

	mcpServer.AddTool(doltPullBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var workingDatabase string
		workingDatabase, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var remote string
		remote, err = GetRequiredStringArgumentFromCallToolRequest(request, RemoteNameCallToolArgumentName)
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
		tx, err = NewDatabaseTransactionUsingDatabase(ctx, config, workingDatabase)
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
			err = tx.ExecContext(ctx, fmt.Sprintf(DoltPullBranchToolForceSQLQueryFormatString, remote, branch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, fmt.Sprintf(DoltPullBranchToolSQLQueryFormatString, remote, branch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(fmt.Sprintf(DoltPullBranchToolCallSuccessFormatString, branch))
		return
	})
}
