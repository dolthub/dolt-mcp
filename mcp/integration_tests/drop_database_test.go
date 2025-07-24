package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDropDatabaseSetupSQL = "CREATE DATABASE foo;"

func testDropDatabaseToolInvalidArguments(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolMustExist(s, ctx, client, serverInfo, tools.DropDatabaseToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropDatabaseToolName,
				},
			},
		},
		{
			description:   "Empty database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Non-existent database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName: "bar",
					},
				},
			},
		},
	}

	for _, request := range requests {
		dropDatabaseCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, dropDatabaseCallToolResult.IsError)
		} else {
			require.False(s.t, dropDatabaseCallToolResult.IsError)
		}

		require.NotNil(s.t, dropDatabaseCallToolResult)
		require.NotEmpty(s.t, dropDatabaseCallToolResult.Content)
	}
}

func testDropDatabaseToolSuccess(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolMustExist(s, ctx, client, serverInfo, tools.DropDatabaseToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description: "Drops existing database",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName: "foo",
					},
				},
			},
		},
		{
			description: "Drops non-existent database",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName: "foo",
						tools.IfExistsCallToolArgumentName: true,
					},
				},
			},
		},
	}

	for _, request := range requests {
		dropDatabaseCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)
		require.False(s.t, dropDatabaseCallToolResult.IsError)
		require.NotNil(s.t, dropDatabaseCallToolResult)
		require.NotEmpty(s.t, dropDatabaseCallToolResult.Content)
		resultString, err := resultToString(dropDatabaseCallToolResult)
		require.NoError(s.t, err)
		require.Contains(s.t, resultString, "successfully dropped database")
	}
}
