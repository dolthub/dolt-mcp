package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateDoltBranchToolName                              = "create_dolt_branch"
	CreateDoltBranchToolOriginalBranchArgumentDescription = "The name of the branch to copy."
	CreateDoltBranchToolNewBranchArgumentDescription      = "The name of the new branch."
	CreateDoltBranchToolForceArgumentDescription          = "If true, will force the creation of the new branch even if it already exists."
	CreateDoltBranchToolDescription                       = "Creates a new branch from the specified original branch."
	CreateDoltBranchToolCallSuccessFormatString = "successfully created branch: %s"
)

func NewCreateDoltBranchTool() mcp.Tool {
	return mcp.NewTool(
		CreateDoltBranchToolName,
		mcp.WithDescription(CreateDoltBranchToolDescription),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString(
			WorkingDatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingDatabaseCallToolArgumentDescription),
		),
		mcp.WithString(
			OriginalBranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CreateDoltBranchToolOriginalBranchArgumentDescription),
		),
		mcp.WithString(
			NewBranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CreateDoltBranchToolNewBranchArgumentDescription),
		),
		mcp.WithBoolean(
			ForceCallToolArgumentName,
			mcp.Description(CreateDoltBranchToolForceArgumentDescription),
		),
	)
}

func RegisterCreateDoltBranchTool(server pkg.Server) {
	mcpServer := server.MCP()
	createDoltBranchTool := NewCreateDoltBranchTool()

	mcpServer.AddTool(createDoltBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error

		var workingDatabase string
		workingDatabase, err = GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var originalBranch string
		originalBranch, err = GetRequiredStringArgumentFromCallToolRequest(request, OriginalBranchCallToolArgumentName)
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

		dialect := server.Dialect()
		config := server.DBConfig()

		var tx db.DatabaseTransaction
		tx, err = NewDatabaseTransactionUsingDatabase(ctx, config, dialect, workingDatabase)
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
			err = tx.ExecContext(ctx, dialect.CallProcedure(db.DoltBranch, "-f", "-c", originalBranch, newBranch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, dialect.CallProcedure(db.DoltBranch, "-c", originalBranch, newBranch))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(fmt.Sprintf(CreateDoltBranchToolCallSuccessFormatString, newBranch))
		return
	})
}
