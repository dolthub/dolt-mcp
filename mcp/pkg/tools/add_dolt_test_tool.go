package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	AddDoltTestToolName                = "add_dolt_test"
	AddDoltTestToolDescription         = "Adds or updates a row in the dolt_tests table."
	AddDoltTestToolSuccessFormatString = "successfully upserted test: %s"
)

// NewAddDoltTestTool defines a write tool to upsert a test definition
func NewAddDoltTestTool() mcp.Tool {
	return mcp.NewTool(
		AddDoltTestToolName,
		mcp.WithDescription(AddDoltTestToolDescription),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString(WorkingDatabaseCallToolArgumentName, mcp.Required(), mcp.Description(WorkingDatabaseCallToolArgumentDescription)),
		mcp.WithString(WorkingBranchCallToolArgumentName, mcp.Required(), mcp.Description(WorkingBranchCallToolArgumentDescription)),
		mcp.WithString(TestNameCallToolArgumentName, mcp.Required(), mcp.Description("Unique test_name identifier")),
		mcp.WithString(TestGroupCallToolArgumentName, mcp.Description("Optional test_group for grouping tests")),
		mcp.WithString(QueryCallToolArgumentName, mcp.Required(), mcp.Description("Read-only SQL to validate")),
		mcp.WithString(AssertionTypeCallToolArgumentName, mcp.Required(), mcp.Description("Assertion type: expected_rows | expected_columns | expected_single_value")),
		mcp.WithString(AssertionComparatorCallToolArgumentName, mcp.Required(), mcp.Description("Comparator: == | != | < | > | <= | >=")),
		mcp.WithString(AssertionValueCallToolArgumentName, mcp.Description("Optional assertion value; set empty for NULL")),
	)
}

func RegisterAddDoltTestTool(server pkg.Server) {
	mcpServer := server.MCP()
	tool := NewAddDoltTestTool()

	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		branch, err := GetRequiredStringArgumentFromCallToolRequest(request, WorkingBranchCallToolArgumentName)
		if err != nil { return mcp.NewToolResultError(err.Error()), nil }
		dbName, err := GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
		if err != nil { return mcp.NewToolResultError(err.Error()), nil }

		name, err := GetRequiredStringArgumentFromCallToolRequest(request, TestNameCallToolArgumentName)
		if err != nil { return mcp.NewToolResultError(err.Error()), nil }
		group := GetStringArgumentFromCallToolRequest(request, TestGroupCallToolArgumentName)
		query, err := GetRequiredStringArgumentFromCallToolRequest(request, QueryCallToolArgumentName)
		if err != nil { return mcp.NewToolResultError(err.Error()), nil }
		atype, err := GetRequiredStringArgumentFromCallToolRequest(request, AssertionTypeCallToolArgumentName)
		if err != nil { return mcp.NewToolResultError(err.Error()), nil }
		comp, err := GetRequiredStringArgumentFromCallToolRequest(request, AssertionComparatorCallToolArgumentName)
		if err != nil { return mcp.NewToolResultError(err.Error()), nil }
		aval := GetStringArgumentFromCallToolRequest(request, AssertionValueCallToolArgumentName)

		// Ensure the provided query is read-only, aligning with dolt_test_run rules
		if err := ValidateReadQuery(query); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		config := server.DBConfig()
		tx, err := NewDatabaseTransactionUsingDatabaseOnBranch(ctx, config, dbName, branch)
		if err != nil { return mcp.NewToolResultError(err.Error()), nil }
		defer func() {
			if rerr := CommitTransactionOrRollbackOnError(ctx, tx, err); rerr != nil && err == nil {
				result = mcp.NewToolResultError(rerr.Error())
			}
		}()

		// Upsert semantics via REPLACE INTO to simplify
		// dolt_tests schema: test_name (PK), test_group, test_query, assertion_type, assertion_comparator, assertion_value
		var stmt string
		qName := singleQuoteEscape(name)
		qGroup := singleQuoteEscape(group)
		qQuery := singleQuoteEscape(query)
		qType := singleQuoteEscape(atype)
		qComp := singleQuoteEscape(comp)
		qVal := singleQuoteEscape(aval)

		if group == "" && aval == "" {
			stmt = fmt.Sprintf("REPLACE INTO dolt_tests VALUES ('%s', NULL, '%s', '%s', '%s', NULL)", qName, qQuery, qType, qComp)
		} else if group == "" {
			stmt = fmt.Sprintf("REPLACE INTO dolt_tests VALUES ('%s', NULL, '%s', '%s', '%s', '%s')", qName, qQuery, qType, qComp, qVal)
		} else if aval == "" {
			stmt = fmt.Sprintf("REPLACE INTO dolt_tests VALUES ('%s', '%s', '%s', '%s', '%s', NULL)", qName, qGroup, qQuery, qType, qComp)
		} else {
			stmt = fmt.Sprintf("REPLACE INTO dolt_tests VALUES ('%s', '%s', '%s', '%s', '%s', '%s')", qName, qGroup, qQuery, qType, qComp, qVal)
		}

		if err = tx.ExecContext(ctx, stmt); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf(AddDoltTestToolSuccessFormatString, name)), nil
	})
}
