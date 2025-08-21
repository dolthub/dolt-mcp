package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testRemoveDoltRemoteSetupSQL = "CALL DOLT_REMOTE('add', 'origin', 'file://myurl');"

func testRemoveDoltRemoteToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.RemoveDoltRemoteToolName)

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
					Name: tools.RemoveDoltRemoteToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName: "origin",
					},
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.RemoveDoltRemoteToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.RemoteNameCallToolArgumentName:      "origin",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.RemoveDoltRemoteToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.RemoteNameCallToolArgumentName:      "origin",
					},
				},
			},
		},
		{
			description:   "Missing remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.RemoveDoltRemoteToolName,
				},
			},
		},
		{
			description:   "Empty remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.RemoveDoltRemoteToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName: "",
					},
				},
			},
		},
	}

	for _, request := range requests {
		removeDoltRemoteCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, removeDoltRemoteCallToolResult.IsError)
		} else {
			require.False(s.t, removeDoltRemoteCallToolResult.IsError)
		}

		require.NotNil(s.t, removeDoltRemoteCallToolResult)
		require.NotEmpty(s.t, removeDoltRemoteCallToolResult.Content)
	}
}

func testRemoveDoltRemoteToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.RemoveDoltRemoteToolName)

	requireTableHasNRows(s, ctx, "dolt_remotes", 1)

	removeDoltRemoteCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.RemoveDoltRemoteToolName,
			Arguments: map[string]any{
				tools.RemoteNameCallToolArgumentName:      "origin",
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	removeDoltRemoteCallToolResult, err := client.CallTool(ctx, removeDoltRemoteCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, removeDoltRemoteCallToolResult.IsError)
	require.NotNil(s.t, removeDoltRemoteCallToolResult)
	require.NotEmpty(s.t, removeDoltRemoteCallToolResult.Content)
	resultString, err := resultToString(removeDoltRemoteCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully removed remote")

	requireTableHasNRows(s, ctx, "dolt_remotes", 0)
}
