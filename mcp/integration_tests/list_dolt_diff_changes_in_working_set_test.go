package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testListDoltDiffChangesInWorkingSetSetupSQL = `CREATE TABLE ` + "`" + `t1` + "`" + ` (pk int primary key);
CALL DOLT_COMMIT('-Am', 'add t1');
INSERT INTO ` + "`" + `t1` + "`" + ` VALUES (1);
INSERT INTO ` + "`" + `t1` + "`" + ` VALUES (2);
`

func testListDoltDiffChangesInWorkingSetToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltDiffChangesInWorkingSetToolName)

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
					Name: tools.ListDoltDiffChangesInWorkingSetToolName,
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
					Name: tools.ListDoltDiffChangesInWorkingSetToolName,
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
					Name: tools.ListDoltDiffChangesInWorkingSetToolName,
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
					Name: tools.ListDoltDiffChangesInWorkingSetToolName,
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
					Name: tools.ListDoltDiffChangesInWorkingSetToolName,
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
					Name: tools.ListDoltDiffChangesInWorkingSetToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		listDoltDiffChangesInWorkingSetCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, listDoltDiffChangesInWorkingSetCallToolResult.IsError)
		} else {
			require.False(s.t, listDoltDiffChangesInWorkingSetCallToolResult.IsError)
		}

		require.NotNil(s.t, listDoltDiffChangesInWorkingSetCallToolResult)
		require.NotEmpty(s.t, listDoltDiffChangesInWorkingSetCallToolResult.Content)
	}
}

func testListDoltDiffChangesInWorkingSetToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltDiffChangesInWorkingSetToolName)

	listDoltCommitsCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.ListDoltDiffChangesInWorkingSetToolName,
			Arguments: map[string]any{
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	listDoltDiffChangesInWorkingSetCallToolResult, err := client.CallTool(ctx, listDoltCommitsCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, listDoltDiffChangesInWorkingSetCallToolResult.IsError)
	require.NotNil(s.t, listDoltDiffChangesInWorkingSetCallToolResult)
	require.NotEmpty(s.t, listDoltDiffChangesInWorkingSetCallToolResult.Content)
	resultString, err := resultToString(listDoltDiffChangesInWorkingSetCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "WORKING")
	require.Contains(s.t, resultString, "t1")
}
