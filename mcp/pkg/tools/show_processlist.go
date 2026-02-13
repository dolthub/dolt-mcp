package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ShowProcesslistToolName        = "show_processlist"
	ShowProcesslistToolDescription = "Show the server process list (SHOW [FULL] PROCESSLIST)."
	ShowProcesslistFullArgName     = "full"
	ShowProcesslistFullArgDesc     = "If true, uses SHOW FULL PROCESSLIST."
)

func showProcesslistSQL(full bool) string {
	if full {
		return "SHOW FULL PROCESSLIST;"
	}
	return "SHOW PROCESSLIST;"
}

func NewShowProcesslistTool() mcp.Tool {
	return mcp.NewTool(
		ShowProcesslistToolName,
		mcp.WithDescription(ShowProcesslistToolDescription),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString(
			WorkingBranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingBranchCallToolArgumentDescription),
		),
		mcp.WithString(
			WorkingDatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingDatabaseCallToolArgumentDescription),
		),
		mcp.WithBoolean(
			ShowProcesslistFullArgName,
			mcp.Description(ShowProcesslistFullArgDesc),
		),
	)
}

func RegisterShowProcesslistTool(server pkg.Server) {
	mcpServer := server.MCP()
	tool := NewShowProcesslistTool()

	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error

		workingBranch, err := GetRequiredStringArgumentFromCallToolRequest(request, WorkingBranchCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		workingDatabase, err := GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		full := GetBooleanArgumentFromCallToolRequest(request, ShowProcesslistFullArgName)

		config := server.DBConfig()
		tx, err := NewDatabaseTransactionUsingDatabaseOnBranch(ctx, config, workingDatabase, workingBranch)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer func() { _ = tx.Rollback(ctx) }()

		formattedResult, err := tx.QueryContext(ctx, showProcesslistSQL(full), db.ResultFormatMarkdown)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(formattedResult), nil
	})
}
