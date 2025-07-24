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

	mcpServer.AddTool(describeTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tableToDescribe, err := GetRequiredStringArgumentFromCallToolRequest(request, TableCallToolArgumentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		result, err := database.QueryContext(ctx, fmt.Sprintf(DescribeTableToolSQLQueryFormatString, tableToDescribe), db.ResultFormatMarkdown)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(result), nil
	})
}

