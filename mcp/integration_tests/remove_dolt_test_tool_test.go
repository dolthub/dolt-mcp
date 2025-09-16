package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// Remove test setup: pre-create a test to delete
var testRemoveDoltTestSetupSQL = `
REPLACE INTO dolt_tests VALUES ('test_remove_me', 'grp', 'SELECT 1', 'expected_single_value', '==', '1');
`

func testRemoveDoltTestToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()
	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.RemoveDoltTestToolName)

	req := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tools.RemoveDoltTestToolName,
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName:   testBranchName,
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.TestNameCallToolArgumentName:        "test_remove_me",
		},
	}}
	res, err := client.CallTool(ctx, req)
	require.NoError(s.t, err)
	require.False(s.t, res.IsError)
}
