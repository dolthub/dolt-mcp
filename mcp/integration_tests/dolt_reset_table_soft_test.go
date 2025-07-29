package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDoltResetTableSoftSetupSQL = `CREATE TABLE ` + "`" + `resetme` + "`" + ` (pk int primary key);
INSERT INTO ` + "`" + `resetme` + "`" + ` VALUES (1);
CALL DOLT_ADD('resetme');
INSERT INTO ` + "`" + `resetme` + "`" + ` VALUES (2);
`

func testDoltResetTableSoftToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltResetTableSoftToolName)

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
					Name: tools.DoltResetTableSoftToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName: "resetme",
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
					Name: tools.DoltResetTableSoftToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "",
						tools.TableCallToolArgumentName: "resetme",
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
					Name: tools.DoltResetTableSoftToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName: "resetme",
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
					Name: tools.DoltResetTableSoftToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.TableCallToolArgumentName: "resetme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltResetTableSoftToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.TableCallToolArgumentName: "resetme",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltResetTableSoftToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
						tools.TableCallToolArgumentName: "resetme",
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
					Name: tools.DoltResetTableSoftToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName, 
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
					Name: tools.DoltResetTableSoftToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName: testBranchName, 
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
					Name: tools.DoltResetTableSoftToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName: "bar",
						tools.WorkingBranchCallToolArgumentName: testBranchName, 
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		doltResetTableSoftCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, doltResetTableSoftCallToolResult.IsError)
		} else {
			require.False(s.t, doltResetTableSoftCallToolResult.IsError)
		}

		require.NotNil(s.t, doltResetTableSoftCallToolResult)
		require.NotEmpty(s.t, doltResetTableSoftCallToolResult.Content)
	}
}

func testDoltResetTableSoftToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltResetTableSoftToolName)

	resetMeNewTableIsStaged, err := getTableStagedStatus(s, ctx, "resetme", testDoltStatusNewTable)
	require.NoError(s.t, err)
	require.True(s.t, resetMeNewTableIsStaged)

	resetMeModifiedTableIsStaged, err := getTableStagedStatus(s, ctx, "resetme", testDoltStatusModifiedTable)
	require.NoError(s.t, err)
	require.False(s.t, resetMeModifiedTableIsStaged)

	requireTableHasNRows(s, ctx, "resetme", 2)
	
	doltResetTableSoftCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.DoltResetTableSoftToolName,
			Arguments: map[string]any{
				tools.TableCallToolArgumentName: "resetme",
				tools.WorkingBranchCallToolArgumentName: testBranchName, 
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	doltResetTableSoftCallToolResult, err := client.CallTool(ctx, doltResetTableSoftCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, doltResetTableSoftCallToolResult.IsError)
	require.NotNil(s.t, doltResetTableSoftCallToolResult)
	require.NotEmpty(s.t, doltResetTableSoftCallToolResult.Content)
	resultString, err := resultToString(doltResetTableSoftCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully soft reset table")
	
	resetMeNewTableIsStaged, err = getTableStagedStatus(s, ctx, "resetme", testDoltStatusNewTable)
	require.NoError(s.t, err)
	require.False(s.t, resetMeNewTableIsStaged)

	_, err = getTableStagedStatus(s, ctx, "resetme", testDoltStatusModifiedTable)
	require.Error(s.t, err)

	requireTableHasNRows(s, ctx, "resetme", 2)
}

