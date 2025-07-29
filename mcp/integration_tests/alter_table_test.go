package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testAlterTableToolInvalidArguments(s *testSuite, testBranchName string) {
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
			description:   "Missing working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: "ALTER TABLE `people` ADD COLUMN `age` INT NOT NULL;",
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
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "",
						tools.QueryCallToolArgumentName:         "ALTER TABLE `people` ADD COLUMN `age` INT NOT NULL;",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
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
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
						tools.QueryCallToolArgumentName:         "ALTER TABLE `people` ADD COLUMN `age` INT NOT NULL;",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.QueryCallToolArgumentName:         "ALTER TABLE `people` ADD COLUMN `age` INT NOT NULL;",
					},
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.QueryCallToolArgumentName:           "ALTER TABLE `people` ADD COLUMN `age` INT NOT NULL;",
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
						tools.QueryCallToolArgumentName:         "ALTER TABLE `people` ADD COLUMN `age` INT NOT NULL;",
					},
				},
			},
		},
		{
			description:   "Missing query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName,
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
					Name: tools.AlterTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName:         "",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
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
						tools.QueryCallToolArgumentName:         "insert into people values (uuid(), 'homer', 'simpson');",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
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

func testAlterTableToolSuccess(s *testSuite, testBranchName string) {
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
				tools.QueryCallToolArgumentName:         "ALTER TABLE `people` ADD COLUMN `age` INT NOT NULL;",
				tools.WorkingBranchCallToolArgumentName: testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
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
