package integration_tests

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDoltFetchBranchSetupSQL = "CALL DOLT_REMOTE('add', 'origin', 'http://localhost:2222/test');"
var testDoltFetchBranchTeardownSQL = `CALL DOLT_REMOTE('remove', 'origin');
CALL DOLT_BRANCH('-D', 'fetchme');
`

func testDoltFetchBranchToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltFetchBranchToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName: "origin",
						tools.BranchCallToolArgumentName:     "fetchme",
					},
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:      "origin",
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.BranchCallToolArgumentName:          "fetchme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:      "origin",
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.BranchCallToolArgumentName:          "fetchme",
					},
				},
			},
		},
		{
			description:   "Missing remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:          "fetchme",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:      "",
						tools.BranchCallToolArgumentName:          "fetchme",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Nonexistent remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:      "doesnotexist",
						tools.BranchCallToolArgumentName:          "fetchme",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Missing branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:      "origin",
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
					Name: tools.DoltFetchBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:      "origin",
						tools.BranchCallToolArgumentName:          "",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Nonexistent branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:      "origin",
						tools.BranchCallToolArgumentName:          "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		doltFetchBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, doltFetchBranchCallToolResult.IsError)
		} else {
			require.False(s.t, doltFetchBranchCallToolResult.IsError)
		}

		require.NotNil(s.t, doltFetchBranchCallToolResult)
		require.NotEmpty(s.t, doltFetchBranchCallToolResult.Content)
	}
}

func testDoltFetchBranchToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	setupRemoteDatabaseSQL := `CREATE TABLE t1 (pk int PRIMARY KEY);
INSERT INTO t1 VALUES (1);
CALL DOLT_COMMIT('-Am', 'add t1 with value 1');
SELECT ACTIVE_BRANCH() INTO @current_branch;
CALL DOLT_BRANCH('-c', @current_branch, 'fetchme');
`

	fileRemoteDatabase := NewFileRemoteDatabase(s, mcpTestDatabaseName)
	err := fileRemoteDatabase.Setup(ctx, setupRemoteDatabaseSQL)
	defer fileRemoteDatabase.Teardown(ctx)

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltFetchBranchToolName)

	doltFetchBranchCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.DoltFetchBranchToolName,
			Arguments: map[string]any{
				tools.RemoteNameCallToolArgumentName:      "origin",
				tools.BranchCallToolArgumentName:          "fetchme",
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	doltFetchBranchCallToolResult, err := client.CallTool(ctx, doltFetchBranchCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, doltFetchBranchCallToolResult.IsError)
	require.NotNil(s.t, doltFetchBranchCallToolResult)
	require.NotEmpty(s.t, doltFetchBranchCallToolResult.Content)
	resultString, err := resultToString(doltFetchBranchCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully fetched branch")

	_, err = s.testDb.ExecContext(ctx, "CALL DOLT_CHECKOUT('fetchme');")
	require.NoError(s.t, err)

	_, err = s.testDb.ExecContext(ctx, fmt.Sprintf("CALL DOLT_CHECKOUT('%s');", testBranchName))
	require.NoError(s.t, err)
}
