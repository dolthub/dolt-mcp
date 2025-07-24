package tools

import (
	"context"
	"errors"

	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/dolthub/dolt-mcp/mcp/pkg"
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
			QueryCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CreateTableToolQueryArgumentDescription),
		),
	)

	mcpServer.AddTool(createTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		createTableStatement, err := GetRequiredStringArgumentFromCallToolRequest(request, QueryCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		err = ValidateCreateTableQuery(createTableStatement)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		err = database.ExecContext(ctx, createTableStatement)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(CreateTableToolCallSuccessMessage), nil
	})
}

