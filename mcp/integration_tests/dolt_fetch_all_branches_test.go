package integration_tests

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDoltFetchAllBranchesSetupSQL = "CALL DOLT_REMOTE('add', 'origin', 'http://localhost:2222/test');"

var testDoltFetchAllBranchesTeardownSQL = `CALL DOLT_REMOTE('remove', 'origin');
CALL DOLT_BRANCH('-D', 'fetchme');
CALL DOLT_BRANCH('-D', 'fetchme2');
CALL DOLT_BRANCH('-D', 'fetchme3');
`

func testDoltFetchAllBranchesToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltFetchAllBranchesToolName)

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
					Name: tools.DoltFetchAllBranchesToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName: "origin",
					},
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchAllBranchesToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:      "origin",
						tools.WorkingDatabaseCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchAllBranchesToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:      "origin",
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
					},
				},
			},
		},
		{
			description:   "Missing remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltFetchAllBranchesToolName,
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
					Name: tools.DoltFetchAllBranchesToolName,
					Arguments: map[string]any{
						tools.RemoteURLCallToolArgumentName:       "",
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
					Name: tools.DoltFetchAllBranchesToolName,
					Arguments: map[string]any{
						tools.RemoteURLCallToolArgumentName:       "doesnotexist",
						tools.BranchCallToolArgumentName:          "fetchme",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		doltFetchAllBranchesCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, doltFetchAllBranchesCallToolResult.IsError)
		} else {
			require.False(s.t, doltFetchAllBranchesCallToolResult.IsError)
		}

		require.NotNil(s.t, doltFetchAllBranchesCallToolResult)
		require.NotEmpty(s.t, doltFetchAllBranchesCallToolResult.Content)
	}
}

func testDoltFetchAllBranchesToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	setupRemoteDatabaseSQL := `CREATE TABLE t1 (pk int PRIMARY KEY);
INSERT INTO t1 VALUES (1);
CALL DOLT_COMMIT('-Am', 'add t1 with value 1');
SELECT ACTIVE_BRANCH() INTO @current_branch;
CALL DOLT_BRANCH('-c', @current_branch, 'fetchme');
CALL DOLT_BRANCH('-c', @current_branch, 'fetchme2');
CALL DOLT_BRANCH('-c', @current_branch, 'fetchme3');
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

	requireToolExists(s, ctx, client, serverInfo, tools.DoltFetchAllBranchesToolName)

	doltFetchAllBranchesCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.DoltFetchAllBranchesToolName,
			Arguments: map[string]any{
				tools.RemoteNameCallToolArgumentName:      "origin",
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	doltFetchAllBranchesCallToolResult, err := client.CallTool(ctx, doltFetchAllBranchesCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, doltFetchAllBranchesCallToolResult.IsError)
	require.NotNil(s.t, doltFetchAllBranchesCallToolResult)
	require.NotEmpty(s.t, doltFetchAllBranchesCallToolResult.Content)
	resultString, err := resultToString(doltFetchAllBranchesCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully fetched branch")

	// checkout the remote branches to show theyve been fetched
	_, err = s.testDb.ExecContext(ctx, "CALL DOLT_CHECKOUT('fetchme');")
	require.NoError(s.t, err)
	_, err = s.testDb.ExecContext(ctx, "CALL DOLT_CHECKOUT('fetchme2');")
	require.NoError(s.t, err)
	_, err = s.testDb.ExecContext(ctx, "CALL DOLT_CHECKOUT('fetchme3');")
	require.NoError(s.t, err)

	// return to test branch before teardown
	_, err = s.testDb.ExecContext(ctx, fmt.Sprintf("CALL DOLT_CHECKOUT('%s');", testBranchName))
	require.NoError(s.t, err)
}
