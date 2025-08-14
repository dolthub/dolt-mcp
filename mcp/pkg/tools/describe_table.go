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

func NewDescribeTableTool() mcp.Tool {
    return mcp.NewTool(
        DescribeTableToolName,
        mcp.WithDescription(DescribeTableToolDescription),
        mcp.WithReadOnlyHintAnnotation(true),
        mcp.WithDestructiveHintAnnotation(false),
        mcp.WithIdempotentHintAnnotation(true),
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
            TableCallToolArgumentName,
            mcp.Required(),
            mcp.Description(DescribeTableToolTableArgumentDescription),
        ),
    )
}

func RegisterDescribeTableTool(server pkg.Server) {
    mcpServer := server.MCP()
    describeTableTool := NewDescribeTableTool()

	mcpServer.AddTool(describeTableTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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

		var tableToDescribe string
		tableToDescribe, err = GetRequiredStringArgumentFromCallToolRequest(request, TableCallToolArgumentName)
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

