package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testStageTableForDoltCommitSetupSQL = DialectSQL{
	db.DialectMySQL:    `CREATE TABLE stageme (pk int primary key);`,
	db.DialectPostgres: `CREATE TABLE stageme (pk int primary key);`,
}

func testStageTableForDoltCommitToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.StageTableForDoltCommitToolName)

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
					Name: tools.StageTableForDoltCommitToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:           "stageme",
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
					Name: tools.StageTableForDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "",
						tools.TableCallToolArgumentName:           "stageme",
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
					Name: tools.StageTableForDoltCommitToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:         "stageme",
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
					Name: tools.StageTableForDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.TableCallToolArgumentName:           "stageme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.StageTableForDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.TableCallToolArgumentName:           "stageme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.StageTableForDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.TableCallToolArgumentName:           "stageme",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Missing table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.StageTableForDoltCommitToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.StageTableForDoltCommitToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:           "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Non-existent table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.StageTableForDoltCommitToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:           "bar",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		stageTableForDoltCommitCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, stageTableForDoltCommitCallToolResult.IsError)
		} else {
			require.False(s.t, stageTableForDoltCommitCallToolResult.IsError)
		}

		require.NotNil(s.t, stageTableForDoltCommitCallToolResult)
		require.NotEmpty(s.t, stageTableForDoltCommitCallToolResult.Content)
	}
}

func testStageTableForDoltCommitToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.StageTableForDoltCommitToolName)

	tableStatuses, err := getDoltStatus(s, ctx, "stageme")
	require.NoError(s.t, err)

	for _, ts := range tableStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.False(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}
	}

	stageTableForDoltCommitCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.StageTableForDoltCommitToolName,
			Arguments: map[string]any{
				tools.TableCallToolArgumentName:           "stageme",
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	stageTableForDoltCommitCallToolResult, err := client.CallTool(ctx, stageTableForDoltCommitCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, stageTableForDoltCommitCallToolResult.IsError)
	require.NotNil(s.t, stageTableForDoltCommitCallToolResult)
	require.NotEmpty(s.t, stageTableForDoltCommitCallToolResult.Content)
	resultString, err := resultToString(stageTableForDoltCommitCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully staged table")

	tableStatuses, err = getDoltStatus(s, ctx, "stageme")
	require.NoError(s.t, err)

	for _, ts := range tableStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.True(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}
	}
}
