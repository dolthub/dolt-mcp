package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateDoltBranchFromHeadToolName                         = "create_dolt_branch_from_head"
	CreateDoltBranchFromHeadToolNewBranchArgumentDescription = "The name of the new branch."
	CreateDoltBranchFromHeadToolForceArgumentDescription     = "If true, will force the creation of the new branch even if it already exists."
	CreateDoltBranchFromHeadToolDescription                  = "Creates a new branch from current HEAD."
	CreateDoltBranchFromHeadToolCallSuccessFormatString      = "successfully created branch: %s"
	CreateDoltBranchFromHeadToolSQLQueryFormatString             = "CALL DOLT_BRANCH('%s');"
	CreateDoltBranchFromHeadToolForceSQLQueryFormatString        = "CALL DOLT_BRANCH('-f', '%s');"
)

func RegisterCreateDoltBranchFromHeadTool(server pkg.Server) {
	mcpServer := server.MCP()

	createDoltBranchFromHeadTool := mcp.NewTool(
		CreateDoltBranchFromHeadToolName,
		mcp.WithDescription(CreateDoltBranchFromHeadToolDescription),
		mcp.WithString(
			WorkingBranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingBranchCallToolArgumentDescription),
		),
		mcp.WithString(
			NewBranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CreateDoltBranchToolNewBranchArgumentDescription),
		),
		mcp.WithBoolean(
			ForceCallToolArgumentName,
			mcp.Description(CreateDoltBranchFromHeadToolForceArgumentDescription),
		),
	)

	mcpServer.AddTool(createDoltBranchFromHeadTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var workingBranch string
		workingBranch, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingBranchCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var newBranch string
		newBranch, err = GetRequiredStringArgumentFromCallToolRequest(request, NewBranchCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		force := GetBooleanArgumentFromCallToolRequest(request, ForceCallToolArgumentName)

		config := server.DBConfig()
		config.Branch = workingBranch
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

		if force {
			err = tx.ExecContext(ctx, fmt.Sprintf(CreateDoltBranchFromHeadToolForceSQLQueryFormatString, newBranch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, fmt.Sprintf(CreateDoltBranchFromHeadToolSQLQueryFormatString, newBranch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(fmt.Sprintf(CreateDoltBranchFromHeadToolCallSuccessFormatString, newBranch))
		return
	})
}

