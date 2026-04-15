package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CloneDatabaseToolName                              = "clone_database"
	CloneDatabaseToolRemoteURLArgumentDescription      = "The url of the remote database to clone."
	CloneDatabaseToolNameArgumentDescription           = "The local name of the cloned database."
	CloneDatabaseToolDescription                       = "Clones a database from the specified remote URL."
	CloneDatabaseToolCallSuccessFormatString           = "successfully cloned database: %s"
)

func NewCloneDatabaseTool() mcp.Tool {
	return mcp.NewTool(
		CloneDatabaseToolName,
		mcp.WithDescription(CloneDatabaseToolDescription),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithString(
			RemoteURLCallToolArgumentName,
			mcp.Required(),
			mcp.Description(CloneDatabaseToolRemoteURLArgumentDescription),
		),
		mcp.WithString(
			NameCallToolArgumentName,
			mcp.Description(CloneDatabaseToolNameArgumentDescription),
		),
	)
}

func RegisterCloneDatabaseTool(server pkg.Server) {
	mcpServer := server.MCP()
	cloneDatabaseTool := NewCloneDatabaseTool()

	mcpServer.AddTool(cloneDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var url string
		url, err = GetRequiredStringArgumentFromCallToolRequest(request, RemoteURLCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		localName := GetStringArgumentFromCallToolRequest(request, NameCallToolArgumentName)

		dialect := server.Dialect()
		config := server.DBConfig()
		var tx db.DatabaseTransaction
		tx, err = db.NewDatabaseTransaction(ctx, config)
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

		if localName != "" {
			err = tx.ExecContext(ctx, dialect.CallProcedure(db.DoltClone, url, localName))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		} else {
			err = tx.ExecContext(ctx, dialect.CallProcedure(db.DoltClone, url))
			if err != nil {
				result = mcp.NewToolResultError(err.Error())
				return
			}
		}

		result = mcp.NewToolResultText(fmt.Sprintf(CloneDatabaseToolCallSuccessFormatString, url))
		return
	})
}
