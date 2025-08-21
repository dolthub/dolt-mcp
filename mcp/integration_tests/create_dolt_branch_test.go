package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testCreateDoltBranchTeardownSQL = "CALL DOLT_BRANCH('-D', 'valid');"

func testCreateDoltBranchToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CreateDoltBranchToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.NewBranchCallToolArgumentName:     "valid",
						tools.OriginalBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Empty working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "",
						tools.NewBranchCallToolArgumentName:     "valid",
						tools.OriginalBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Non-existent working_database argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: "doesnotexist",
						tools.NewBranchCallToolArgumentName:     "valid",
						tools.OriginalBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Missing original_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.NewBranchCallToolArgumentName:     "valid",
					},
				},
			},
		},
		{
			description:   "Empty original_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.OriginalBranchCallToolArgumentName: "",
						tools.NewBranchCallToolArgumentName:      "valid",
					},
				},
			},
		},
		{
			description:   "Missing new_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.WorkingBranchCallToolArgumentName:  testBranchName,
						tools.OriginalBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Empty new_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.NewBranchCallToolArgumentName:      "",
						tools.OriginalBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Existing branch",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
						tools.NewBranchCallToolArgumentName:      testBranchName,
						tools.OriginalBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		createDoltBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, createDoltBranchCallToolResult.IsError)
		} else {
			require.False(s.t, createDoltBranchCallToolResult.IsError)
		}

		require.NotNil(s.t, createDoltBranchCallToolResult)
		require.NotEmpty(s.t, createDoltBranchCallToolResult.Content)
	}
}

func testCreateDoltBranchToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CreateDoltBranchToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description: "Creates new branch the doesnt exist",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.OriginalBranchCallToolArgumentName: testBranchName,
						tools.NewBranchCallToolArgumentName:      "valid",
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
					},
				},
			},
		},
		{
			description: "Forces new branch even if branch exists",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchToolName,
					Arguments: map[string]any{
						tools.NewBranchCallToolArgumentName:      testBranchName,
						tools.OriginalBranchCallToolArgumentName: testBranchName,
						tools.WorkingDatabaseCallToolArgumentName:  mcpTestDatabaseName,
						tools.ForceCallToolArgumentName:          true,
					},
				},
			},
		},
	}

	for _, request := range requests {
		createDoltBranchCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)
		require.False(s.t, createDoltBranchCallToolResult.IsError)
		require.NotNil(s.t, createDoltBranchCallToolResult)
		require.NotEmpty(s.t, createDoltBranchCallToolResult.Content)
		resultString, err := resultToString(createDoltBranchCallToolResult)
		require.NoError(s.t, err)
		require.Contains(s.t, resultString, "successfully created branch")
	}
}

