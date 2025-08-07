package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testMergeDoltBranchSetupSQL = `SELECT ACTIVE_BRANCH() INTO @current_branch;
CALL DOLT_BRANCH('-c', @current_branch, 'mergeme');
CALL DOLT_CHECKOUT('mergeme');
INSERT INTO ` + "`" + `people` + "`" + ` VALUES (UUID(), 'mark', 'twain');
CALL DOLT_COMMIT('-Am', 'insert mark twain');
CALL DOLT_CHECKOUT(@current_branch);
`

func testMergeDoltBranchToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.MergeDoltBranchToolName)

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
					Name: tools.MergeDoltBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:          "mergeme",
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
					Name: tools.MergeDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "",
						tools.BranchCallToolArgumentName:          "mergeme",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MergeDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.BranchCallToolArgumentName:          "mergeme",
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
					Name: tools.MergeDoltBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:        "mergeme",
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
					Name: tools.MergeDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.BranchCallToolArgumentName:          "mergeme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MergeDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.BranchCallToolArgumentName:          "mergeme",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
					},
				},
			},
		},
		{
			description:   "Missing branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MergeDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MergeDoltBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:          "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Non-existent branch",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.MergeDoltBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:          "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		mergeDoltBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, mergeDoltBranchCallToolResult.IsError)
		} else {
			require.False(s.t, mergeDoltBranchCallToolResult.IsError)
		}

		require.NotNil(s.t, mergeDoltBranchCallToolResult)
		require.NotEmpty(s.t, mergeDoltBranchCallToolResult.Content)
	}
}

func testMergeDoltBranchToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.MergeDoltBranchToolName)

	mergeDoltBranchCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.MergeDoltBranchToolName,
			Arguments: map[string]any{
				tools.BranchCallToolArgumentName:        "mergeme",
				tools.WorkingBranchCallToolArgumentName: testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	mergeDoltBranchCallToolResult, err := client.CallTool(ctx, mergeDoltBranchCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, mergeDoltBranchCallToolResult.IsError)
	require.NotNil(s.t, mergeDoltBranchCallToolResult)
	require.NotEmpty(s.t, mergeDoltBranchCallToolResult.Content)
	resultString, err := resultToString(mergeDoltBranchCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully merged branch")
}
