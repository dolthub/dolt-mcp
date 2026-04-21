package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testCreateTableT1Query = DialectSQL{
	db.DialectMySQL:    "CREATE TABLE `t1` (pk int primary key);",
	db.DialectPostgres: `CREATE TABLE "t1" (pk int primary key);`,
}

var testCreateTablePlacesQuery = DialectSQL{
	db.DialectMySQL: "\n" +
		"CREATE TABLE `places`(\n" +
		"	`id` VARCHAR(36) PRIMARY KEY,\n" +
		"	`name` VARCHAR(1024) NOT NULL,\n" +
		"	`address` VARCHAR(1024) NOT NULL,\n" +
		"	`city` VARCHAR(1024) NOT NULL,\n" +
		"	`country` VARCHAR(1024) NOT NULL\n" +
		");",
	db.DialectPostgres: `
CREATE TABLE "places"(
	"id" VARCHAR(36) PRIMARY KEY,
	"name" VARCHAR(1024) NOT NULL,
	"address" VARCHAR(1024) NOT NULL,
	"city" VARCHAR(1024) NOT NULL,
	"country" VARCHAR(1024) NOT NULL
);`,
}

func testCreateTableToolInvalidArguments(s *testSuite, testBranchName string) {
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
			description:   "Missing working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName: testCreateTableT1Query.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Empty working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "",
						tools.QueryCallToolArgumentName:         testCreateTableT1Query.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
						tools.QueryCallToolArgumentName:         testCreateTableT1Query.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.QueryCallToolArgumentName:         testCreateTableT1Query.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.QueryCallToolArgumentName:           testCreateTableT1Query.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.QueryCallToolArgumentName:           testCreateTableT1Query.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Missing query argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateTableToolName,
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
					Name: tools.CreateTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName:           "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
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
						tools.QueryCallToolArgumentName:           "insert into people values (uuid(), 'homer', 'simpson');",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
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

func testCreateTableToolSuccess(s *testSuite, testBranchName string) {
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
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
				tools.QueryCallToolArgumentName: testCreateTablePlacesQuery.Get(s.dialectType),
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
