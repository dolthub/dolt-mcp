package tools

import (
	"context"
	"errors"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CallDoltBranchToolName                              = "call_dolt_branch"
	CallDoltBranchToolOriginalBranchArgumentDescription = "The name of the original branch."
	CallDoltBranchToolNewBranchArgumentDescription      = "The name of the new branch."
	CallDoltBranchToolCreateBranchArgumentDescription   = "If true, copies the new branch from original branch."
	CallDoltBranchToolMoveBranchArgumentDescription     = "If true, moves the original branch to be the new branch."
	CallDoltBranchToolDeleteBranchArgumentDescription   = "If true, deletes original branch."
	CallDoltBranchToolForceArgumentDescription          = "If true, will force the operation to succeed. If the operation is create, the new branch will be created again if it already exists. If the operation is delete, it will delete the original branch even if there are uncommitted changes in the working set."
	CallDoltBranchToolDescription                       = "Calls the DOLT_BRANCH stored procedure."
	CallDoltBranchToolCallSuccessMessage                = "successfully called DOLT_BRANCH"
)

var ErrNewBranchArgumentMissing = errors.New("new_branch argument missing")
var ErrOnlyOneOperationOptionAllowed = errors.New("specify only one operation option")

func ValidateCallDoltBranchCallToolRequest(request mcp.CallToolRequest) error {
	_, err := GetRequiredStringArgumentFromCallToolRequest(request, OriginalBranchCallToolArgumentName)
	if err != nil {
		return err
	}

	deleteArg := GetBooleanArgumentFromCallToolRequest(request, DeleteCallToolArgumentName)
	createArg := GetBooleanArgumentFromCallToolRequest(request, CreateCallToolArgumentName)
	moveArg := GetBooleanArgumentFromCallToolRequest(request, MoveCallToolArgumentName)

	if deleteArg && createArg || deleteArg && moveArg || moveArg && createArg {
		return ErrOnlyOneOperationOptionAllowed
	}
	
	newBranch := GetStringArgumentFromCallToolRequest(request, NewBranchCallToolArgumentName)
	if !deleteArg {
		if newBranch == "" {
			return ErrNewBranchArgumentMissing
		}
	}

	return nil
}

func RegisterCallDoltBranchTool(server pkg.Server) {
	mcpServer := server.MCP()

	createTableTool := mcp.NewTool(
		CreateTableToolName,
		mcp.WithDescription(CreateTableToolDescription),
		mcp.WithString(
			QueryCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CreateTableToolQueryArgumentDescription),
		),
	)

	mcpServer.AddTool(createTableTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var createTableStatement string
		createTableStatement, err = GetRequiredStringArgumentFromCallToolRequest(request, QueryCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		err = ValidateCreateTableQuery(createTableStatement)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		config := server.DBConfig()
		var tx db.DatabaseTransaction
		tx, err = db.NewDatabaseTransaction(ctx, config)
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

		err = tx.ExecContext(ctx, createTableStatement)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(CreateTableToolCallSuccessMessage)
		return
	})
}
