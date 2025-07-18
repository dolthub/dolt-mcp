package integration_tests

import (
	"context"
	"testing"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testSuiteHTTPURL = "http://0.0.0.0:8080/mcp"

func TestPrimitiveToolsetV1(t *testing.T) {
	RunTest(t, "TestListDatabasesTool", testListDatabasesTool)
}

func testListDatabasesTool(s *testSuite) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolMustExist(s, ctx, client, serverInfo, pkg.ListDatabasesToolName)

	listDatabasesCallToolParams := mcp.CallToolParams{
		Name: pkg.ListDatabasesToolName,
	}

	listDatabasesCallToolRequest := mcp.CallToolRequest{
		Params: listDatabasesCallToolParams,
	}

	listDatabasesCallToolResult, err := client.CallTool(ctx, &listDatabasesCallToolRequest)
	require.NoError(s.t, err)
	require.NotNil(s.t, listDatabasesCallToolResult)
	require.False(s.t, listDatabasesCallToolResult.IsError)
	require.NotEmpty(s.t, listDatabasesCallToolResult.Content)

}

func requireToolMustExist(s *testSuite, ctx context.Context, client *TestClient, serverInfo *mcp.InitializeResult, toolName string) {
	require.NotNil(s.t, serverInfo.Capabilities.Tools)
	listToolsResult, err := client.ListTools(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, listToolsResult)
	found := false
	for _, tool := range listToolsResult.Tools {
		if tool.Name == toolName {
			found = true
			break
		}
	}
	require.True(s.t, found)
}

