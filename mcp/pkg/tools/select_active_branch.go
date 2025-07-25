package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	SelectActiveBranchToolName        = "select_active_branch"
	SelectActiveBranchToolSQLQuery    = "SELECT ACTIVE_BRANCH();"
	SelectActiveBranchToolDescription = "Displays the currently checked out branch."
)

func RegisterSelectActiveBranchTool(server pkg.Server) {
	mcpServer := server.MCP()

	selectActiveBranchTool := mcp.NewTool(SelectActiveBranchToolName, mcp.WithDescription(SelectActiveBranchToolDescription))
	mcpServer.AddTool(selectActiveBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error

		config := server.DBConfig()

		var tx db.DatabaseTransaction
		tx, err = db.NewDatabaseTransaction(ctx, config)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		defer func() {
			tx.Rollback(ctx)
		}()

		var formattedResult string
		formattedResult, err = tx.QueryContext(ctx, SelectActiveBranchToolSQLQuery, db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}
