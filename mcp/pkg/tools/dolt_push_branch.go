package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DoltPushBranchToolName                          = "dolt_push_branch"
	DoltPushBranchToolRemoteNameArgumentDescription = "The name of the remote to push the branch to."
	DoltPushBranchToolBranchArgumentDescription     = "The name of the local branch to push."
	DoltPushBranchToolForceArgumentDescription      = "If true, the specified branch is force pushed."
	DoltPushBranchToolDescription                   = "Pushes the specified branch to the remote."
	DoltPushBranchToolCallSuccessFormatString       = "successfully pushed branch: %s"
	DoltPushBranchToolSQLQueryFormatString          = "CALL DOLT_PUSH('%s', '%s');"
	DoltPushBranchToolForceSQLQueryFormatString     = "CALL DOLT_PUSH('--force', '%s', '%s');"
)

func RegisterDoltPushBranchTool(server pkg.Server) {
	mcpServer := server.MCP()

	doltPushBranchTool := mcp.NewTool(
		DoltPushBranchToolName,
		mcp.WithDescription(DoltPushBranchToolDescription),
		mcp.WithString(
			RemoteNameCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DoltPushBranchToolRemoteNameArgumentDescription),
		),
		mcp.WithString(
			BranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DoltPushBranchToolBranchArgumentDescription),
		),
		mcp.WithBoolean(
			ForceCallToolArgumentName,
			mcp.Description(DoltPushBranchToolForceArgumentDescription),
		),
	)

	mcpServer.AddTool(doltPushBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
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
		tx, err = db.NewDatabaseTransaction(ctx, config)
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
			err = tx.ExecContext(ctx, fmt.Sprintf(DoltPushBranchToolForceSQLQueryFormatString, remote, branch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, fmt.Sprintf(DoltPushBranchToolSQLQueryFormatString, remote, branch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(fmt.Sprintf(DoltPushBranchToolCallSuccessFormatString, branch))
		return
	})
}
