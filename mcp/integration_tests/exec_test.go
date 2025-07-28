package integration_tests

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testExecToolInvalidArguments(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ExecToolName)

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
					Name: tools.ExecToolName,
				},
			},
		},
		{
			description:   "Empty query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ExecToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Invalid SQL WRITE query",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ExecToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: "this is not sql",
					},
				},
			},
		},
	}

	for _, request := range requests {
		execCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, execCallToolResult.IsError)
		} else {
			require.False(s.t, execCallToolResult.IsError)
		}

		require.NotNil(s.t, execCallToolResult)
		require.NotEmpty(s.t, execCallToolResult.Content)
	}
}

func testExecToolSuccess(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ExecToolName)

	execToolCallRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.ExecToolName,
			Arguments: map[string]any{
				tools.QueryCallToolArgumentName: "INSERT INTO people (id, first_name, last_name) VALUES (UUID(), 'homer', 'simpson');",
			},
		},
	}

	execCallToolResult, err := client.CallTool(ctx, execToolCallRequest)
	require.NoError(s.t, err)
	resultStr, err := resultToString(execCallToolResult)
	require.NoError(s.t, err)
	fmt.Println("DUSTIN:", resultStr)
	require.False(s.t, execCallToolResult.IsError)
	require.NotNil(s.t, execCallToolResult)
	require.NotEmpty(s.t, execCallToolResult.Content)
	// resultStr, err := resultToString(execCallToolResult)
	// require.NoError(s.t, err)
	require.Contains(s.t, resultStr, "tim")
	require.Contains(s.t, resultStr, "aaron")
	require.Contains(s.t, resultStr, "brian")
}

