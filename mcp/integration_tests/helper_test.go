package integration_tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testSuiteHTTPURL = "http://0.0.0.0:8080/mcp"
var testDoltStatusNewTable = "new table"
var testDoltStatusModifiedTable = "modified"

type TableStatus struct {
	Status string
	Staged bool
	TableName string
}

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

func requireTableHasNRows(s *testSuite, ctx context.Context, tableName string, numberOfRows int) {
	var actualCount int 

	row := s.testDb.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) AS count FROM `%s`;", tableName))

	err := row.Scan(&actualCount)
	require.NoError(s.t, err)

	err = row.Err()
	require.NoError(s.t, err)

	require.Equal(s.t, numberOfRows, actualCount)
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

func getDoltStatus(s *testSuite, ctx context.Context, tableName string) ([]*TableStatus, error) {
	rows, err := s.testDb.QueryContext(ctx, fmt.Sprintf("SELECT * FROM dolt_status WHERE table_name = '%s';", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tableStatuses := make([]*TableStatus, 0)
	for rows.Next() {
		var tableName string
		var staged bool
		var status string
		if err := rows.Scan(&tableName, &staged, &status); err != nil {
			return nil, err
		}

		tableStatuses = append(tableStatuses, &TableStatus{
			TableName: tableName,
			Staged: staged,
			Status: status,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return tableStatuses, nil
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

