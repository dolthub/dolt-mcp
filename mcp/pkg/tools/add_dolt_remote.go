package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	AddDoltRemoteToolName                          = "add_dolt_remote"
	AddDoltRemoteToolRemoteNameArgumentDescription = "The name of the remote to add."
	AddDoltRemoteToolRemoteURLArgumentDescription  = "The URL of the remote to add."
	AddDoltRemoteToolSQLQueryFormatString          = "CALL DOLT_REMOTE('add', '%s', '%s');"
	AddDoltRemoteToolDescription                   = "Adds a remote to the Dolt server."
	AddDoltRemoteToolCallSuccessFormatString       = "successfully added remote: %s"
)

func NewAddDoltRemoteTool() mcp.Tool {
    return mcp.NewTool(
        AddDoltRemoteToolName,
        mcp.WithDescription(AddDoltRemoteToolDescription),
        mcp.WithReadOnlyHintAnnotation(false),
        mcp.WithDestructiveHintAnnotation(false),
        mcp.WithIdempotentHintAnnotation(true),
        mcp.WithOpenWorldHintAnnotation(false),
        mcp.WithString(
            RemoteNameCallToolArgumentName,
            mcp.Required(),
            mcp.Description(AddDoltRemoteToolRemoteNameArgumentDescription),
        ),
        mcp.WithString(
            RemoteURLCallToolArgumentName,
            mcp.Required(),
            mcp.Description(AddDoltRemoteToolRemoteURLArgumentDescription),
        ),
    )
}

func RegisterAddDoltRemoteTool(server pkg.Server) {
    mcpServer := server.MCP()
    addDoltRemoteTool := NewAddDoltRemoteTool()

	mcpServer.AddTool(addDoltRemoteTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var name string
		name, err = GetRequiredStringArgumentFromCallToolRequest(request, RemoteNameCallToolArgumentName)
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		var url string
		url, err = GetRequiredStringArgumentFromCallToolRequest(request, RemoteURLCallToolArgumentName)
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
			rerr := CommitTransactionOrRollbackOnError(ctx, tx, err)
			if rerr != nil {
				result = mcp.NewToolResultError(rerr.Error())
			}
		}()

		err = tx.ExecContext(ctx, fmt.Sprintf(AddDoltRemoteToolSQLQueryFormatString, name, url))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(AddDoltRemoteToolCallSuccessFormatString, name))
		return
	})
}
