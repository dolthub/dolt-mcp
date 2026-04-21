package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testDescribeTableToolQuery = DialectSQL{
	db.DialectMySQL:    "DESCRIBE `people`;",
	db.DialectPostgres: `DESCRIBE "people";`,
}

func testDescribeTableToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DescribeTableToolName)

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
					Name: tools.DescribeTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName:           testDescribeTableToolQuery.Get(s.dialectType),
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
					Name: tools.DescribeTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.QueryCallToolArgumentName:           testDescribeTableToolQuery.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DescribeTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   "doesnotexist",
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.QueryCallToolArgumentName:           testDescribeTableToolQuery.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DescribeTableToolName,
					Arguments: map[string]any{
						tools.QueryCallToolArgumentName:         testDescribeTableToolQuery.Get(s.dialectType),
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DescribeTableToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.QueryCallToolArgumentName:           testDescribeTableToolQuery.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DescribeTableToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.QueryCallToolArgumentName:           testDescribeTableToolQuery.Get(s.dialectType),
					},
				},
			},
		},
		{
			description:   "Missing table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DescribeTableToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description:   "Empty table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DescribeTableToolName,
					Arguments: map[string]any{
						tools.TableCallToolArgumentName:           "",
						tools.WorkingBranchCallToolArgumentName:   testBranchName,
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		describeTableCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, describeTableCallToolResult.IsError)
		} else {
			require.False(s.t, describeTableCallToolResult.IsError)
		}

		require.NotNil(s.t, describeTableCallToolResult)
		require.NotEmpty(s.t, describeTableCallToolResult.Content)
	}
}

func testDescribeTableToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.DescribeTableToolName)

	describeTableRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.DescribeTableToolName,
			Arguments: map[string]any{
				tools.TableCallToolArgumentName:           "people",
				tools.WorkingBranchCallToolArgumentName:   testBranchName,
				tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			},
		},
	}

	dropDatabaseCallToolResult, err := client.CallTool(ctx, describeTableRequest)
	require.NoError(s.t, err)
	require.False(s.t, dropDatabaseCallToolResult.IsError)
	require.NotNil(s.t, dropDatabaseCallToolResult)
	require.NotEmpty(s.t, dropDatabaseCallToolResult.Content)
	resultString, err := resultToString(dropDatabaseCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "id")
	require.Contains(s.t, resultString, "first_name")
	require.Contains(s.t, resultString, "last_name")
}
