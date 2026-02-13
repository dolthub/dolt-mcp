package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	KillProcessToolName        = "kill_process"
	KillProcessToolDescription = "Kill a running process by id (KILL [QUERY] <id>)."

	KillProcessIDArgName    = "process_id"
	KillProcessIDArgDesc    = "The process id to kill (from SHOW PROCESSLIST). Must be a positive integer."
	KillProcessQueryArgName = "kill_query"
	KillProcessQueryArgDesc = "If true, uses KILL QUERY <id> instead of KILL <id>."

	KillProcessToolSuccessMessage = "successfully killed process"
)

func getRequiredPositiveProcessIDFromCallToolRequest(request mcp.CallToolRequest) (int64, error) {
	id, err := request.RequireInt(KillProcessIDArgName)
	if err != nil || id <= 0 {
		return 0, status.Errorf(codes.InvalidArgument, "%s must be a positive integer", KillProcessIDArgName)
	}
	return int64(id), nil
}

func killProcessSQL(id int64, killQuery bool) string {
	if killQuery {
		return fmt.Sprintf("KILL QUERY %d;", id)
	}
	return fmt.Sprintf("KILL %d;", id)
}

func NewKillProcessTool() mcp.Tool {
	return mcp.NewTool(
		KillProcessToolName,
		mcp.WithDescription(KillProcessToolDescription),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
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
		mcp.WithNumber(
			KillProcessIDArgName,
			mcp.Required(),
			mcp.Description(KillProcessIDArgDesc),
		),
		mcp.WithBoolean(
			KillProcessQueryArgName,
			mcp.Description(KillProcessQueryArgDesc),
		),
	)
}

func RegisterKillProcessTool(server pkg.Server) {
	mcpServer := server.MCP()
	tool := NewKillProcessTool()

	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error

		workingBranch, err := GetRequiredStringArgumentFromCallToolRequest(request, WorkingBranchCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		workingDatabase, err := GetRequiredStringArgumentFromCallToolRequest(request, WorkingDatabaseCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		processID, err := getRequiredPositiveProcessIDFromCallToolRequest(request)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		killQuery := GetBooleanArgumentFromCallToolRequest(request, KillProcessQueryArgName)

		config := server.DBConfig()

		var tx db.DatabaseTransaction
		tx, err = NewDatabaseTransactionUsingDatabaseOnBranch(ctx, config, workingDatabase, workingBranch)
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

		err = tx.ExecContext(ctx, killProcessSQL(processID, killQuery))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(KillProcessToolSuccessMessage)
		return
	})
}
