package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testCreateDoltCommitSetupSQL = `CREATE TABLE ` + "`" + `commitme` + "`" + ` (pk int primary key);
CALL DOLT_ADD('commitme');
`

func testCreateDoltCommitToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CreateDoltCommitToolName)

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
					Name: tools.CreateDoltCommitToolName,
					Arguments: map[string]any{
						tools.MessageCallToolArgumentName:         "commit table commitme",
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
					Name: tools.CreateDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.MessageCallToolArgumentName:         "commit table commitme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.MessageCallToolArgumentName:         "commit table commitme",
					},
				},
			},
		},
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltCommitToolName,
					Arguments: map[string]any{
						tools.MessageCallToolArgumentName:       "commit table commitme",
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
					Name: tools.CreateDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.MessageCallToolArgumentName:         "commit table commitme",
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
					Name: tools.CreateDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.MessageCallToolArgumentName:         "commit table commitme",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
					},
				},
			},
		},
		{
			description:   "Missing message argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty message argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltCommitToolName,
					Arguments: map[string]any{
						tools.MessageCallToolArgumentName:         "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		createDoltCommitCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, createDoltCommitCallToolResult.IsError)
		} else {
			require.False(s.t, createDoltCommitCallToolResult.IsError)
		}

		require.NotNil(s.t, createDoltCommitCallToolResult)
		require.NotEmpty(s.t, createDoltCommitCallToolResult.Content)
	}
}

func testCreateDoltCommitToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CreateDoltCommitToolName)

	commitMeIsStaged, err := getTableStagedStatus(s, ctx, "commitme", testDoltStatusNewTable)
	require.NoError(s.t, err)
	require.True(s.t, commitMeIsStaged)

	preCommitSha, err := getLastCommitHash(s, ctx)
	require.NoError(s.t, err)

	createDoltCommitCallToolParams := mcp.CallToolParams{
		Name: tools.CreateDoltCommitToolName,
		Arguments: map[string]any{
			tools.MessageCallToolArgumentName:         "commit table commitme",
			tools.WorkingBranchCallToolArgumentName:   testBranchName,
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
		},
	}

	createDoltCommitCallToolRequest := mcp.CallToolRequest{
		Params: createDoltCommitCallToolParams,
	}

	createDoltCommitCallToolResult, err := client.CallTool(ctx, createDoltCommitCallToolRequest)
	require.NoError(s.t, err)
	require.NotNil(s.t, createDoltCommitCallToolResult)
	require.False(s.t, createDoltCommitCallToolResult.IsError)
	require.NotEmpty(s.t, createDoltCommitCallToolResult.Content)
	resultStr, err := resultToString(createDoltCommitCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultStr, "successfully committed changes")

	_, err = getTableStagedStatus(s, ctx, "commitme", testDoltStatusNewTable)
	require.Error(s.t, err)

	postCommitSha, err := getLastCommitHash(s, ctx)
	require.NoError(s.t, err)
	require.NotEqual(s.t, preCommitSha, postCommitSha)
}
