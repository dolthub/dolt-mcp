package integration_tests

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDoltPullBranchTeardownSQL = "DROP DATABASE alt;"

func testDoltPullBranchToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltPullBranchToolName)

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
					Name: tools.DoltPullBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:     "pullme",
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
					Name: tools.DoltPullBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:     "pullme",
						tools.RemoteNameCallToolArgumentName: "origin",
						tools.WorkingDatabaseCallToolArgumentDescription: "",
					},
				},
			},
		},
		{
			description:   "Nonexistent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPullBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:     "pullme",
						tools.RemoteNameCallToolArgumentName: "origin",
						tools.WorkingDatabaseCallToolArgumentDescription: "doesnotexist",
					},
				},
			},
		},
		{
			description:   "Missing remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPullBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName: "pullme",
					},
				},
			},
		},
		{
			description:   "Empty remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPullBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName: "",
						tools.BranchCallToolArgumentName:     "pullme",
					},
				},
			},
		},
		{
			description:   "Nonexistent remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPullBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName: "foo",
						tools.BranchCallToolArgumentName:     "pullme",
					},
				},
			},
		},
		{
			description:   "Missing branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPullBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName: "origin",
					},
				},
			},
		},
		{
			description:   "Empty branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPullBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName: "origin",
						tools.BranchCallToolArgumentName:     "",
					},
				},
			},
		},
		{
			description:   "Nonexistent branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPullBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:     "foo",
						tools.RemoteNameCallToolArgumentName: "origin",
					},
				},
			},
		},
	}

	for _, request := range requests {
		doltPullBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, doltPullBranchCallToolResult.IsError)
		} else {
			require.False(s.t, doltPullBranchCallToolResult.IsError)
		}

		require.NotNil(s.t, doltPullBranchCallToolResult)
		require.NotEmpty(s.t, doltPullBranchCallToolResult.Content)
	}
}

func testDoltPullBranchToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	remoteSetupSQL := `SELECT ACTIVE_BRANCH() INTO @current_branch;
CALL DOLT_BRANCH('-c', @current_branch, 'pullme');
CALL DOLT_CHECKOUT('pullme');
CREATE TABLE t1 (pk INT PRIMARY KEY);
INSERT INTO t1 VALUES (1);
CALL DOLT_COMMIT('-Am', 'add t1 and 1');
CALL DOLT_CHECKOUT(@current_branch);`

	fileRemoteDatabase := NewFileRemoteDatabase(s, "alt")
	err := fileRemoteDatabase.Setup(ctx, remoteSetupSQL)
	defer fileRemoteDatabase.Teardown(ctx)

	// clone the remote database alt so that the local db shares
	// a common ancestor with the remote
	_, err = s.testDb.ExecContext(ctx, "CALL DOLT_CLONE('http://localhost:2222/alt');")
	require.NoError(s.t, err)

	// add a commit to the remote so it can be pulled
	_, err = fileRemoteDatabase.testDB.ExecContext(ctx, `SELECT ACTIVE_BRANCH() INTO @current_branch;
CALL DOLT_CHECKOUT('pullme');
INSERT INTO t1 VALUES (2);
CALL DOLT_COMMIT('-Am', 'add 2 to t1');
CALL DOLT_CHECKOUT(@current_branch);
`)
	require.NoError(s.t, err)

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltPullBranchToolName)

	doltPullBranchCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.DoltPullBranchToolName,
			Arguments: map[string]any{
				tools.BranchCallToolArgumentName:     "pullme",
				tools.RemoteNameCallToolArgumentName: "origin",
				tools.WorkingDatabaseCallToolArgumentName: "alt",
			},
		},
	}

	doltPullBranchCallToolResult, err := client.CallTool(ctx, doltPullBranchCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, doltPullBranchCallToolResult.IsError)
	require.NotNil(s.t, doltPullBranchCallToolResult)
	require.NotEmpty(s.t, doltPullBranchCallToolResult.Content)
	resultString, err := resultToString(doltPullBranchCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully pulled branch")

	_, err = s.testDb.ExecContext(ctx, "USE alt;")
	require.NoError(s.t, err)

	// checkout the remote branches to show theyve been fetched
	_, err = s.testDb.ExecContext(ctx, "CALL DOLT_CHECKOUT('pullme');")
	require.NoError(s.t, err)

	_, err = s.testDb.ExecContext(ctx, fmt.Sprintf("USE %s;", mcpTestDatabaseName))
	require.NoError(s.t, err)

	// return to test branch before teardown
	_, err = s.testDb.ExecContext(ctx, fmt.Sprintf("CALL DOLT_CHECKOUT('%s');", testBranchName))
	require.NoError(s.t, err)
}
