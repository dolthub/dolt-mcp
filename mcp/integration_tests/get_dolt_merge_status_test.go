package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testGetDoltMergeStatusToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.GetDoltMergeStatusToolName)

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
					Name: tools.GetDoltMergeStatusToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.GetDoltMergeStatusToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.GetDoltMergeStatusToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.GetDoltMergeStatusToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.GetDoltMergeStatusToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.GetDoltMergeStatusToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		getDoltMergeStatusCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, getDoltMergeStatusCallToolResult.IsError)
		} else {
			require.False(s.t, getDoltMergeStatusCallToolResult.IsError)
		}

		require.NotNil(s.t, getDoltMergeStatusCallToolResult)
		require.NotEmpty(s.t, getDoltMergeStatusCallToolResult.Content)
	}
}

func testGetDoltMergeStatusToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.GetDoltMergeStatusToolName)

	getDoltMergeStatusCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.GetDoltMergeStatusToolName,
			Arguments: map[string]any{
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	getDoltMergeStatusCallToolResult, err := client.CallTool(ctx, getDoltMergeStatusCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, getDoltMergeStatusCallToolResult.IsError)
	require.NotNil(s.t, getDoltMergeStatusCallToolResult)
	require.NotEmpty(s.t, getDoltMergeStatusCallToolResult.Content)
	resultString, err := resultToString(getDoltMergeStatusCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "is_merging")
	require.Contains(s.t, resultString, "0")
}
