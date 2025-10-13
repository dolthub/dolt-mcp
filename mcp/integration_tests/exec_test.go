package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testExecToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ExecToolName)

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
					Name: tools.ExecToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName:           "INSERT INTO people (id, first_name, last_name) VALUES (UUID(), 'homer', 'simpson');",
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
					Name: tools.ExecToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.QueryCallToolArgumentName:           "INSERT INTO people (id, first_name, last_name) VALUES (UUID(), 'homer', 'simpson');",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.QueryCallToolArgumentName:           "INSERT INTO people (id, first_name, last_name) VALUES (UUID(), 'homer', 'simpson');",
					},
				},
			},
		},
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ExecToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName:         "INSERT INTO people (id, first_name, last_name) VALUES (UUID(), 'homer', 'simpson');",
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
					Name: tools.ExecToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.QueryCallToolArgumentName:           "INSERT INTO people (id, first_name, last_name) VALUES (UUID(), 'homer', 'simpson');",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.QueryCallToolArgumentName:           "INSERT INTO people (id, first_name, last_name) VALUES (UUID(), 'homer', 'simpson');",
					},
				},
			},
		},
		{
			description:   "Missing query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ExecToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ExecToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName:           "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Invalid SQL WRITE query",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ExecToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName:           "this is not sql",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		execCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, execCallToolResult.IsError)
		} else {
			require.False(s.t, execCallToolResult.IsError)
		}

		require.NotNil(s.t, execCallToolResult)
		require.NotEmpty(s.t, execCallToolResult.Content)
	}
}

func testExecToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ExecToolName)

	execToolCallRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.ExecToolName,
			Arguments: map[string]any{
				tools.QueryCallToolArgumentName:           "INSERT INTO people (id, first_name, last_name) VALUES (UUID(), 'homer', 'simpson');",
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	execCallToolResult, err := client.CallTool(ctx, execToolCallRequest)
	require.NoError(s.t, err)
	require.False(s.t, execCallToolResult.IsError)
	require.NotNil(s.t, execCallToolResult)
	require.NotEmpty(s.t, execCallToolResult.Content)
	resultStr, err := resultToString(execCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultStr, "successfully executed write")
}
