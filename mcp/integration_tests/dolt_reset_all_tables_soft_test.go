package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDoltResetAllTablesSoftSetupSQL = `CREATE TABLE ` + "`" + `resetmeone` + "`" + ` (pk int primary key);
CREATE TABLE ` + "`" + `resetmetwo` + "`" + ` (pk int primary key);
INSERT INTO ` + "`" + `resetmeone` + "`" + ` VALUES (1);
INSERT INTO ` + "`" + `resetmetwo` + "`" + ` VALUES (1);
CALL DOLT_ADD('resetmeone');
CALL DOLT_ADD('resetmetwo');
INSERT INTO ` + "`" + `resetmeone` + "`" + ` VALUES (2);
INSERT INTO ` + "`" + `resetmetwo` + "`" + ` VALUES (2);
`

func testDoltResetAllTablesSoftToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltResetAllTablesSoftToolName)

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
					Name: tools.DoltResetAllTablesSoftToolName,
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
					Name: tools.DoltResetAllTablesSoftToolName,
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
					Name: tools.DoltResetAllTablesSoftToolName,
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
					Name: tools.DoltResetAllTablesSoftToolName,
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
					Name: tools.DoltResetAllTablesSoftToolName,
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
					Name: tools.DoltResetAllTablesSoftToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		doltResetAllTablesSoftCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, doltResetAllTablesSoftCallToolResult.IsError)
		} else {
			require.False(s.t, doltResetAllTablesSoftCallToolResult.IsError)
		}

		require.NotNil(s.t, doltResetAllTablesSoftCallToolResult)
		require.NotEmpty(s.t, doltResetAllTablesSoftCallToolResult.Content)
	}
}

func testDoltResetAllTablesSoftToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltResetAllTablesSoftToolName)

	resetMeTables := []string{
		"resetmeone",
		"resetmetwo",
	}

	for _, resetMeTable := range resetMeTables {
		tableStatuses, err := getDoltStatus(s, ctx, resetMeTable)
		require.NoError(s.t, err)

		for _, ts := range tableStatuses {
			if ts.Status == testDoltStatusNewTable {
				require.True(s.t, ts.Staged)
			} else if ts.Status == testDoltStatusModifiedTable {
				require.False(s.t, ts.Staged)
			}
		}

		requireTableHasNRows(s, ctx, resetMeTable, 2)
	}

	doltResetAllTablesSoftCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.DoltResetAllTablesSoftToolName,
			Arguments: map[string]any{
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	doltResetAllTablesSoftCallToolResult, err := client.CallTool(ctx, doltResetAllTablesSoftCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, doltResetAllTablesSoftCallToolResult.IsError)
	require.NotNil(s.t, doltResetAllTablesSoftCallToolResult)
	require.NotEmpty(s.t, doltResetAllTablesSoftCallToolResult.Content)
	resultString, err := resultToString(doltResetAllTablesSoftCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully soft reset tables")

	for _, resetMeTable := range resetMeTables {
		tableStatuses, err := getDoltStatus(s, ctx, resetMeTable)
		require.NoError(s.t, err)

		for _, ts := range tableStatuses {
			if ts.Status == testDoltStatusNewTable {
				require.False(s.t, ts.Staged)
			} else if ts.Status == testDoltStatusModifiedTable {
				require.False(s.t, ts.Staged)
			}
		}

		requireTableHasNRows(s, ctx, resetMeTable, 2)
	}
}
