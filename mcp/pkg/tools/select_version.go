package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	SelectVersionToolName        = "select_version"
	SelectVersionToolSQLQuery    = "SELECT DOLT_VERSION();"
	SelectVersionToolDescription = "Displays the version of the Dolt server."
)

func NewSelectVersionTool() mcp.Tool {
	return mcp.NewTool(
		SelectVersionToolName,
		mcp.WithDescription(SelectVersionToolDescription),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
	)
}

func RegisterSelectVersionTool(server pkg.Server) {
	mcpServer := server.MCP()
	selectVersionTool := NewSelectVersionTool()
	mcpServer.AddTool(selectVersionTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
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
		formattedResult, err = tx.QueryContext(ctx, SelectVersionToolSQLQuery, db.ResultFormatMarkdown)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(formattedResult)
		return
	})
}
