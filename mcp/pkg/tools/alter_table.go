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
	AlterTableToolName                     = "alter_table"
	AlterTableToolQueryArgumentDescription = "The ALTER TABLE statement to run."
	AlterTableToolDescription              = "Alters a table."
	AlterTableToolCallSuccessMessage       = "successfully altered table"
)

var ErrInvalidAlterTableSQLQuery = errors.New("invalid alter table statement")

func ValidateAlterTableQuery(query string) error {
	sqlStatement, err := ParseSQLQuery(query)
	if err != nil {
		return err
	}

	switch sqlStatement.(type) {
	case *sqlparser.AlterTable:
		return nil
	}

	return ErrInvalidAlterTableSQLQuery
}

func NewAlterTableTool() mcp.Tool {
    return mcp.NewTool(
        AlterTableToolName,
        mcp.WithDescription(AlterTableToolDescription),
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
            mcp.Description(AlterTableToolQueryArgumentDescription),
        ),
    )
}

func RegisterAlterTableTool(server pkg.Server) {
    mcpServer := server.MCP()
    alterTableTool := NewAlterTableTool()

	mcpServer.AddTool(alterTableTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var alterTableStatement string
		alterTableStatement, err = GetRequiredStringArgumentFromCallToolRequest(request, QueryCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		err = ValidateAlterTableQuery(alterTableStatement)
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

		err = tx.ExecContext(ctx, alterTableStatement)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(AlterTableToolCallSuccessMessage)
		return
	})
}
