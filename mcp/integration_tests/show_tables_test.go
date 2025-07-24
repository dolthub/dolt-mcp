package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testShowTablesTool(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ShowTablesToolName)

	showTablesCallToolParams := mcp.CallToolParams{
		Name: tools.ShowTablesToolName,
	}

	showTablesCallToolRequest := mcp.CallToolRequest{
		Params: showTablesCallToolParams,
	}

	showTablesCallToolResult, err := client.CallTool(ctx, showTablesCallToolRequest)
	require.NoError(s.t, err)
	require.NotNil(s.t, showTablesCallToolResult)
	require.False(s.t, showTablesCallToolResult.IsError)
	require.NotEmpty(s.t, showTablesCallToolResult.Content)
	resultStr, err := resultToString(showTablesCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultStr, "people")
}

