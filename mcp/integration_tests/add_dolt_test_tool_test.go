package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testAddDoltTestToolInvalidArguments(s *testSuite, testBranchName string) {
    ctx := context.Background()
    client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
    require.NoError(s.t, err)
    require.NotNil(s.t, client)

    serverInfo, err := client.Initialize(ctx)
    require.NoError(s.t, err)
    require.NotNil(s.t, serverInfo)

    requireToolExists(s, ctx, client, serverInfo, tools.AddDoltTestToolName)

    // Base valid arguments
    baseArgs := map[string]any{
        tools.WorkingBranchCallToolArgumentName:         testBranchName,
        tools.WorkingDatabaseCallToolArgumentName:       mcpTestDatabaseName,
        tools.TestNameCallToolArgumentName:              "test_missing_args",
        tools.QueryCallToolArgumentName:                 "SELECT 1",
        tools.AssertionTypeCallToolArgumentName:         "expected_single_value",
        tools.AssertionComparatorCallToolArgumentName:   "==",
        // Optional: tools.TestGroupCallToolArgumentName, tools.AssertionValueCallToolArgumentName
    }

    // Table of required arguments to omit one-by-one
    cases := []struct{
        description string
        missingKey  string
    }{
        {description: "Missing working_branch", missingKey: tools.WorkingBranchCallToolArgumentName},
        {description: "Missing working_database", missingKey: tools.WorkingDatabaseCallToolArgumentName},
        {description: "Missing test_name", missingKey: tools.TestNameCallToolArgumentName},
        {description: "Missing query", missingKey: tools.QueryCallToolArgumentName},
        {description: "Missing assertion_type", missingKey: tools.AssertionTypeCallToolArgumentName},
        {description: "Missing assertion_comparator", missingKey: tools.AssertionComparatorCallToolArgumentName},
    }

    for _, c := range cases {
        // Copy base args and remove the missing key
        args := make(map[string]any, len(baseArgs)-1)
        for k, v := range baseArgs {
            if k == c.missingKey {
                continue
            }
            args[k] = v
        }

        req := mcp.CallToolRequest{Params: mcp.CallToolParams{
            Name:      tools.AddDoltTestToolName,
            Arguments: args,
        }}
        res, err := client.CallTool(ctx, req)
        require.NoError(s.t, err, c.description)
        require.True(s.t, res.IsError, c.description)
        require.NotNil(s.t, res, c.description)
        require.NotEmpty(s.t, res.Content, c.description)
    }
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
