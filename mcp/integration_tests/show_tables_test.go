package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testShowTablesToolInvalidArguments(s *testSuite, _ string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ShowTablesToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ShowTablesToolName,
				},
			},
		},
		{
			description:   "Empty working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ShowTablesToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ShowTablesToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
					},
				},
			},
		},
	}

	for _, request := range requests {
		showTablesCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, showTablesCallToolResult.IsError)
		} else {
			require.False(s.t, showTablesCallToolResult.IsError)
		}

		require.NotNil(s.t, showTablesCallToolResult)
		require.NotEmpty(s.t, showTablesCallToolResult.Content)
	}
}

func testShowTablesToolSuccess(s *testSuite, testBranchName string) {
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
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName: testBranchName,
		},
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

