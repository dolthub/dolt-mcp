package integration_tests

import (
	"context"
	"strings"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testSelectActiveBranchToolInvalidArguments(s *testSuite, _ string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.SelectActiveBranchToolName)

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
					Name: tools.SelectActiveBranchToolName,
				},
			},
		},
		{
			description:   "Empty working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.SelectActiveBranchToolName,
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
					Name: tools.SelectActiveBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
					},
				},
			},
		},
	}

	for _, request := range requests {
		selectActiveBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, selectActiveBranchCallToolResult.IsError)
		} else {
			require.False(s.t, selectActiveBranchCallToolResult.IsError)
		}

		require.NotNil(s.t, selectActiveBranchCallToolResult)
		require.NotEmpty(s.t, selectActiveBranchCallToolResult.Content)
	}
}

func testSelectActiveBranchToolSuccess(s *testSuite, testBranchName string) {
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
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName: testBranchName,
		},
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
	require.Contains(s.t, resultStr, testBranchName)
}
