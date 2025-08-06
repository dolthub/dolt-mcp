package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListDoltDiffChangesByTableNameToolName                          = "list_dolt_diff_changes_by_table_name"
	ListDoltDiffChangesByTableNameToolTableArgumentDescription      = "The name of the table."
	ListDoltDiffChangesByTableNameToolFromCommitArgumentDescription = "The 'from' commit of the Dolt diff."
	ListDoltDiffChangesByTableNameToolToCommitArgumentDescription   = "The 'to' commit of the Dolt diff."
	ListDoltDiffChangesByTableNameToolDescription                   = "Lists dolt_diff changes for the specified table between two Dolt commits."
	ListDoltDiffChangesByTableNameToolSQLQueryFormatString          = "SELECT * FROM dolt_diff_%s WHERE from_commit = '%s' AND to_commit = '%s';"
)

func RegisterListDoltDiffChangesByTableNameTool(server pkg.Server) {
	mcpServer := server.MCP()

	listDoltDiffChangesByTableNameTool := mcp.NewTool(
		ListDoltDiffChangesByTableNameToolName,
		mcp.WithDescription(ListDoltDiffChangesByTableNameToolDescription),
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
			TableCallToolArgumentName,
			mcp.Required(),
			mcp.Description(ListDoltDiffChangesByTableNameToolTableArgumentDescription),
		),
		mcp.WithString(
			FromCommitCallToolArgumentName,
			mcp.Required(),
			mcp.Description(ListDoltDiffChangesByTableNameToolFromCommitArgumentDescription),
		),
		mcp.WithString(
			ToCommitCallToolArgumentName,
			mcp.Required(),
			mcp.Description(ListDoltDiffChangesByTableNameToolFromCommitArgumentDescription),
		),
	)

	mcpServer.AddTool(listDoltDiffChangesByTableNameTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var table string
		table, err = GetRequiredStringArgumentFromCallToolRequest(request, TableCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var fromCommit string
		fromCommit, err = GetRequiredStringArgumentFromCallToolRequest(request, FromCommitCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var toCommit string
		toCommit, err = GetRequiredStringArgumentFromCallToolRequest(request, ToCommitCallToolArgumentName)
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
		formattedResult, err = tx.QueryContext(ctx, fmt.Sprintf(ListDoltDiffChangesByTableNameToolSQLQueryFormatString, table, fromCommit, toCommit), db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}
