package integration_tests

import (
	"context"
	"strings"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testSelectVersionTool(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.SelectVersionToolName)

	selectVersionParams := mcp.CallToolParams{
		Name: tools.SelectVersionToolName,
	}

	selectVersionCallToolRequest := mcp.CallToolRequest{
		Params: selectVersionParams,
	}

	selectVersionCallToolResult, err := client.CallTool(ctx, selectVersionCallToolRequest)
	require.NoError(s.t, err)
	require.NotNil(s.t, selectVersionCallToolResult)
	require.False(s.t, selectVersionCallToolResult.IsError)
	require.NotEmpty(s.t, selectVersionCallToolResult.Content)
	resultStr, err := resultToString(selectVersionCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, strings.ToLower(resultStr), "dolt_version()")
}

