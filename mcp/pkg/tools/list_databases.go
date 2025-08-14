package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListDatabasesToolName     = "list_databases"
	ListDatabasesToolSQLQuery = "SHOW DATABASES;"
	ListDatabasesToolDescription = "Lists all databases in the Dolt server."
)

func NewListDatabasesTool() mcp.Tool {
    return mcp.NewTool(
        ListDatabasesToolName,
        mcp.WithDescription(ListDatabasesToolDescription),
        mcp.WithReadOnlyHintAnnotation(true),
        mcp.WithDestructiveHintAnnotation(false),
        mcp.WithIdempotentHintAnnotation(true),
        mcp.WithOpenWorldHintAnnotation(false),
    )
}

func RegisterListDatabasesTool(server pkg.Server) {
    mcpServer := server.MCP()
    listDatabasesTool := NewListDatabasesTool()
	mcpServer.AddTool(listDatabasesTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error

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
		formattedResult, err = tx.QueryContext(ctx, ListDatabasesToolSQLQuery, db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}

