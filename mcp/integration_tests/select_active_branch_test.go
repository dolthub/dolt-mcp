package integration_tests

import (
	"context"
	"strings"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testSelectActiveBranchTool(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.SelectActiveBranchToolName)

	selectActiveBranchParams := mcp.CallToolParams{
		Name: tools.SelectActiveBranchToolName,
	}

	selectActiveBranchCallToolRequest := mcp.CallToolRequest{
		Params: selectActiveBranchParams,
	}

	selectActiveBranchCallToolResult, err := client.CallTool(ctx, selectActiveBranchCallToolRequest)
	require.NoError(s.t, err)
	require.NotNil(s.t, selectActiveBranchCallToolResult)
	require.False(s.t, selectActiveBranchCallToolResult.IsError)
	require.NotEmpty(s.t, selectActiveBranchCallToolResult.Content)
	resultStr, err := resultToString(selectActiveBranchCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, strings.ToLower(resultStr), "active_branch()")
}

