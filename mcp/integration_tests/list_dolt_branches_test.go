package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testListDoltBranchesToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltBranchesToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltBranchesToolName,
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltBranchesToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltBranchesToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
					},
				},
			},
		},
	}

	for _, request := range requests {
		listDoltBranchesCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, listDoltBranchesCallToolResult.IsError)
		} else {
			require.False(s.t, listDoltBranchesCallToolResult.IsError)
		}

		require.NotNil(s.t, listDoltBranchesCallToolResult)
		require.NotEmpty(s.t, listDoltBranchesCallToolResult.Content)
	}
}

func testListDoltBranchesToolSuccess(s *testSuite, _ string) {
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
		Arguments: map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
		},
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

