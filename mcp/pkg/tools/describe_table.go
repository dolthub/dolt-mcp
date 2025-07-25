package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DescribeTableToolName                            = "describe_table"
	DescribeTableToolTableArgumentDescription        = "The name of the table to describe."
	DescribeTableToolSQLQueryFormatString            = "DESCRIBE `%s`;"
	DescribeTableToolDescription                     = "Describes a table in the current database."
)

func RegisterDescribeTableTool(server pkg.Server) {
	mcpServer := server.MCP()

	describeTableTool := mcp.NewTool(
		DescribeTableToolName,
		mcp.WithDescription(DescribeTableToolDescription),
		mcp.WithString(
			TableCallToolArgumentName,
			mcp.Required(),
			mcp.Description(DescribeTableToolTableArgumentDescription),
		),
	)

	mcpServer.AddTool(describeTableTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var tableToDescribe string
		tableToDescribe, err = GetRequiredStringArgumentFromCallToolRequest(request, TableCallToolArgumentName)
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
			tx.Rollback(ctx)
		}()

		var formattedResult string
		formattedResult, err = tx.QueryContext(ctx, fmt.Sprintf(DescribeTableToolSQLQueryFormatString, tableToDescribe), db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}

