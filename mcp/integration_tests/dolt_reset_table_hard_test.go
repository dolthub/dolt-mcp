package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDoltResetTableHardSetupSQL = `CREATE TABLE ` + "`" + `resetme` + "`" + ` (pk int primary key);
CALL DOLT_COMMIT('-Am', 'add table resetme');
INSERT INTO ` + "`" + `resetme` + "`" + ` VALUES (1);
CALL DOLT_ADD('resetme');
INSERT INTO ` + "`" + `resetme` + "`" + ` VALUES (2);
`

func testDoltResetTableHardToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltResetTableHardToolName)

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
					Name: tools.DoltResetTableHardToolName,
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
					Name: tools.DoltResetTableHardToolName,
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
					Name: tools.DoltResetTableHardToolName,
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
					Name: tools.DoltResetTableHardToolName,
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
					Name: tools.DoltResetTableHardToolName,
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
					Name: tools.DoltResetTableHardToolName,
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
					Name: tools.DoltResetTableHardToolName,
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
					Name: tools.DoltResetTableHardToolName,
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
					Name: tools.DoltResetTableHardToolName,
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
		doltResetTableHardCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, doltResetTableHardCallToolResult.IsError)
		} else {
			require.False(s.t, doltResetTableHardCallToolResult.IsError)
		}

		require.NotNil(s.t, doltResetTableHardCallToolResult)
		require.NotEmpty(s.t, doltResetTableHardCallToolResult.Content)
	}
}

func testDoltResetTableHardToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltResetTableHardToolName)

	tableStatuses, err := getDoltStatus(s, ctx, "resetme")
	require.NoError(s.t, err)
	
	oneFalse := false
	oneTrue := false
	for _, ts := range tableStatuses {
		if ts.Status == testDoltStatusModifiedTable {
			if ts.Staged {
				oneTrue = true
			} else {
				oneFalse = true
			}
		}
	}
	require.True(s.t, oneTrue)
	require.True(s.t, oneFalse)

	requireTableHasNRows(s, ctx, "resetme", 2)
	
	doltResetTableHardCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.DoltResetTableHardToolName,
			Arguments: map[string]any{
				tools.TableCallToolArgumentName: "resetme",
				tools.WorkingBranchCallToolArgumentName: testBranchName, 
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	doltResetTableHardCallToolResult, err := client.CallTool(ctx, doltResetTableHardCallToolRequest)
	require.NoError(s.t, err)
	resultString, err := resultToString(doltResetTableHardCallToolResult)
	require.NoError(s.t, err)
	require.False(s.t, doltResetTableHardCallToolResult.IsError)
	require.NotNil(s.t, doltResetTableHardCallToolResult)
	require.NotEmpty(s.t, doltResetTableHardCallToolResult.Content)
	// resultString, err := resultToString(doltResetTableHardCallToolResult)
	// require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully hard reset table")

	tableStatuses, err = getDoltStatus(s, ctx, "resetme")
	require.NoError(s.t, err)
	
	for _, ts := range tableStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.True(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}
	}

	requireTableHasNRows(s, ctx, "resetme", 0)
}
