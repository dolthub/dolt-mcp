package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testDescribeTableToolInvalidArguments(s *testSuite) {
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
			description:   "Missing table argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.DescribeTableToolName,
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
						tools.TableCallToolArgumentName: "",
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

func testDescribeTableToolSuccess(s *testSuite) {
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
				tools.TableCallToolArgumentName: "people",
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

