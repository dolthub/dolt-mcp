package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testAlterTableToolInvalidArguments(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.AlterTableToolName)

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
					Name: tools.AlterTableToolName,
				},
			},
		},
		{
			description:   "Empty query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: "",
					},
				},
			},
		},
		{
			description:   "Invalid alter table statement",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: "insert into people values (uuid(), 'homer', 'simpson');",
					},
				},
			},
		},
	}

	for _, request := range requests {
		alterTableCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, alterTableCallToolResult.IsError)
		} else {
			require.False(s.t, alterTableCallToolResult.IsError)
		}

		require.NotNil(s.t, alterTableCallToolResult)
		require.NotEmpty(s.t, alterTableCallToolResult.Content)
	}
}

func testAlterTableToolSuccess(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.AlterTableToolName)

	alterTableToolCallRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.AlterTableToolName,
			Arguments: map[string]any{
				tools.QueryCallToolArgumentName: "ALTER TABLE `people` ADD COLUMN `age` INT NOT NULL;", 
			},
		},
	}

	alterTableCallToolResult, err := client.CallTool(ctx, alterTableToolCallRequest)
	require.NoError(s.t, err)
	require.False(s.t, alterTableCallToolResult.IsError)
	require.NotNil(s.t, alterTableCallToolResult)
	require.NotEmpty(s.t, alterTableCallToolResult.Content)
	resultStr, err := resultToString(alterTableCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultStr, "successfully altered table")
}

