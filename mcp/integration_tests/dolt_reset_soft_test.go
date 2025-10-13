package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDoltResetSoftSetupSQL = `CREATE TABLE ` + "`" + `resetme` + "`" + ` (pk int primary key);
INSERT INTO ` + "`" + `resetme` + "`" + ` VALUES (1);
CALL DOLT_ADD('resetme');
INSERT INTO ` + "`" + `resetme` + "`" + ` VALUES (2);
`

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
                        tools.RevisionCallToolArgumentName:     testBranchName,
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

	tableStatuses, err := getDoltStatus(s, ctx, "resetme")
	require.NoError(s.t, err)

	for _, ts := range tableStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.True(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}
	}

	requireTableHasNRows(s, ctx, "resetme", 2)

    doltResetSoftCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
            Name: tools.DoltResetSoftToolName,
            Arguments: map[string]any{
                tools.RevisionCallToolArgumentName:        testBranchName,
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

	tableStatuses, err = getDoltStatus(s, ctx, "resetme")
	require.NoError(s.t, err)

	for _, ts := range tableStatuses {
		if ts.Status == testDoltStatusNewTable {
			require.False(s.t, ts.Staged)
		} else if ts.Status == testDoltStatusModifiedTable {
			require.False(s.t, ts.Staged)
		}
	}

	requireTableHasNRows(s, ctx, "resetme", 2)
}
