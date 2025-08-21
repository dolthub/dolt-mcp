package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testListDoltRemotesSetupSQL = "CALL DOLT_REMOTE('add', 'origin', 'file://myoriginremote');"
var testListDoltRemotesTeardownSQL = "CALL DOLT_REMOTE('remove', 'origin');"
// var testListDoltRemotesTeardownSQL = "SELECT 1;"

func testListDoltRemotesToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltRemotesToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltRemotesToolName,
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltRemotesToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.ListDoltRemotesToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
					},
				},
			},
		},
	}

	for _, request := range requests {
		listDoltRemotesCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, listDoltRemotesCallToolResult.IsError)
		} else {
			require.False(s.t, listDoltRemotesCallToolResult.IsError)
		}

		require.NotNil(s.t, listDoltRemotesCallToolResult)
		require.NotEmpty(s.t, listDoltRemotesCallToolResult.Content)
	}
}

func testListDoltRemotesToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.ListDoltRemotesToolName)

	listDoltRemotesCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.ListDoltRemotesToolName,
			Arguments: map[string]any{
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	listDoltRemotesCallToolResult, err := client.CallTool(ctx, listDoltRemotesCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, listDoltRemotesCallToolResult.IsError)
	require.NotNil(s.t, listDoltRemotesCallToolResult)
	require.NotEmpty(s.t, listDoltRemotesCallToolResult.Content)
	resultString, err := resultToString(listDoltRemotesCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "origin")
	require.Contains(s.t, resultString, "file://")
}
