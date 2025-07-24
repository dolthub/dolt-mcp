package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testCreateTableToolInvalidArguments(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CreateTableToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
				},
			},
		},
		{
			description:   "Empty query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Invalid create table statement",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: "insert into people values (uuid(), 'homer', 'simpson');",
					},
				},
			},
		},
	}

	for _, request := range requests {
		createTableCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, createTableCallToolResult.IsError)
		} else {
			require.False(s.t, createTableCallToolResult.IsError)
		}

		require.NotNil(s.t, createTableCallToolResult)
		require.NotEmpty(s.t, createTableCallToolResult.Content)
	}
}

func testCreateTableToolSuccess(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CreateTableToolName)

	createTableToolCallRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.CreateTableToolName,
			Arguments: map[string]any{
				tools.QueryCallToolArgumentName: `
CREATE TABLE ` + "`" + `places` + "`" + `(
	` + "`" + `id` + "`" + `VARCHAR(36) PRIMARY KEY,
	` + "`" + `name` + "`" + `VARCHAR(1024) NOT NULL,
	` + "`" + `address` + "`" + `VARCHAR(1024) NOT NULL,
	` + "`" + `city` + "`" + `VARCHAR(1024) NOT NULL,
	` + "`" + `country` + "`" + `VARCHAR(1024) NOT NULL
);`,
			},
		},
	}

	createTableCallToolResult, err := client.CallTool(ctx, createTableToolCallRequest)
	require.NoError(s.t, err)
	require.False(s.t, createTableCallToolResult.IsError)
	require.NotNil(s.t, createTableCallToolResult)
	require.NotEmpty(s.t, createTableCallToolResult.Content)
	resultStr, err := resultToString(createTableCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultStr, "successfully created table")
}

