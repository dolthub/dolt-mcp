package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testUnstageAllTablesSetupSQL = `CREATE TABLE ` + "`" + `stagemeone` + "`" + ` (pk int primary key);
CREATE TABLE ` + "`" + `stagemetwo` + "`" + ` (pk int primary key);
CALL DOLT_ADD('stagemeone');
CALL DOLT_ADD('stagemetwo');
`

func testUnstageAllTablesToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.UnstageAllTablesToolName)

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
					Name: tools.UnstageAllTablesToolName,
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
					Name: tools.UnstageAllTablesToolName,
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
					Name: tools.UnstageAllTablesToolName,
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
					Name: tools.UnstageAllTablesToolName,
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
					Name: tools.UnstageAllTablesToolName,
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
					Name: tools.UnstageAllTablesToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		unstageAllTablesCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, unstageAllTablesCallToolResult.IsError)
		} else {
			require.False(s.t, unstageAllTablesCallToolResult.IsError)
		}

		require.NotNil(s.t, unstageAllTablesCallToolResult)
		require.NotEmpty(s.t, unstageAllTablesCallToolResult.Content)
	}
}

func testUnstageAllTablesToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.UnstageAllTablesToolName)

	tableOneStatuses, err := getDoltStatus(s, ctx, "stagemeone")
	require.NoError(s.t, err)

	for _, ts := range tableOneStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.True(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}
	}

	tableTwoStatuses, err := getDoltStatus(s, ctx, "stagemetwo")
	require.NoError(s.t, err)

	for _, ts := range tableTwoStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.True(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}
	}

	unstageAllTablesCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.UnstageAllTablesToolName,
			Arguments: map[string]any{
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	unstageAllTablesCallToolResult, err := client.CallTool(ctx, unstageAllTablesCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, unstageAllTablesCallToolResult.IsError)
	require.NotNil(s.t, unstageAllTablesCallToolResult)
	require.NotEmpty(s.t, unstageAllTablesCallToolResult.Content)
	resultString, err := resultToString(unstageAllTablesCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully unstaged tables")

	tableOneStatuses, err = getDoltStatus(s, ctx, "stagemeone")
	require.NoError(s.t, err)

	for _, ts := range tableOneStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.False(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}
	}

	tableTwoStatuses, err = getDoltStatus(s, ctx, "stagemetwo")
	require.NoError(s.t, err)

	for _, ts := range tableTwoStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.False(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}
	}
}
