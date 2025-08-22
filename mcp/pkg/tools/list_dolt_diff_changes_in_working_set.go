package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListDoltDiffChangesInWorkingSetToolName        = "list_dolt_diff_changes_in_working_set"
	ListDoltDiffChangesInWorkingSetToolSQLQuery    = "SELECT * FROM dolt_diff WHERE commit_hash='WORKING';"
	ListDoltDiffChangesInWorkingSetToolDescription = "Lists all dolt_diff changes in the current working set."
)

func NewListDoltDiffChangesInWorkingSetTool() mcp.Tool {
    return mcp.NewTool(
        ListDoltDiffChangesInWorkingSetToolName,
        mcp.WithDescription(ListDoltDiffChangesInWorkingSetToolDescription),
        mcp.WithReadOnlyHintAnnotation(true),
        mcp.WithDestructiveHintAnnotation(false),
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
    )
}

func RegisterListDoltDiffChangesInWorkingSetTool(server pkg.Server) {
    mcpServer := server.MCP()
    listDoltDiffChangesInWorkingSetTool := NewListDoltDiffChangesInWorkingSetTool()
	mcpServer.AddTool(listDoltDiffChangesInWorkingSetTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		config := server.DBConfig()

		var tx db.DatabaseTransaction
		tx, err = NewDatabaseTransactionUsingDatabaseOnBranch(ctx, config, workingDatabase, workingBranch)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		defer func() {
			tx.Rollback(ctx)
		}()

		var formattedResult string
		formattedResult, err = tx.QueryContext(ctx, ListDoltDiffChangesInWorkingSetToolSQLQuery, db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}
