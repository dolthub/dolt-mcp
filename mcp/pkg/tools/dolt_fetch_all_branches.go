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
	DoltFetchAllBranchesToolDescription                   = "Fetches all branches from the remote."
	DoltFetchAllBranchesToolCallSuccessMessage            = "successfully fetched branches"
	DoltFetchAllBranchesToolSQLQueryFormatString          = "CALL DOLT_FETCH('%s');"
)

func NewDoltFetchAllBranchesTool() mcp.Tool {
    return mcp.NewTool(
        DoltFetchAllBranchesToolName,
        mcp.WithDescription(DoltFetchAllBranchesToolDescription),
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
            mcp.Description(DoltFetchAllBranchesToolRemoteNameArgumentDescription),
        ),
    )
}

func RegisterDoltFetchAllBranchesTool(server pkg.Server) {
    mcpServer := server.MCP()
    doltFetchAllBranchesTool := NewDoltFetchAllBranchesTool()

	mcpServer.AddTool(doltFetchAllBranchesTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		err = tx.ExecContext(ctx, fmt.Sprintf(DoltFetchAllBranchesToolSQLQueryFormatString, remote))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(DoltFetchAllBranchesToolCallSuccessMessage)
		return
	})
}

