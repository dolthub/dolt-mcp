package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDoltPushBranchSetupSQL = `SELECT ACTIVE_BRANCH() INTO @current_branch;
CALL DOLT_BRANCH('-c', @current_branch, 'pushme');
CALL DOLT_BRANCH('-c', @current_branch, 'forcepushme');
CALL DOLT_CHECKOUT('pushme');
CREATE TABLE t1 (pk INT PRIMARY KEY);
INSERT INTO t1 VALUES (1);
CALL DOLT_COMMIT('-Am', 'add t1 and 1');
CALL DOLT_CHECKOUT('forcepushme');
CREATE TABLE t1 (pk INT PRIMARY KEY);
INSERT INTO t1 VALUES (1);
CALL DOLT_COMMIT('-Am', 'add t1 and 1');
CALL DOLT_REMOTE('add', 'origin', 'http://localhost:2222/test');
CALL DOLT_CHECKOUT(@current_branch);`

var testDoltPushBranchTeardownSQL = `CALL DOLT_BRANCH('-D', 'pushme');
CALL DOLT_BRANCH('-D', 'forcepushme');
CALL DOLT_REMOTE('remove', 'origin');`

func testDoltPushBranchToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltPushBranchToolName)

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
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:     "pushme",
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
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:                 "pushme",
						tools.RemoteNameCallToolArgumentName:             "origin",
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
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:                 "pushme",
						tools.RemoteNameCallToolArgumentName:             "origin",
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
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:                 "pushme",
						tools.WorkingDatabaseCallToolArgumentDescription: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:             "",
						tools.BranchCallToolArgumentName:                 "pushme",
						tools.WorkingDatabaseCallToolArgumentDescription: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Nonexistent remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:             "foo",
						tools.BranchCallToolArgumentName:                 "pushme",
						tools.WorkingDatabaseCallToolArgumentDescription: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Missing branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:             "origin",
						tools.WorkingDatabaseCallToolArgumentDescription: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName:             "origin",
						tools.BranchCallToolArgumentName:                 "",
						tools.WorkingDatabaseCallToolArgumentDescription: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Nonexistent branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:                 "foo",
						tools.RemoteNameCallToolArgumentName:             "origin",
						tools.WorkingDatabaseCallToolArgumentDescription: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		doltPushBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, doltPushBranchCallToolResult.IsError)
		} else {
			require.False(s.t, doltPushBranchCallToolResult.IsError)
		}

		require.NotNil(s.t, doltPushBranchCallToolResult)
		require.NotEmpty(s.t, doltPushBranchCallToolResult.Content)
	}
}

func testDoltPushBranchToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	fileRemoteDatabase := NewFileRemoteDatabase(s, mcpTestDatabaseName)
	err := fileRemoteDatabase.Setup(ctx, "")
	defer fileRemoteDatabase.Teardown(ctx)

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltPushBranchToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description: "Should be able to push branch",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:          "pushme",
						tools.RemoteNameCallToolArgumentName:      "origin",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description: "Should be able to force push a branch",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltPushBranchToolName,
					Arguments: map[string]any{
						tools.BranchCallToolArgumentName:          "forcepushme",
						tools.RemoteNameCallToolArgumentName:      "origin",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.ForceCallToolArgumentName:           true,
					},
				},
			},
		},
	}

	for _, request := range requests {
		doltPushBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)
		require.False(s.t, doltPushBranchCallToolResult.IsError)
		require.NotNil(s.t, doltPushBranchCallToolResult)
		require.NotEmpty(s.t, doltPushBranchCallToolResult.Content)
		resultString, err := resultToString(doltPushBranchCallToolResult)
		require.NoError(s.t, err)
		require.Contains(s.t, resultString, "successfully pushed branch")
	}
}
