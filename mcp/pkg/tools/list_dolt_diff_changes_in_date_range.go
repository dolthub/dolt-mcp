package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListDoltDiffChangesInDateRangeToolName                         = "list_dolt_diff_changes_in_date_range"
	ListDoltDiffChangesInDateRangeToolStartDateArgumentDescription = "The start date of the range in the format 'YYYY-MM-DD'."
	ListDoltDiffChangesInDateRangeToolEndDateArgumentDescription   = "The end date of the range in the format 'YYYY-MM-DD'."
	ListDoltDiffChangesInDateRangeToolDescription                  = "Lists dolt_diff changes that were created between the specified start and end dates."
	ListDoltDiffChangesInDateRangeToolSQLQueryFormatString         = "SELECT * FROM dolt_diff WHERE date BETWEEN '%s' AND '%s';"
)

func RegisterListDoltDiffChangesInDateRangeTool(server pkg.Server) {
	mcpServer := server.MCP()

	listDoltDiffChangesInDateRangeTool := mcp.NewTool(
		ListDoltDiffChangesInDateRangeToolName,
		mcp.WithDescription(CreateDoltBranchToolDescription),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
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
		mcp.WithString(
			StartDateCallToolArgumentName,
			mcp.Required(),
			mcp.Description(ListDoltDiffChangesInDateRangeToolStartDateArgumentDescription),
		),
		mcp.WithString(
			EndDateCallToolArgumentName,
			mcp.Required(),
			mcp.Description(ListDoltDiffChangesInDateRangeToolEndDateArgumentDescription),
		),
	)

	mcpServer.AddTool(listDoltDiffChangesInDateRangeTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var startDate string
		startDate, err = GetRequiredStringArgumentFromCallToolRequest(request, StartDateCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var endDate string
		endDate, err = GetRequiredStringArgumentFromCallToolRequest(request, EndDateCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		config := server.DBConfig()

		var tx db.DatabaseTransaction
		tx, err = NewDatabaseTransactionOnBranchUsingDatabase(ctx, config, workingBranch, workingDatabase)
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

		var formattedResult string
		formattedResult, err = tx.QueryContext(ctx, fmt.Sprintf(ListDoltDiffChangesInDateRangeToolSQLQueryFormatString, startDate, endDate), db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}
