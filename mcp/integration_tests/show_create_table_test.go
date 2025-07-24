package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testShowCreateTableTool(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ShowCreateTableToolName)

	showCreateTableCallToolParams := mcp.CallToolParams{
		Name: tools.ShowCreateTableToolName,
		Arguments: map[string]any{
			tools.TableCallToolArgumentName: "people",
		},
	}

	showCreateTableCallToolRequest := mcp.CallToolRequest{
		Params: showCreateTableCallToolParams,
	}

	showCreateTableCallToolResult, err := client.CallTool(ctx, showCreateTableCallToolRequest)
	require.NoError(s.t, err)
	require.NotNil(s.t, showCreateTableCallToolResult)
	require.False(s.t, showCreateTableCallToolResult.IsError)
	require.NotEmpty(s.t, showCreateTableCallToolResult.Content)
	resultStr, err := resultToString(showCreateTableCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultStr, "people")
	require.Contains(s.t, resultStr, "id")
	require.Contains(s.t, resultStr, "first_name")
	require.Contains(s.t, resultStr, "last_name")
}

