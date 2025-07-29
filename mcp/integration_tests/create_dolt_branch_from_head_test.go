package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testCreateDoltBranchFromHeadToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CreateDoltBranchFromHeadToolName)

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
					Name: tools.CreateDoltBranchFromHeadToolName,
					Arguments: map[string]any{
						tools.NewBranchCallToolArgumentName: "valid",
					},
				},
			},
		},
		{
			description:   "Empty working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchFromHeadToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "",
						tools.NewBranchCallToolArgumentName:     "valid",
					},
				},
			},
		},
		{
			description:   "Non-existent working_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchFromHeadToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: "doesnotexist",
						tools.NewBranchCallToolArgumentName:     "valid",
					},
				},
			},
		},
		{
			description:   "Missing new_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchFromHeadToolName,
					Arguments: map[string]any{
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Empty new_branch argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchFromHeadToolName,
					Arguments: map[string]any{
						tools.NewBranchCallToolArgumentName:     "",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description:   "Existing branch",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchFromHeadToolName,
					Arguments: map[string]any{
						tools.NewBranchCallToolArgumentName:     testBranchName,
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
	}

	for _, request := range requests {
		createDoltBranchFromHeadCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, createDoltBranchFromHeadCallToolResult.IsError)
		} else {
			require.False(s.t, createDoltBranchFromHeadCallToolResult.IsError)
		}

		require.NotNil(s.t, createDoltBranchFromHeadCallToolResult)
		require.NotEmpty(s.t, createDoltBranchFromHeadCallToolResult.Content)
	}
}

func testCreateDoltBranchFromHeadToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CreateDoltBranchFromHeadToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description: "Creates new branch the doesnt exist",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchFromHeadToolName,
					Arguments: map[string]any{
						tools.NewBranchCallToolArgumentName:     "valid",
						tools.WorkingBranchCallToolArgumentName: testBranchName,
					},
				},
			},
		},
		{
			description: "Forces new branch even if branch exists",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CreateDoltBranchFromHeadToolName,
					Arguments: map[string]any{
						tools.NewBranchCallToolArgumentName:     testBranchName,
						tools.WorkingBranchCallToolArgumentName: testBranchName,
						tools.ForceCallToolArgumentName: true,
					},
				},
			},
		},
	}

	for _, request := range requests {
		createDoltBranchFromHeadCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)
		require.False(s.t, createDoltBranchFromHeadCallToolResult.IsError)
		require.NotNil(s.t, createDoltBranchFromHeadCallToolResult)
		require.NotEmpty(s.t, createDoltBranchFromHeadCallToolResult.Content)
		resultString, err := resultToString(createDoltBranchFromHeadCallToolResult)
		require.NoError(s.t, err)
		require.Contains(s.t, resultString, "successfully created branch")
	}
}

