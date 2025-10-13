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
	QueryToolName                     = "query"
	QueryToolQueryArgumentDescription = "The query to run."
	QueryToolDescription              = "Executes a READ query."
)

var ErrInvalidSQLReadQuery = errors.New("invalid read query")

func ValidateReadQuery(query string) error {
	sqlStatement, err := ParseSQLQuery(query)
	if err != nil {
		return err
	}

	switch sqlStatement.(type) {
	case *sqlparser.Select, *sqlparser.OtherRead:
		// TODO: make sure we're covering all valid reads here
		return nil
	}

	return ErrInvalidSQLReadQuery
}

func NewQueryTool() mcp.Tool {
	return mcp.NewTool(
		QueryToolName,
		mcp.WithDescription(QueryToolDescription),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString(
			WorkingBranchCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingBranchCallToolArgumentDescription),
		),
		mcp.WithString(
			WorkingDatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(WorkingDatabaseCallToolArgumentDescription),
		),
		mcp.WithString(
			QueryCallToolArgumentName,
			mcp.Required(),
			mcp.Description(QueryToolQueryArgumentDescription),
		),
	)
}

func RegisterQueryTool(server pkg.Server) {
	mcpServer := server.MCP()
	queryTool := NewQueryTool()

	mcpServer.AddTool(queryTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		err = ValidateReadQuery(query)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		config := server.DBConfig()

		var tx db.DatabaseTransaction
		tx, err = NewDatabaseTransactionUsingDatabaseOnBranch(ctx, config, workingDatabase, workingBranch)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		defer func() {
			tx.Rollback(ctx)
		}()

		var formattedResult string
		formattedResult, err = tx.QueryContext(ctx, query, db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}
