package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	RemoveDoltTestToolName                = "remove_dolt_test"
	RemoveDoltTestToolDescription         = "Removes a row from the dolt_tests table."
	RemoveDoltTestToolSuccessFormatString = "successfully removed test: %s"
)

// NewRemoveDoltTestTool defines a write tool to delete a test by name
func NewRemoveDoltTestTool() mcp.Tool {
	return mcp.NewTool(
		RemoveDoltTestToolName,
		mcp.WithDescription(RemoveDoltTestToolDescription),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString(WorkingDatabaseCallToolArgumentName, mcp.Required(), mcp.Description(WorkingDatabaseCallToolArgumentDescription)),
		mcp.WithString(WorkingBranchCallToolArgumentName, mcp.Required(), mcp.Description(WorkingBranchCallToolArgumentDescription)),
		mcp.WithString(TestNameCallToolArgumentName, mcp.Required(), mcp.Description("Unique test_name identifier")),
	)
}

func RegisterRemoveDoltTestTool(server pkg.Server) {
	mcpServer := server.MCP()
	tool := NewRemoveDoltTestTool()

	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		branch, err := GetRequiredStringArgumentFromCallToolRequest(request, WorkingBranchCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dbName, err := GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		name, err := GetRequiredStringArgumentFromCallToolRequest(request, TestNameCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		config := server.DBConfig()
		tx, err := NewDatabaseTransactionUsingDatabaseOnBranch(ctx, config, dbName, branch)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer func() {
			if rerr := CommitTransactionOrRollbackOnError(ctx, tx, err); rerr != nil && err == nil {
				result = mcp.NewToolResultError(rerr.Error())
			}
		}()

		query := fmt.Sprintf("DELETE FROM dolt_tests WHERE test_name = '%s'", singleQuoteEscape(name))
		if err = tx.ExecContext(ctx, query); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf(RemoveDoltTestToolSuccessFormatString, name)), nil
	})
}
