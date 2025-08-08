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

func RegisterDoltFetchBranchTool(server pkg.Server) {
	mcpServer := server.MCP()

	doltFetchBranchTool := mcp.NewTool(
		DoltFetchBranchToolName,
		mcp.WithDescription(DoltFetchBranchToolDescription),
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

	mcpServer.AddTool(doltFetchBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		err = tx.ExecContext(ctx, fmt.Sprintf(DoltFetchBranchToolSQLQueryFormatString, remote, branch))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		branchesStr, err := tx.QueryContext(ctx, "select * from dolt_branches;", db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		fmt.Println("DUSTIN: branchesStr:", branchesStr)
		result = mcp.NewToolResultText(fmt.Sprintf(DoltFetchBranchToolCallSuccessFormatString, branch))
		return
	})
}
