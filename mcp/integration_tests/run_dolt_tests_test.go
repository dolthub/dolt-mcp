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

	// Run a single test by name
	reqByName := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tools.RunDoltTestsToolName,
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName:   testBranchName,
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.TargetCallToolArgumentName:          "test_people_count",
		},
	}}
	resByName, err := client.CallTool(ctx, reqByName)
	require.NoError(s.t, err)
	require.False(s.t, resByName.IsError)
	outByName, err := resultToString(resByName)
	require.NoError(s.t, err)
	require.Contains(s.t, outByName, "test_people_count")
	require.NotContains(s.t, outByName, "test_people_columns")

	// Run tests by group name
	reqByGroup := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tools.RunDoltTestsToolName,
		Arguments: map[string]any{
			tools.WorkingBranchCallToolArgumentName:   testBranchName,
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.TargetCallToolArgumentName:          "people",
		},
	}}
	resByGroup, err := client.CallTool(ctx, reqByGroup)
	require.NoError(s.t, err)
	require.False(s.t, resByGroup.IsError)
	outByGroup, err := resultToString(resByGroup)
	require.NoError(s.t, err)
	require.Contains(s.t, outByGroup, "test_people_count")
	require.Contains(s.t, outByGroup, "test_people_columns")
}
