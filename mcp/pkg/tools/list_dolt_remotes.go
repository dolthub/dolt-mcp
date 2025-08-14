package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListDoltRemotesToolName        = "list_dolt_remotes"
	ListDoltRemotesToolSQLQuery    = "SELECT * FROM dolt_remotes;"
	ListDoltRemotesToolDescription = "Lists the Dolt server's remotes."
)

func RegisterListDoltRemotesTool(server pkg.Server) {
    mcpServer := server.MCP()

    listDoltRemotesTool := mcp.NewTool(
        ListDoltRemotesToolName,
        mcp.WithDescription(ListDoltRemotesToolDescription),
        mcp.WithReadOnlyHintAnnotation(true),
        mcp.WithDestructiveHintAnnotation(false),
        mcp.WithIdempotentHintAnnotation(true),
        mcp.WithOpenWorldHintAnnotation(false),
    )
    mcpServer.AddTool(listDoltRemotesTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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
		formattedResult, err = tx.QueryContext(ctx, ListDoltRemotesToolSQLQuery, db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}

