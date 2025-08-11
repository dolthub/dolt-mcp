package integration_tests

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testCloneDatabaseTeardownSQL = "DROP DATABASE alt;"

func testCloneDatabaseToolInvalidArguments(s *testSuite, testBranchName string) {
	ctx := context.Background()

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CloneDatabaseToolName)

	requests := []struct {
		description   string
		request       mcp.CallToolRequest
		errorExpected bool
	}{
		{
			description:   "Missing remote url argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CloneDatabaseToolName,
				},
			},
		},
		{
			description:   "Empty remote url argument",
			errorExpected: true,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: tools.CloneDatabaseToolName,
					Arguments: map[string]any{
						tools.RemoteURLCallToolArgumentName:  "",
					},
				},
			},
		},
	}

	for _, request := range requests {
		cloneDatabaseCallToolResult, err := client.CallTool(ctx, request.request)
		require.NoError(s.t, err)

		if request.errorExpected {
			require.True(s.t, cloneDatabaseCallToolResult.IsError)
		} else {
			require.False(s.t, cloneDatabaseCallToolResult.IsError)
		}

		require.NotNil(s.t, cloneDatabaseCallToolResult)
		require.NotEmpty(s.t, cloneDatabaseCallToolResult.Content)
	}
}

func testCloneDatabaseToolSuccess(s *testSuite, testBranchName string) {
	ctx := context.Background()

	setupRemoteDatabaseSQL := `CREATE TABLE t1 (pk int PRIMARY KEY);
INSERT INTO t1 VALUES (1);
CALL DOLT_COMMIT('-Am', 'add t1 with value 1');`

	fileRemoteDatabase := NewFileRemoteDatabase(s, "alt")
	err := fileRemoteDatabase.Setup(ctx, setupRemoteDatabaseSQL)
	defer fileRemoteDatabase.Teardown(ctx)

	client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
	require.NoError(s.t, err)
	require.NotNil(s.t, client)

	serverInfo, err := client.Initialize(ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, serverInfo)

	requireToolExists(s, ctx, client, serverInfo, tools.CloneDatabaseToolName)

	cloneDatabaseCallToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: tools.CloneDatabaseToolName,
			Arguments: map[string]any{
				tools.NameCallToolArgumentName: "alt",
				tools.RemoteURLCallToolArgumentName:  fileRemoteDatabase.GetRemoteURL(),
			},
		},
	}

	cloneDatabaseCallToolResult, err := client.CallTool(ctx, cloneDatabaseCallToolRequest)
	require.NoError(s.t, err)
	require.False(s.t, cloneDatabaseCallToolResult.IsError)
	require.NotNil(s.t, cloneDatabaseCallToolResult)
	require.NotEmpty(s.t, cloneDatabaseCallToolResult.Content)
	resultString, err := resultToString(cloneDatabaseCallToolResult)
	require.NoError(s.t, err)
	require.Contains(s.t, resultString, "successfully cloned database")
}
