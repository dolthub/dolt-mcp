package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DoltFetchAllBranchesToolName                          = "dolt_fetch_all_branches"
	DoltFetchAllBranchesToolRemoteNameArgumentDescription = "The name of the remote to fetch all branches from."
	DoltFetchAllBranchesToolForceArgumentDescription      = "If true, all branches are force fetched."
	DoltFetchAllBranchesToolDescription                   = "Fetches all branches from the remote."
	DoltFetchAllBranchesToolCallSuccessMessage            = "successfully fetched branches"
	DoltFetchAllBranchesToolSQLQueryFormatString          = "CALL DOLT_FETCH('%s');"
	DoltFetchAllBranchesToolForceSQLQueryFormatString     = "CALL DOLT_FETCH('%s', '--force');"
)

func RegisterDoltFetchAllBranchesTool(server pkg.Server) {
	mcpServer := server.MCP()

	doltFetchAllBranchesTool := mcp.NewTool(
		DoltFetchAllBranchesToolName,
		mcp.WithDescription(DoltFetchAllBranchesToolDescription),
		mcp.WithString(
			RemoteNameCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DoltFetchAllBranchesToolRemoteNameArgumentDescription),
		),
		mcp.WithBoolean(
			ForceCallToolArgumentName,
			mcp.Description(DoltFetchAllBranchesToolForceArgumentDescription),
		),
	)

	mcpServer.AddTool(doltFetchAllBranchesTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var remote string
		remote, err = GetRequiredStringArgumentFromCallToolRequest(request, RemoteNameCallToolArgumentName)
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
			err = tx.ExecContext(ctx, fmt.Sprintf(DoltFetchAllBranchesToolForceSQLQueryFormatString, remote))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, fmt.Sprintf(DoltFetchAllBranchesToolSQLQueryFormatString, remote))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(DoltFetchAllBranchesToolCallSuccessMessage)
		return
	})
}

