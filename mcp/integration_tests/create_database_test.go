package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testCreateDatabaseTeardownSQL = "DROP DATABASE foo;"

func testCreateDatabaseToolInvalidArguments(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolMustExist(s, ctx, client, serverInfo, tools.CreateDatabaseToolName)

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
					Name: tools.CreateDatabaseToolName,
				},
			},
		},
		{
			description:   "Empty database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Existing database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		createDatabaseCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, createDatabaseCallToolResult.IsError)
		} else {
			require.False(s.t, createDatabaseCallToolResult.IsError)
		}

		require.NotNil(s.t, createDatabaseCallToolResult)
		require.NotEmpty(s.t, createDatabaseCallToolResult.Content)
	}
}

func testCreateDatabaseToolSuccess(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolMustExist(s, ctx, client, serverInfo, tools.CreateDatabaseToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description: "Create non-existent database",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName: "foo",
					},
				},
			},
		},
		{
			description: "Create existing database",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDatabaseToolName,
					Arguments: map[string]any{
						tools.DatabaseCallToolArgumentName:    "foo",
						tools.IfNotExistsCallToolArgumentName: true,
					},
				},
			},
		},
	}

	for _, request := range requests {
		createDatabaseCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)
		require.False(s.t, createDatabaseCallToolResult.IsError)
		require.NotNil(s.t, createDatabaseCallToolResult)
		require.NotEmpty(s.t, createDatabaseCallToolResult.Content)
		resultString, err := resultToString(createDatabaseCallToolResult)
		require.NoError(s.t, err)
		require.Contains(s.t, resultString, "successfully created database")
	}
}

