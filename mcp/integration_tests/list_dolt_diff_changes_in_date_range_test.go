package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testListDoltDiffChangesInDateRangeSetupSQL = `CREATE TABLE ` + "`" + `t1` + "`" + ` (pk int primary key);
INSERT INTO ` + "`" + `t1` + "`" + ` VALUES (1);
CALL DOLT_COMMIT('-Am', 'add t1');
CALL DOLT_COMMIT('--amend', '--date=2022-06-01');
INSERT INTO ` + "`" + `t1` + "`" + ` VALUES (2);
CALL DOLT_ADD('t1');
`

func testListDoltDiffChangesInDateRangeToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltDiffChangesInDateRangeToolName)

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
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.StartDateCallToolArgumentName:       "2021-01-01",
						tools.EndDateCallToolArgumentName:         "2023-01-01",
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
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.StartDateCallToolArgumentName:       "2021-01-01",
						tools.EndDateCallToolArgumentName:         "2023-01-01",
						tools.WorkingBranchCallToolArgumentName:   "",
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
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.StartDateCallToolArgumentName:       "2021-01-01",
						tools.EndDateCallToolArgumentName:         "2023-01-01",
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
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
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.StartDateCallToolArgumentName:     "2021-01-01",
						tools.EndDateCallToolArgumentName:       "2023-01-01",
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
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.StartDateCallToolArgumentName:       "2021-01-01",
						tools.EndDateCallToolArgumentName:         "2023-01-01",
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
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.StartDateCallToolArgumentName:       "2021-01-01",
						tools.EndDateCallToolArgumentName:         "2023-01-01",
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
					},
				},
			},
		},
		{
			description:   "Missing start date argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.EndDateCallToolArgumentName:         "2023-01-01",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty start date argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.StartDateCallToolArgumentName:       "",
						tools.EndDateCallToolArgumentName:         "2023-01-01",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
					},
				},
			},
		},
		{
			description:   "Missing end date argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.StartDateCallToolArgumentName:       "2021-01-01",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty end date argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesInDateRangeToolName,
					Arguments: map[string]any{
						tools.EndDateCallToolArgumentName:         "",
						tools.StartDateCallToolArgumentName:       "2021-01-01",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		listDoltDiffChangesInDateRangeCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, listDoltDiffChangesInDateRangeCallToolResult.IsError)
		} else {
			require.False(s.t, listDoltDiffChangesInDateRangeCallToolResult.IsError)
		}

		require.NotNil(s.t, listDoltDiffChangesInDateRangeCallToolResult)
		require.NotEmpty(s.t, listDoltDiffChangesInDateRangeCallToolResult.Content)
	}
}

func testListDoltDiffChangesInDateRangeToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltDiffChangesInDateRangeToolName)

	listDoltDiffChangesInDateRangeCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.ListDoltDiffChangesInDateRangeToolName,
			Arguments: map[string]any{
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
				tools.StartDateCallToolArgumentName:       "2021-01-01",
				tools.EndDateCallToolArgumentName:         "2023-01-01",
			},
		},
	}

	listDoltDiffChangesInDateRangeCallToolResult, err := client.CallTool(ctx, listDoltDiffChangesInDateRangeCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, listDoltDiffChangesInDateRangeCallToolResult.IsError)
	require.NotNil(s.t, listDoltDiffChangesInDateRangeCallToolResult)
	require.NotEmpty(s.t, listDoltDiffChangesInDateRangeCallToolResult.Content)
	resultString, err := resultToString(listDoltDiffChangesInDateRangeCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "2022-06-01")
}
