package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// Setup SQL for run tests success: define two tests
var testRunDoltTestsSetupSQL = `
INSERT INTO dolt_tests VALUES ('test_people_count', 'people', 'SELECT COUNT(*) FROM people', 'expected_single_value', '>=', '1');
INSERT INTO dolt_tests VALUES ('test_people_columns', 'people', 'SELECT * FROM people LIMIT 1', 'expected_columns', '==', '3');
`

func testRunDoltTestsToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()
	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.RunDoltTestsToolName)

	// Missing working_branch
	req1 := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tools.RunDoltTestsToolName,
		Arguments: map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
		},
	}}
	res, err := client.CallTool(ctx, req1)
	require.NoError(s.t, err)
	require.True(s.t, res.IsError)

	// Missing working_database
	req2 := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tools.RunDoltTestsToolName,
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName: testBranchName,
		},
	}}
	res2, err := client.CallTool(ctx, req2)
	require.NoError(s.t, err)
	require.True(s.t, res2.IsError)
}

func testRunDoltTestsToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()
	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.RunDoltTestsToolName)

	// Run all tests
	reqAll := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tools.RunDoltTestsToolName,
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName:   testBranchName,
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
		},
	}}
	resAll, err := client.CallTool(ctx, reqAll)
	require.NoError(s.t, err)
	require.False(s.t, resAll.IsError)
	out, err := resultToString(resAll)
	require.NoError(s.t, err)
	require.Contains(s.t, out, "test_people_count")
	require.Contains(s.t, out, "test_people_columns")
}

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
