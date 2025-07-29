package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testMoveDoltBranchSetupSQL = `SELECT ACTIVE_BRANCH() INTO @current_branch;
CALL DOLT_BRANCH('-c', @current_branch, 'moveme');
CALL DOLT_BRANCH('-c', @current_branch, 'forcemoveme');
`
var testMoveDoltBranchTeardownSQL = `CALL DOLT_BRANCH('-D', 'imoved');
CALL DOLT_BRANCH('-D', 'iforcemoved');`

func testMoveDoltBranchToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.MoveDoltBranchToolName)

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
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.OldNameCallToolArgumentName: "moveme",
						tools.NewNameCallToolArgumentName: "imoved",
					},
				},
			},
		},
		{
			description:   "Empty working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "",
						tools.NewNameCallToolArgumentName:       "imoved",
						tools.OldNameCallToolArgumentName:       "moveme",
					},
				},
			},
		},
		{
			description:   "Missing old_name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.NewNameCallToolArgumentName:       "imoved",
					},
				},
			},
		},
		{
			description:   "Empty old_name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.OldNameCallToolArgumentName:       "",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.NewNameCallToolArgumentName:       "imoved",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
						tools.NewNameCallToolArgumentName:       "imoved",
						tools.OldNameCallToolArgumentName:       testBranchName,
					},
				},
			},
		},
		{
			description:   "Missing new_name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.OldNameCallToolArgumentName:       "moveme",
					},
				},
			},
		},
		{
			description:   "Empty new_name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.NewNameCallToolArgumentName:       "",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.OldNameCallToolArgumentName:       "moveme",
					},
				},
			},
		},
		{
			description:   "Existing branch",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.NewNameCallToolArgumentName:       testBranchName,
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.OldNameCallToolArgumentName:       "moveme",
					},
				},
			},
		},
	}

	for _, request := range requests {
		moveDoltBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, moveDoltBranchCallToolResult.IsError)
		} else {
			require.False(s.t, moveDoltBranchCallToolResult.IsError)
		}

		require.NotNil(s.t, moveDoltBranchCallToolResult)
		require.NotEmpty(s.t, moveDoltBranchCallToolResult.Content)
	}
}

func testMoveDoltBranchToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.MoveDoltBranchToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description: "Renames a branch",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.NewNameCallToolArgumentName:       "imoved",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.OldNameCallToolArgumentName:       "moveme",
					},
				},
			},
		},
		{
			description: "Forces branch rename even if branch exists",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MoveDoltBranchToolName,
					Arguments: map[string]any{
						tools.NewNameCallToolArgumentName:       "iforcemoved",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.OldNameCallToolArgumentName:       "forcemoveme",
						tools.ForceCallToolArgumentName:         true,
					},
				},
			},
		},
	}

	for _, request := range requests {
		moveDoltBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)
		require.False(s.t, moveDoltBranchCallToolResult.IsError)
		require.NotNil(s.t, moveDoltBranchCallToolResult)
		require.NotEmpty(s.t, moveDoltBranchCallToolResult.Content)
		resultString, err := resultToString(moveDoltBranchCallToolResult)
		require.NoError(s.t, err)
		require.Contains(s.t, resultString, "successfully moved branch")
	}
}
