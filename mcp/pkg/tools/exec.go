package tools

import (
	"context"
	"errors"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ExecToolName                     = "exec"
	ExecToolQueryArgumentDescription = "The query to run."
	ExecToolDescription              = "Executes a WRITE query."
	ExecToolCallSuccessMessage       = "successfully executed write"
)

var ErrInvalidSQLWriteQuery = errors.New("invalid write query")

func ValidateWriteQuery(query string) error {
	sqlStatement, err := ParseSQLQuery(query)
	if err != nil {
		return err
	}

	switch sqlStatement.(type) {
	case *sqlparser.Insert, *sqlparser.Update, *sqlparser.Delete:
		// TODO: make sure we're covering all valid writes here
		return nil
	}

	return ErrInvalidSQLWriteQuery
}

func NewExecTool() mcp.Tool {
    return mcp.NewTool(
        ExecToolName,
        mcp.WithDescription(ExecToolDescription),
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
        mcp.WithString(
            QueryCallToolArgumentName,
            mcp.Required(),
            mcp.Description(ExecToolQueryArgumentDescription),
        ),
    )
}

func RegisterExecTool(server pkg.Server) {
    mcpServer := server.MCP()
    execTool := NewExecTool()

	mcpServer.AddTool(execTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var query string
		query, err = GetRequiredStringArgumentFromCallToolRequest(request, QueryCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		err = ValidateWriteQuery(query)
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

		err = tx.ExecContext(ctx, query)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(ExecToolCallSuccessMessage)
		return
	})
}
