package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	MoveDoltBranchToolName                       = "move_dolt_branch"
	MoveDoltBranchToolOldNameArgumentDescription = "The name of the branch to move/rename."
	MoveDoltBranchToolNewNameArgumentDescription = "The new name of the branch."
	MoveDoltBranchToolForceArgumentDescription   = "If true, will force the original branch to be moved to its new name even if a branch of that name already exists."
	MoveDoltBranchToolDescription                = "Moves/renames a branch from the specified original branch to the provided new name."
	MoveDoltBranchToolCallSuccessFormatString    = "successfully moved branch: %s"
	MoveDoltBranchToolSQLQueryFormatString       = "CALL DOLT_BRANCH('-m', '%s', '%s');"
	MoveDoltBranchToolForceSQLQueryFormatString  = "CALL DOLT_BRANCH('-f', '-m', '%s', '%s');"
)

func NewMoveDoltBranchTool() mcp.Tool {
    return mcp.NewTool(
        MoveDoltBranchToolName,
        mcp.WithDescription(MoveDoltBranchToolDescription),
        mcp.WithReadOnlyHintAnnotation(false),
        mcp.WithDestructiveHintAnnotation(false),
        mcp.WithIdempotentHintAnnotation(false),
        mcp.WithOpenWorldHintAnnotation(false),
        mcp.WithString(
            OldNameCallToolArgumentName,
            mcp.Required(),
            mcp.Description(MoveDoltBranchToolOldNameArgumentDescription),
        ),
        mcp.WithString(
            NewNameCallToolArgumentName,
            mcp.Required(),
            mcp.Description(MoveDoltBranchToolNewNameArgumentDescription),
        ),
        mcp.WithBoolean(
            ForceCallToolArgumentName,
            mcp.Description(MoveDoltBranchToolForceArgumentDescription),
        ),
    )
}

func RegisterMoveDoltBranchTool(server pkg.Server) {
    mcpServer := server.MCP()
    moveDoltBranchTool := NewMoveDoltBranchTool()

	mcpServer.AddTool(moveDoltBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var oldName string
		oldName, err = GetRequiredStringArgumentFromCallToolRequest(request, OldNameCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var newName string
		newName, err = GetRequiredStringArgumentFromCallToolRequest(request, NewNameCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		force := GetBooleanArgumentFromCallToolRequest(request, ForceCallToolArgumentName)

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

		if force {
			err = tx.ExecContext(ctx, fmt.Sprintf(MoveDoltBranchToolForceSQLQueryFormatString, oldName, newName))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, fmt.Sprintf(MoveDoltBranchToolSQLQueryFormatString, oldName, newName))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(fmt.Sprintf(MoveDoltBranchToolCallSuccessFormatString, newName))
		return
	})
}

