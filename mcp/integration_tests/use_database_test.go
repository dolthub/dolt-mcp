package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testUseDatabaseToolInvalidArguments(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolMustExist(s, ctx, client, serverInfo, tools.UseDatabaseToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.UseDatabaseToolName,
				},
			},
		},
		{
			description:   "Empty database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.UseDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Non-existent database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.UseDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName: "foo",
					},
				},
			},
		},
	}

	for _, request := range requests {
		useDatabaseCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, useDatabaseCallToolResult.IsError)
		} else {
			require.False(s.t, useDatabaseCallToolResult.IsError)
		}

		require.NotNil(s.t, useDatabaseCallToolResult)
		require.NotEmpty(s.t, useDatabaseCallToolResult.Content)
	}
}

func testUseDatabaseToolSuccess(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolMustExist(s, ctx, client, serverInfo, tools.UseDatabaseToolName)

	useDatabaseToolCallRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.UseDatabaseToolName,
			Arguments: map[string]any{
				tools.DatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	useDatabaseCallToolResult, err := client.CallTool(ctx, useDatabaseToolCallRequest)
	require.NoError(s.t, err)
	require.False(s.t, useDatabaseCallToolResult.IsError)
	require.NotNil(s.t, useDatabaseCallToolResult)
	require.NotEmpty(s.t, useDatabaseCallToolResult.Content)
}

