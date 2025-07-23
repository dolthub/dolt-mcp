package integration_tests

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testSuiteHTTPURL = "http://0.0.0.0:8080/mcp"

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

