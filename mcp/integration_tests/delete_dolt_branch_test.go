package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDeleteDoltBranchSetupSQL = `CALL DOLT_BRANCH('-c', 'main', 'deleteme');
CALL DOLT_BRANCH('-c', 'main', 'forcedeleteme');
SELECT ACTIVE_BRANCH() INTO @current_branch;
CALL DOLT_CHECKOUT('forcedeleteme');
INSERT INTO `+ "`" +`people`+ "`" +` VALUES (UUID(), 'mark', 'twain');
CALL DOLT_CHECKOUT(@current_branch);
`

func testDeleteDoltBranchToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DeleteDoltBranchToolName)

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
					Name: tools.DeleteDoltBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName: "deleteme",
					},
				},
			},
		},
		{
			description:   "Empty working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DeleteDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "",
						tools.BranchCallToolArgumentName:        "deleteme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DeleteDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
						tools.BranchCallToolArgumentName:        "deleteme",
					},
				},
			},
		},
		{
			description:   "Missing branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DeleteDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Empty branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DeleteDoltBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:        "",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent branch",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DeleteDoltBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:        "doesnotexist",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		deleteDoltBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, deleteDoltBranchCallToolResult.IsError)
		} else {
			require.False(s.t, deleteDoltBranchCallToolResult.IsError)
		}

		require.NotNil(s.t, deleteDoltBranchCallToolResult)
		require.NotEmpty(s.t, deleteDoltBranchCallToolResult.Content)
	}
}

func testDeleteDoltBranchToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DeleteDoltBranchToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description: "Deletes a branch",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DeleteDoltBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:        "deleteme",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description: "Forces branch deletion that has outstanding changes",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DeleteDoltBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:        "forcedeleteme",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.ForceCallToolArgumentName:         true,
					},
				},
			},
		},
	}

	for _, request := range requests {
		deleteDoltBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)
		require.False(s.t, deleteDoltBranchCallToolResult.IsError)
		require.NotNil(s.t, deleteDoltBranchCallToolResult)
		require.NotEmpty(s.t, deleteDoltBranchCallToolResult.Content)
		resultString, err := resultToString(deleteDoltBranchCallToolResult)
		require.NoError(s.t, err)
		require.Contains(s.t, resultString, "successfully deleted branch")
	}
}
