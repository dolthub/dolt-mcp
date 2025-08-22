package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testAddDoltRemoteTeardownSQL = "CALL DOLT_REMOTE('remove', 'origin');"

func testAddDoltRemoteToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.AddDoltRemoteToolName)

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
					Name: tools.AddDoltRemoteToolName,
					Arguments: map[string]any{
						tools.RemoteURLCallToolArgumentName: "file://myurl",
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
					Name: tools.AddDoltRemoteToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.RemoteURLCallToolArgumentName: "file://myurl",
						tools.RemoteNameCallToolArgumentName: "origin",
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AddDoltRemoteToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.RemoteURLCallToolArgumentName: "file://myurl",
						tools.RemoteNameCallToolArgumentName: "origin",
					},
				},
			},
		},
		{
			description:   "Missing remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AddDoltRemoteToolName,
					Arguments: map[string]any{
						tools.RemoteURLCallToolArgumentName: "file://myurl",
					},
				},
			},
		},
		{
			description:   "Empty remote name argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AddDoltRemoteToolName,
					Arguments: map[string]any{
						tools.RemoteURLCallToolArgumentName:  "file://myurl",
						tools.RemoteNameCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Missing remote url argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AddDoltRemoteToolName,
					Arguments: map[string]any{
						tools.RemoteNameCallToolArgumentName: "origin",
					},
				},
			},
		},
		{
			description:   "Empty remote url argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AddDoltRemoteToolName,
					Arguments: map[string]any{
						tools.RemoteURLCallToolArgumentName:  "",
						tools.RemoteNameCallToolArgumentName: "origin",
					},
				},
			},
		},
	}

	for _, request := range requests {
		addDoltRemoteCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, addDoltRemoteCallToolResult.IsError)
		} else {
			require.False(s.t, addDoltRemoteCallToolResult.IsError)
		}

		require.NotNil(s.t, addDoltRemoteCallToolResult)
		require.NotEmpty(s.t, addDoltRemoteCallToolResult.Content)
	}
}

func testAddDoltRemoteToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.AddDoltRemoteToolName)

	requireTableHasNRows(s, ctx, "dolt_remotes", 0)

	addDoltRemoteCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.AddDoltRemoteToolName,
			Arguments: map[string]any{
				tools.RemoteNameCallToolArgumentName: "origin",
				tools.RemoteURLCallToolArgumentName:  "file://myurl",
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	addDoltRemoteCallToolResult, err := client.CallTool(ctx, addDoltRemoteCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, addDoltRemoteCallToolResult.IsError)
	require.NotNil(s.t, addDoltRemoteCallToolResult)
	require.NotEmpty(s.t, addDoltRemoteCallToolResult.Content)
	resultString, err := resultToString(addDoltRemoteCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully added remote")

	requireTableHasNRows(s, ctx, "dolt_remotes", 1)
}
