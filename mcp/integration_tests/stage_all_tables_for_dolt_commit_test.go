package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testStageAllTablesForDoltCommitSetupSQL = `CREATE TABLE ` + "`" + `stagemeone` + "`" + ` (pk int primary key);
CREATE TABLE ` + "`" + `stagemetwo` + "`" + ` (pk int primary key);
`

func testStageAllTablesForDoltCommitToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.StageAllTablesForDoltCommitToolName)

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
					Name: tools.StageAllTablesForDoltCommitToolName,
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
					Name: tools.StageAllTablesForDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "",
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
					Name: tools.StageAllTablesForDoltCommitToolName,
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
					Name: tools.StageAllTablesForDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.StageAllTablesForDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.StageAllTablesForDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		stageAllTablesForDoltCommitCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, stageAllTablesForDoltCommitCallToolResult.IsError)
		} else {
			require.False(s.t, stageAllTablesForDoltCommitCallToolResult.IsError)
		}

		require.NotNil(s.t, stageAllTablesForDoltCommitCallToolResult)
		require.NotEmpty(s.t, stageAllTablesForDoltCommitCallToolResult.Content)
	}
}

func testStageAllTablesForDoltCommitToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.StageAllTablesForDoltCommitToolName)

	stageMeOneIsStaged, err := getTableStagedStatus(s, ctx, "stagemeone")
	require.NoError(s.t, err)
	require.False(s.t, stageMeOneIsStaged)

	stageMeTwoIsStaged, err := getTableStagedStatus(s, ctx, "stagemetwo")
	require.NoError(s.t, err)
	require.False(s.t, stageMeTwoIsStaged)

	stageAllTablesForDoltCommitCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.StageAllTablesForDoltCommitToolName,
			Arguments: map[string]any{
				tools.WorkingBranchCallToolArgumentName: testBranchName, 
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	stageAllTablesForDoltCommitCallToolResult, err := client.CallTool(ctx, stageAllTablesForDoltCommitCallToolRequest )
	require.NoError(s.t, err)
	require.False(s.t, stageAllTablesForDoltCommitCallToolResult.IsError)
	require.NotNil(s.t, stageAllTablesForDoltCommitCallToolResult)
	require.NotEmpty(s.t, stageAllTablesForDoltCommitCallToolResult.Content)
	resultString, err := resultToString(stageAllTablesForDoltCommitCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully staged tables")
	
	stageMeOneIsStaged, err = getTableStagedStatus(s, ctx, "stagemeone")
	require.NoError(s.t, err)
	require.True(s.t, stageMeOneIsStaged)

	stageMeTwoIsStaged, err = getTableStagedStatus(s, ctx, "stagemetwo")
	require.NoError(s.t, err)
	require.True(s.t, stageMeTwoIsStaged)
}

