package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testQueryToolInvalidArguments(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.QueryToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.QueryToolName,
				},
			},
		},
		{
			description:   "Empty query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.QueryToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Invalid SQL READ query",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.QueryToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: "this is not sql",
					},
				},
			},
		},
	}

	for _, request := range requests {
		queryCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, queryCallToolResult.IsError)
		} else {
			require.False(s.t, queryCallToolResult.IsError)
		}

		require.NotNil(s.t, queryCallToolResult)
		require.NotEmpty(s.t, queryCallToolResult.Content)
	}
}

func testQueryToolSuccess(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.QueryToolName)

	queryToolCallRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.QueryToolName,
			Arguments: map[string]any{
				tools.QueryCallToolArgumentName: "SELECT * FROM people;",
			},
		},
	}

	queryCallToolResult, err := client.CallTool(ctx, queryToolCallRequest)
	require.NoError(s.t, err)
	require.False(s.t, queryCallToolResult.IsError)
	require.NotNil(s.t, queryCallToolResult)
	require.NotEmpty(s.t, queryCallToolResult.Content)
	resultStr, err := resultToString(queryCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultStr, "tim")
	require.Contains(s.t, resultStr, "aaron")
	require.Contains(s.t, resultStr, "brian")
}

