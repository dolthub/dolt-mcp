package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDropTableSetupSQL = `CREATE TABLE ` + "`" + `places` + "`" + `(
` + "`" + `id` + "`" + `VARCHAR(36) PRIMARY KEY,
` + "`" + `name` + "`" + `VARCHAR(1024) NOT NULL,
` + "`" + `address` + "`" + `VARCHAR(1024) NOT NULL,
` + "`" + `city` + "`" + `VARCHAR(1024) NOT NULL,
` + "`" + `country` + "`" + `VARCHAR(1024) NOT NULL);`

func testDropTableToolInvalidArguments(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DropTableToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropTableToolName,
				},
			},
		},
		{
			description:   "Empty table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Non-existent table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName: "bar",
					},
				},
			},
		},
	}

	for _, request := range requests {
		dropTableCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, dropTableCallToolResult.IsError)
		} else {
			require.False(s.t, dropTableCallToolResult.IsError)
		}

		require.NotNil(s.t, dropTableCallToolResult)
		require.NotEmpty(s.t, dropTableCallToolResult.Content)
	}
}

func testDropTableToolSuccess(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DropTableToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description: "Drops existing table",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName: "places",
					},
				},
			},
		},
		{
			description: "Drops non-existent database",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DropTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName: "foo",
						tools.IfExistsCallToolArgumentName: true,
					},
				},
			},
		},
	}

	for _, request := range requests {
		dropTableCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)
		require.False(s.t, dropTableCallToolResult.IsError)
		require.NotNil(s.t, dropTableCallToolResult)
		require.NotEmpty(s.t, dropTableCallToolResult.Content)
		resultString, err := resultToString(dropTableCallToolResult)
		require.NoError(s.t, err)
		require.Contains(s.t, resultString, "successfully dropped table")
	}
}

