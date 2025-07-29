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
	CreateTableToolName                     = "create_table"
	CreateTableToolQueryArgumentDescription = "The CREATE TABLE statement to run."
	CreateTableToolDescription              = "Creates a table."
	CreateTableToolCallSuccessMessage       = "successfully created table"
)

var ErrInvalidCreateTableSQLQuery = errors.New("invalid create table statement")

func ValidateCreateTableQuery(query string) error {
	sqlStatement, err := ParseSQLQuery(query)
	if err != nil {
		return err
	}

	switch sqlStatement.(type) {
	case *sqlparser.DDL:
		// TODO: do more to determine if this is truly a create table statement
		return nil
	}

	return ErrInvalidCreateTableSQLQuery
}

func RegisterCreateTableTool(server pkg.Server) {
	mcpServer := server.MCP()

	createTableTool := mcp.NewTool(
		CreateTableToolName,
		mcp.WithDescription(CreateTableToolDescription),
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
			mcp.Description(CreateTableToolQueryArgumentDescription),
		),
	)

	mcpServer.AddTool(createTableTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		err = tx.ExecContext(ctx, createTableStatement)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(CreateTableToolCallSuccessMessage)
		return
	})
}
