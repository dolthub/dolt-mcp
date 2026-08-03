package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// testDoltResetSoftSetupSQL commits `resetme`, then stages one more row and
// leaves a third row unstaged. That gives the soft reset a commit to move HEAD
// off of, plus staged and working changes that the reset must leave alone.
var testDoltResetSoftSetupSQL = DialectSQL{
	db.DialectMySQL: `CREATE TABLE resetme (pk int primary key);
INSERT INTO resetme VALUES (1);
CALL DOLT_COMMIT('-Am', 'add resetme');
INSERT INTO resetme VALUES (2);
CALL DOLT_ADD('resetme');
INSERT INTO resetme VALUES (3);
`,
	db.DialectPostgres: `CREATE TABLE resetme (pk int primary key);
INSERT INTO resetme VALUES (1);
SELECT dolt_commit('-Am', 'add resetme');
INSERT INTO resetme VALUES (2);
SELECT dolt_add('resetme');
INSERT INTO resetme VALUES (3);
`,
}

func testDoltResetSoftToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltResetSoftToolName)

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
					Name: tools.DoltResetSoftToolName,
					Arguments: map[string]any{
						tools.RevisionCallToolArgumentName:        testBranchName,
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
					Name: tools.DoltResetSoftToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "",
						tools.RevisionCallToolArgumentName:        testBranchName,
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
					Name: tools.DoltResetSoftToolName,
					Arguments: map[string]any{
						tools.RevisionCallToolArgumentName:      testBranchName,
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
					Name: tools.DoltResetSoftToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.RevisionCallToolArgumentName:        testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltResetSoftToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.RevisionCallToolArgumentName:        testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltResetSoftToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.RevisionCallToolArgumentName:        testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Missing revision argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltResetSoftToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty revision argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltResetSoftToolName,
					Arguments: map[string]any{
						tools.RevisionCallToolArgumentName:        "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Non-existent revision argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DoltResetSoftToolName,
					Arguments: map[string]any{
						tools.RevisionCallToolArgumentName:        "bar",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
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

func testDoltResetSoftToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DoltResetSoftToolName)

	commitHashes, err := getCommitHashes(s, ctx)
	require.NoError(s.t, err)
	require.GreaterOrEqual(s.t, len(commitHashes), 2)
	headBeforeReset, parentOfHead := commitHashes[0], commitHashes[1]

	// Both entries are modifications relative to the setup commit: the staged
	// row and the unstaged row.
	tableStatuses, err := getDoltStatus(s, ctx, "resetme")
	require.NoError(s.t, err)
	require.Len(s.t, tableStatuses, 2)
	for _, ts := range tableStatuses {
		require.Equal(s.t, testDoltStatusModifiedTable, ts.Status)
	}

	requireTableHasNRows(s, ctx, "resetme", 3)

	doltResetSoftCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.DoltResetSoftToolName,
			Arguments: map[string]any{
				tools.RevisionCallToolArgumentName:        "HEAD~1",
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	doltResetSoftCallToolResult, err := client.CallTool(ctx, doltResetSoftCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, doltResetSoftCallToolResult.IsError)
	require.NotNil(s.t, doltResetSoftCallToolResult)
	require.NotEmpty(s.t, doltResetSoftCallToolResult.Content)
	resultString, err := resultToString(doltResetSoftCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully soft reset")

	// The branch HEAD moved back one commit...
	headAfterReset, err := getLastCommitHash(s, ctx)
	require.NoError(s.t, err)
	require.NotEqual(s.t, headBeforeReset, headAfterReset)
	require.Equal(s.t, parentOfHead, headAfterReset)

	// ...while the staging area and the working set were left alone. The staged
	// copy of the table now reads as new because HEAD no longer contains it.
	tableStatuses, err = getDoltStatus(s, ctx, "resetme")
	require.NoError(s.t, err)
	require.Len(s.t, tableStatuses, 2)
	for _, ts := range tableStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.True(s.t, ts.Staged)
		} else {
			require.Equal(s.t, testDoltStatusModifiedTable, ts.Status)
			require.False(s.t, ts.Staged)
		}
	}

	requireTableHasNRows(s, ctx, "resetme", 3)
}
