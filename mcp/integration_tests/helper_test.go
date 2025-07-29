package integration_tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testSuiteHTTPURL = "http://0.0.0.0:8080/mcp"

func requireToolExists(s *testSuite, ctx context.Context, client *TestClient, serverInfo *mcp.InitializeResult, toolName string) {
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

func resultToString(result *mcp.CallToolResult) (string, error) {
	var b strings.Builder

	for _, content := range result.Content {
		text, ok := content.(mcp.TextContent)
		if !ok {
			return "", fmt.Errorf("unsupported content type: %T", content)
		}
		b.WriteString(text.Text)
	}

	if result.IsError {
		return "", fmt.Errorf("%s", b.String())
	}

	return b.String(), nil
}

func getTableStagedStatus(s *testSuite, ctx context.Context, tableName string) (bool, error) {
	var staged bool

	row := s.testDb.QueryRowContext(ctx, fmt.Sprintf("SELECT staged FROM dolt_status WHERE table_name = '%s' LIMIT 1;", tableName))

	err := row.Scan(&staged)
	if err != nil {
		return false, err
	}

	err = row.Err()
	if err != nil {
		return false, err
	}

	return staged, nil
}

func getLastCommitHash(s *testSuite, ctx context.Context) (string, error) {
	var hash string

	row := s.testDb.QueryRowContext(ctx, "SELECT commit_hash FROM dolt_log ORDER BY date DESC LIMIT 1;")

	err := row.Scan(&hash)
	if err != nil {
		return "", err
	}

	err = row.Err()
	if err != nil {
		return "", err
	}

	return hash, nil
}

