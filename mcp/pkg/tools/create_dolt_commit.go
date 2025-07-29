package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateDoltCommitToolName                       = "create_dolt_commit"
	CreateDoltCommitToolMessageArgumentDescription = "The message to use in the Dolt commit."
	CreateDoltCommitToolSQLQueryFormatString       = "CALL DOLT_COMMIT('-m', '%s');"
	CreateDoltCommitToolDescription                = "Creates a Dolt commit with the specified message."
	CreateDoltCommitToolCallSuccessMessage         = "successfully committed changes"
)

func RegisterCreateDoltCommitTool(server pkg.Server) {
	mcpServer := server.MCP()

	createDoltCommitTool := mcp.NewTool(
		CreateDoltCommitToolName,
		mcp.WithDescription(CreateDoltCommitToolDescription),
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
			MessageCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CreateDoltCommitToolMessageArgumentDescription),
		),
	)

	mcpServer.AddTool(createDoltCommitTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var message string
		message, err = GetRequiredStringArgumentFromCallToolRequest(request, MessageCallToolArgumentName)
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

		err = tx.ExecContext(ctx, fmt.Sprintf(CreateDoltCommitToolSQLQueryFormatString, message))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(CreateDoltCommitToolCallSuccessMessage)
		return
	})
}

