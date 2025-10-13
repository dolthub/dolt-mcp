package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DoltFetchBranchToolName                          = "dolt_fetch_branch"
	DoltFetchBranchToolRemoteNameArgumentDescription = "The name of the remote to fetch the branch from."
	DoltFetchBranchToolBranchArgumentDescription     = "The name of the remote branch to fetch."
	DoltFetchBranchToolDescription                   = "Fetches the specified branch from the remote."
	DoltFetchBranchToolCallSuccessFormatString       = "successfully fetched branch: %s"
	DoltFetchBranchToolSQLQueryFormatString          = "CALL DOLT_FETCH('%s', '%s');"
)

func NewDoltFetchBranchTool() mcp.Tool {
	return mcp.NewTool(
		DoltFetchBranchToolName,
		mcp.WithDescription(DoltFetchBranchToolDescription),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithString(
			WorkingDatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingDatabaseCallToolArgumentDescription),
		),
		mcp.WithString(
			RemoteNameCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DoltFetchBranchToolRemoteNameArgumentDescription),
		),
		mcp.WithString(
			BranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DoltFetchBranchToolBranchArgumentDescription),
		),
	)
}

func RegisterDoltFetchBranchTool(server pkg.Server) {
	mcpServer := server.MCP()
	doltFetchBranchTool := NewDoltFetchBranchTool()

	mcpServer.AddTool(doltFetchBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		err = tx.ExecContext(ctx, fmt.Sprintf(DoltFetchBranchToolSQLQueryFormatString, remote, branch))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(DoltFetchBranchToolCallSuccessFormatString, branch))
		return
	})
}
