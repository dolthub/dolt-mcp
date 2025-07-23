package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	UseDatabaseToolName                 = "use_database"
	UseDatabaseToolDatabaseArgumentDescription = "The name of the database to use."
	UseDatabaseToolSQLQueryFormatString = "use %s;"
	UseDatabaseToolDescription = "Specifies which database in the Dolt server to use."
	UseDatabaseToolCallSuccessFormatString = "now using database: %s"
)

func RegisterUseDatabaseTool(server pkg.Server) {
	mcpServer := server.MCP()

	useDatabaseTool := mcp.NewTool(
		UseDatabaseToolName,
		mcp.WithDescription(UseDatabaseToolDescription),
		mcp.WithString(
			DatabaseCallToolArgumentName,
			mcp.Required(),
			mcp.Description(UseDatabaseToolDatabaseArgumentDescription),
		),
	)

	mcpServer.AddTool(useDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		databaseToUse, err := GetDatabaseArgumentFromCallToolRequest(request) 
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		database := server.DB()
		err = database.ExecContext(ctx, fmt.Sprintf(UseDatabaseToolSQLQueryFormatString, databaseToUse))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(UseDatabaseToolCallSuccessFormatString, databaseToUse)), nil
	})
}

