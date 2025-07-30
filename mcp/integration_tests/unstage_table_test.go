package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testUnstageTableSetupSQL = `CREATE TABLE ` + "`" + `stageme` + "`" + ` (pk int primary key);
CALL DOLT_ADD('stageme');
`

func testUnstageTableToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.UnstageTableToolName)

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
					Name: tools.UnstageTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:           "stageme",
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
					Name: tools.UnstageTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "",
						tools.TableCallToolArgumentName:           "stageme",
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
					Name: tools.UnstageTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:         "stageme",
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
					Name: tools.UnstageTableToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.TableCallToolArgumentName:           "stageme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.UnstageTableToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.TableCallToolArgumentName:           "stageme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.UnstageTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.TableCallToolArgumentName:           "stageme",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Missing table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.UnstageTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.UnstageTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:           "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Non-existent table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.UnstageTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:           "bar",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		unstageTableCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, unstageTableCallToolResult.IsError)
		} else {
			require.False(s.t, unstageTableCallToolResult.IsError)
		}

		require.NotNil(s.t, unstageTableCallToolResult)
		require.NotEmpty(s.t, unstageTableCallToolResult.Content)
	}
}

func testUnstageTableToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.UnstageTableToolName)

	tableStatuses, err := getDoltStatus(s, ctx, "stageme")
	require.NoError(s.t, err)

	for _, ts := range tableStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.True(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}

	}

	stageTableForDoltCommitCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.UnstageTableToolName,
			Arguments: map[string]any{
				tools.TableCallToolArgumentName:           "stageme",
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	unstageTableCallToolResult, err := client.CallTool(ctx, stageTableForDoltCommitCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, unstageTableCallToolResult.IsError)
	require.NotNil(s.t, unstageTableCallToolResult)
	require.NotEmpty(s.t, unstageTableCallToolResult.Content)
	resultString, err := resultToString(unstageTableCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully unstaged table")

	tableStatuses, err = getDoltStatus(s, ctx, "stageme")
	require.NoError(s.t, err)

	for _, ts := range tableStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.False(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}

	}
}
