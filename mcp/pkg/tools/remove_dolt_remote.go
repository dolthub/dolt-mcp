package tools

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	RemoveDoltRemoteToolName                 = "remove_dolt_remote"
	RemoveDoltRemoteToolRemoteNameArgumentDescription = "The name of the remote to remove."
	RemoveDoltRemoteToolSQLQueryFormatString    = "CALL DOLT_REMOTE('remove', '%s');"
	RemoveDoltRemoteToolDescription             = "Removes a remote from the Dolt server."
	RemoveDoltRemoteToolCallSuccessFormatString = "successfully removed remote: %s"
)

func NewRemoveDoltRemoteTool() mcp.Tool {
    return mcp.NewTool(
        RemoveDoltRemoteToolName,
        mcp.WithDescription(RemoveDoltRemoteToolDescription),
        mcp.WithReadOnlyHintAnnotation(false),
        mcp.WithDestructiveHintAnnotation(false),
        mcp.WithIdempotentHintAnnotation(true),
        mcp.WithOpenWorldHintAnnotation(false),
        mcp.WithString(
            RemoteNameCallToolArgumentName,
            mcp.Required(),
            mcp.Description(RemoveDoltRemoteToolRemoteNameArgumentDescription),
        ),
    )
}

func RegisterRemoveDoltRemoteTool(server pkg.Server) {
    mcpServer := server.MCP()
    removeDoltRemoteTool := NewRemoveDoltRemoteTool()

	mcpServer.AddTool(removeDoltRemoteTool, func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, serverErr error) {
		var err error
		var name string
		name, err = GetRequiredStringArgumentFromCallToolRequest(request, RemoteNameCallToolArgumentName)
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

		err = tx.ExecContext(ctx, fmt.Sprintf(RemoveDoltRemoteToolSQLQueryFormatString, name))
		if err != nil {
			result = mcp.NewToolResultError(err.Error())
			return
		}

		result = mcp.NewToolResultText(fmt.Sprintf(RemoveDoltRemoteToolCallSuccessFormatString, name))
		return
	})
}
