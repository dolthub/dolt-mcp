package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testListDoltBranchesTool(s *testSuite, _ string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltBranchesToolName)

	listDoltBranchesCallToolParams := mcp.CallToolParams{
		Name: tools.ListDoltBranchesToolName,
	}

	listDoltBranchesCallToolRequest := mcp.CallToolRequest{
		Params: listDoltBranchesCallToolParams,
	}

	listDoltBranchesCallToolResult, err := client.CallTool(ctx, listDoltBranchesCallToolRequest)
	require.NoError(s.t, err)
	require.NotNil(s.t, listDoltBranchesCallToolResult)
	require.False(s.t, listDoltBranchesCallToolResult.IsError)
	require.NotEmpty(s.t, listDoltBranchesCallToolResult.Content)
	resultStr, err := resultToString(listDoltBranchesCallToolResult)
	require.NoError(s.t, err)
	require.NotEmpty(s.t, resultStr)
}

