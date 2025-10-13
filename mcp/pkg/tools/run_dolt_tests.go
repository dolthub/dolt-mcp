package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	RunDoltTestsToolName              = "run_dolt_tests"
	RunDoltTestsToolDescription       = "Runs dolt_tests via SELECT * FROM dolt_test_run(). Optionally filter by test name or group."
	RunDoltTestsToolTargetDescription = "Optional filter: '*' for all, a specific test_name, or a test_group. If omitted, runs all tests."
)

// NewRunDoltTestsTool defines a read-only tool that executes dolt_test_run()
func NewRunDoltTestsTool() mcp.Tool {
	return mcp.NewTool(
		RunDoltTestsToolName,
		mcp.WithDescription(RunDoltTestsToolDescription),
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
		mcp.WithString(
			TargetCallToolArgumentName,
			mcp.Description(RunDoltTestsToolTargetDescription),
		),
	)
}

// RegisterRunDoltTestsTool registers the tool which runs dolt_test_run().
func RegisterRunDoltTestsTool(server pkg.Server) {
	mcpServer := server.MCP()
	tool := NewRunDoltTestsTool()

	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error

		var workingBranch string
		workingBranch, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingBranchCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var workingDatabase string
		workingDatabase, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		target := GetStringArgumentFromCallToolRequest(request, TargetCallToolArgumentName)

		config := server.DBConfig()
		var tx db.DatabaseTransaction
		tx, err = NewDatabaseTransactionUsingDatabaseOnBranch(ctx, config, workingDatabase, workingBranch)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer func() { _ = tx.Rollback(ctx) }()

		// Build SELECT for dolt_test_run
		var query string
		if target == "" || target == "*" {
			query = "SELECT * FROM dolt_test_run()"
		} else {
			// Arguments to dolt_test_run are string literals; escape any single quotes
			query = fmt.Sprintf("SELECT * FROM dolt_test_run('%s')", singleQuoteEscape(target))
		}

		var formatted string
		formatted, err = tx.QueryContext(ctx, query, db.ResultFormatMarkdown)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(formatted), nil
	})
}
