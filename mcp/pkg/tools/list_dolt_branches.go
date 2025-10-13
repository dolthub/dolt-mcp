package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListDoltBranchesToolName        = "list_dolt_branches"
	ListDoltBranchesToolSQLQuery    = "SELECT * FROM dolt_branches;"
	ListDoltBranchesToolDescription = "Lists all Dolt branches."
)

func NewListDoltBranchesTool() mcp.Tool {
	return mcp.NewTool(
		ListDoltBranchesToolName,
		mcp.WithDescription(ListDoltBranchesToolDescription),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString(
			WorkingDatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingDatabaseCallToolArgumentDescription),
		),
	)
}

func RegisterListDoltBranchesTool(server pkg.Server) {
	mcpServer := server.MCP()
	listDoltBranchesTool := NewListDoltBranchesTool()
	mcpServer.AddTool(listDoltBranchesTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error

		var workingDatabase string
		workingDatabase, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
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
			tx.Rollback(ctx)
		}()

		var formattedResult string
		formattedResult, err = tx.QueryContext(ctx, ListDoltBranchesToolSQLQuery, db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}
