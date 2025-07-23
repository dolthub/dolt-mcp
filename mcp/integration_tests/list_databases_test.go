package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testListDatabasesTool(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolMustExist(s, ctx, client, serverInfo, tools.ListDatabasesToolName)

	listDatabasesCallToolParams := mcp.CallToolParams{
		Name: tools.ListDatabasesToolName,
	}

	listDatabasesCallToolRequest := mcp.CallToolRequest{
		Params: listDatabasesCallToolParams,
	}

	listDatabasesCallToolResult, err := client.CallTool(ctx, listDatabasesCallToolRequest)
	require.NoError(s.t, err)
	require.NotNil(s.t, listDatabasesCallToolResult)
	require.False(s.t, listDatabasesCallToolResult.IsError)
	require.NotEmpty(s.t, listDatabasesCallToolResult.Content)
}

