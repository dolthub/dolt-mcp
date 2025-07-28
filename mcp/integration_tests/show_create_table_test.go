package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testShowCreateTableToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ShowCreateTableToolName)

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
					Name: tools.ShowCreateTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName: "people",
					},
				},
			},
		},
		{
			description:   "Empty working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ShowCreateTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "",
						tools.TableCallToolArgumentName:         "people",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ShowCreateTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
						tools.TableCallToolArgumentName:         "people",
					},
				},
			},
		},
		{
			description:   "Missing table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ShowCreateTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Empty table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ShowCreateTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:         "",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent table",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ShowCreateTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:         "missing",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		showCreateTableCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, showCreateTableCallToolResult.IsError)
		} else {
			require.False(s.t, showCreateTableCallToolResult.IsError)
		}

		require.NotNil(s.t, showCreateTableCallToolResult)
		require.NotEmpty(s.t, showCreateTableCallToolResult.Content)
	}
}

func testShowCreateTableToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ShowCreateTableToolName)

	showCreateTableCallToolParams := mcp.CallToolParams{
		Name: tools.ShowCreateTableToolName,
		Arguments: map[string]any{
			tools.TableCallToolArgumentName: "people",
			tools.WorkingBranchCallToolArgumentName: testBranchName,
		},
	}

	showCreateTableCallToolRequest := mcp.CallToolRequest{
		Params: showCreateTableCallToolParams,
	}

	showCreateTableCallToolResult, err := client.CallTool(ctx, showCreateTableCallToolRequest)
	require.NoError(s.t, err)
	require.NotNil(s.t, showCreateTableCallToolResult)
	require.False(s.t, showCreateTableCallToolResult.IsError)
	require.NotEmpty(s.t, showCreateTableCallToolResult.Content)
	resultStr, err := resultToString(showCreateTableCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultStr, "people")
	require.Contains(s.t, resultStr, "id")
	require.Contains(s.t, resultStr, "first_name")
	require.Contains(s.t, resultStr, "last_name")
}
