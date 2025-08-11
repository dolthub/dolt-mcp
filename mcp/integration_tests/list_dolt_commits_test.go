package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testListDoltCommitsSetupSQL = `CREATE TABLE ` + "`" + `t1` + "`" + ` (pk int primary key);
CALL DOLT_COMMIT('-Am', 'add t1');
INSERT INTO ` + "`" + `t1` + "`" + ` VALUES (1);
CALL DOLT_COMMIT('-Am', 'insert 1');
INSERT INTO ` + "`" + `t1` + "`" + ` VALUES (2);
CALL DOLT_COMMIT('-Am', 'insert 2');
`

func testListDoltCommitsToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltCommitsToolName)

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
					Name: tools.ListDoltCommitsToolName,
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
					Name: tools.ListDoltCommitsToolName,
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
					Name: tools.ListDoltCommitsToolName,
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
					Name: tools.ListDoltCommitsToolName,
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
					Name: tools.ListDoltCommitsToolName,
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
					Name: tools.ListDoltCommitsToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		listDoltCommitsCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, listDoltCommitsCallToolResult.IsError)
		} else {
			require.False(s.t, listDoltCommitsCallToolResult.IsError)
		}

		require.NotNil(s.t, listDoltCommitsCallToolResult)
		require.NotEmpty(s.t, listDoltCommitsCallToolResult.Content)
	}
}

func testListDoltCommitsToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltCommitsToolName)

	listDoltCommitsCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.ListDoltCommitsToolName,
			Arguments: map[string]any{
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	listDoltCommitsCallToolResult, err := client.CallTool(ctx, listDoltCommitsCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, listDoltCommitsCallToolResult.IsError)
	require.NotNil(s.t, listDoltCommitsCallToolResult)
	require.NotEmpty(s.t, listDoltCommitsCallToolResult.Content)
	resultString, err := resultToString(listDoltCommitsCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "add t1")
	require.Contains(s.t, resultString, "insert 1")
	require.Contains(s.t, resultString, "insert 2")
}
