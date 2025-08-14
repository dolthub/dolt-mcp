package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListDoltDiffChangesByTableNameToolName                                = "list_dolt_diff_changes_by_table_name"
	ListDoltDiffChangesByTableNameToolTableArgumentDescription            = "The name of the table."
	ListDoltDiffChangesByTableNameToolFromCommitArgumentDescription       = "The 'from' commit of the Dolt diff."
	ListDoltDiffChangesByTableNameToolHashOfFromCommitArgumentDescription = "The value to supply to the HASHOF() function for the 'from' commit."
	ListDoltDiffChangesByTableNameToolHashOfToCommitArgumentDescription   = "The value to supply to the HASHOF() function for the 'to' commit."
	ListDoltDiffChangesByTableNameToolToCommitArgumentDescription         = "The 'to' commit of the Dolt diff."
	ListDoltDiffChangesByTableNameToolDescription                         = "Lists dolt_diff changes for the specified table between two Dolt commits."
	ListDoltDiffChangesByTableNameToolSQLQueryFormatString                = "SELECT * FROM dolt_diff_%s WHERE from_commit = %s AND to_commit = %s;"
)

var ErrFromCommitOrHashOfFromCommitArgumentRequiredMessage = "from_commit or hash_of_from_commit argument required"
var ErrSpecifyEitherFromCommitOrHashOfFromCommitMessage = "specify either from_commit or hash_of_from_commit arguments, not both "
var ErrToCommitOrHashOfToCommitArgumentRequiredMessage = "to_commit or hash_of_to_commit argument required"
var ErrSpecifyEitherToCommitOrHashOfToCommitMessage = "specify either to_commit or hash_of_to_commit arguments, not both "

func NewListDoltDiffChangesByTableNameTool() mcp.Tool {
    return mcp.NewTool(
        ListDoltDiffChangesByTableNameToolName,
        mcp.WithDescription(ListDoltDiffChangesByTableNameToolDescription),
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
        mcp.WithString(
            TableCallToolArgumentName,
            mcp.Required(),
            mcp.Description(ListDoltDiffChangesByTableNameToolTableArgumentDescription),
        ),
        mcp.WithString(
            FromCommitCallToolArgumentName,
            mcp.Description(ListDoltDiffChangesByTableNameToolFromCommitArgumentDescription),
        ),
        mcp.WithString(
            ToCommitCallToolArgumentName,
            mcp.Description(ListDoltDiffChangesByTableNameToolFromCommitArgumentDescription),
        ),
        mcp.WithString(
            HashOfFromCommitCallToolArgumentName,
            mcp.Description(ListDoltDiffChangesByTableNameToolFromCommitArgumentDescription),
        ),
        mcp.WithString(
            HashOfToCommitCallToolArgumentName,
            mcp.Description(ListDoltDiffChangesByTableNameToolFromCommitArgumentDescription),
        ),
    )
}

func RegisterListDoltDiffChangesByTableNameTool(server pkg.Server) {
    mcpServer := server.MCP()
    listDoltDiffChangesByTableNameTool := NewListDoltDiffChangesByTableNameTool()

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

		fromCommit := GetStringArgumentFromCallToolRequest(request, FromCommitCallToolArgumentName)
		hashOfFromCommit := GetStringArgumentFromCallToolRequest(request, HashOfFromCommitCallToolArgumentName)

		if fromCommit == "" && hashOfFromCommit == "" {
			result = mcp.NewToolResultError(ErrFromCommitOrHashOfFromCommitArgumentRequiredMessage)
			return
		}

		if fromCommit != "" && hashOfFromCommit != "" {
			result = mcp.NewToolResultError(ErrSpecifyEitherFromCommitOrHashOfFromCommitMessage)
			return
		}

		toCommit := GetStringArgumentFromCallToolRequest(request, ToCommitCallToolArgumentName)
		hashOfToCommit := GetStringArgumentFromCallToolRequest(request, HashOfToCommitCallToolArgumentName)

		if toCommit == "" && hashOfToCommit == "" {
			result = mcp.NewToolResultError(ErrToCommitOrHashOfToCommitArgumentRequiredMessage)
			return
		}

		if toCommit != "" && hashOfToCommit != "" {
			result = mcp.NewToolResultError(ErrSpecifyEitherToCommitOrHashOfToCommitMessage)
			return
		}

		var fromValue string
		if fromCommit != "" {
			fromValue = fmt.Sprintf("'%s'", fromCommit)
		} else {
			fromValue = fmt.Sprintf("HASHOF('%s')", hashOfFromCommit)
		}

		var toValue string
		if toCommit != "" {
			toValue = fmt.Sprintf("'%s'", toCommit)
		} else {
			toValue = fmt.Sprintf("HASHOF('%s')", hashOfToCommit)
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
		formattedResult, err = tx.QueryContext(ctx, fmt.Sprintf(ListDoltDiffChangesByTableNameToolSQLQueryFormatString, table, fromValue, toValue), db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}
