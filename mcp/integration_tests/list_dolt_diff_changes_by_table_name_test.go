package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testListDoltDiffChangesByTableNameSetupSQL = `CREATE TABLE ` + "`" + `t1` + "`" + ` (pk int primary key);
CALL DOLT_COMMIT('-Am', 'add t1');
INSERT INTO ` + "`" + `t1` + "`" + ` VALUES (1);
INSERT INTO ` + "`" + `t1` + "`" + ` VALUES (2);
CALL DOLT_ADD('t1');
`

func testListDoltDiffChangesByTableNameToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltDiffChangesByTableNameToolName)

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
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "t1",
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "t1",
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingBranchCallToolArgumentName:    "",
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "t1",
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingBranchCallToolArgumentName:    "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "t1",
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingBranchCallToolArgumentName:    testBranchName,
					},
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "t1",
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingDatabaseCallToolArgumentName:  "",
						tools.WorkingBranchCallToolArgumentName:    testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "t1",
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingDatabaseCallToolArgumentName:  "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:    testBranchName,
					},
				},
			},
		},
		{
			description:   "Missing table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingBranchCallToolArgumentName:    testBranchName,
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "",
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
						tools.WorkingBranchCallToolArgumentName:    testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "foo",
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
						tools.WorkingBranchCallToolArgumentName:    testBranchName,
					},
				},
			},
		},
		{
			description:   "Missing from commit and hash of from commit argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:           "t1",
						tools.HashOfToCommitCallToolArgumentName:  "HEAD",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty from commit and hash of from commit argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "t1",
						tools.FromCommitCallToolArgumentName:       "",
						tools.HashOfToCommitCallToolArgumentName:   "HEAD",
						tools.HashOfFromCommitCallToolArgumentName: "",
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
						tools.WorkingBranchCallToolArgumentName:    testBranchName,
					},
				},
			},
		},
		{
			description:   "Missing to commit and hash of to commit argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "t1",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.WorkingBranchCallToolArgumentName:    testBranchName,
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty to commit and hash of to commit argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltDiffChangesByTableNameToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:            "t1",
						tools.HashOfToCommitCallToolArgumentName:   "",
						tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
						tools.ToCommitCallToolArgumentName:         "",
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
						tools.WorkingBranchCallToolArgumentName:    testBranchName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		listDoltDiffChangesByTableNameCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, listDoltDiffChangesByTableNameCallToolResult.IsError)
		} else {
			require.False(s.t, listDoltDiffChangesByTableNameCallToolResult.IsError)
		}

		require.NotNil(s.t, listDoltDiffChangesByTableNameCallToolResult)
		require.NotEmpty(s.t, listDoltDiffChangesByTableNameCallToolResult.Content)
	}
}

func testListDoltDiffChangesByTableNameToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltDiffChangesByTableNameToolName)

	previousCommitHash, err := getLastCommitHash(s, ctx)
	require.NoError(s.t, err)

	listDoltDiffChangesByTableNameCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.ListDoltDiffChangesByTableNameToolName,
			Arguments: map[string]any{
				tools.WorkingBranchCallToolArgumentName:    testBranchName,
				tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
				tools.TableCallToolArgumentName:            "t1",
				tools.HashOfToCommitCallToolArgumentName:   "HEAD",
				tools.HashOfFromCommitCallToolArgumentName: "HEAD^",
			},
		},
	}

	listDoltDiffChangesByTableNameCallToolResult, err := client.CallTool(ctx, listDoltDiffChangesByTableNameCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, listDoltDiffChangesByTableNameCallToolResult.IsError)
	require.NotNil(s.t, listDoltDiffChangesByTableNameCallToolResult)
	require.NotEmpty(s.t, listDoltDiffChangesByTableNameCallToolResult.Content)
	resultString, err := resultToString(listDoltDiffChangesByTableNameCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, previousCommitHash)
}
