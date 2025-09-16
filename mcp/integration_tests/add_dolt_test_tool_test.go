package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// Add test setup: no-op; we'll add then verify by running
var testAddDoltTestSetupSQL = ""

func testAddDoltTestToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()
	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.AddDoltTestToolName)

	// Missing required args
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tools.AddDoltTestToolName,
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName:   testBranchName,
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			// missing rest
		},
	}}
	res, err := client.CallTool(ctx, req)
	require.NoError(s.t, err)
	require.True(s.t, res.IsError)
}

func testAddDoltTestToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()
	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.AddDoltTestToolName)
	requireToolExists(s, ctx, client, serverInfo, tools.RunDoltTestsToolName)

	// Add a test
	add := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tools.AddDoltTestToolName,
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName:   testBranchName,
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.TestNameCallToolArgumentName:        "test_added",
			tools.TestGroupCallToolArgumentName:       "grp",
			tools.QueryCallToolArgumentName:           "SELECT 1",
			tools.AssertionTypeCallToolArgumentName:   "expected_single_value",
			tools.AssertionComparatorCallToolArgumentName: "==",
			tools.AssertionValueCallToolArgumentName:  "1",
		},
	}}
	res, err := client.CallTool(ctx, add)
	require.NoError(s.t, err)
	require.False(s.t, res.IsError)

	// Now run by name
	run := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tools.RunDoltTestsToolName,
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName:   testBranchName,
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.TargetCallToolArgumentName:          "test_added",
		},
	}}
	runRes, err := client.CallTool(ctx, run)
	require.NoError(s.t, err)
	require.False(s.t, runRes.IsError)
	out, err := resultToString(runRes)
	require.NoError(s.t, err)
	require.Contains(s.t, out, "test_added")
}
