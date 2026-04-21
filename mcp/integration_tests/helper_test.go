package integration_tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var testSuiteHTTPURL = "http://0.0.0.0:8080/mcp"
var testDoltStatusNewTable = "new table"
var testDoltStatusModifiedTable = "modified"

// countRowsInTableSQL counts the rows in a table. The table name is inserted
// via fmt.Sprintf. Backticks are used for Dolt and double-quotes for Doltgres
// to correctly quote identifiers.
var countRowsInTableSQL = DialectSQL{
	db.DialectMySQL:    "SELECT COUNT(*) AS count FROM `%s`;",
	db.DialectPostgres: `SELECT COUNT(*) AS count FROM "%s";`,
}

// selectDoltStatusForTableSQL selects rows from dolt_status for a given table.
// Doltgres qualifies the table name with the schema (e.g. "public.foo"),
// while Dolt uses the bare table name. The bare table name is inserted
// via fmt.Sprintf.
var selectDoltStatusForTableSQL = DialectSQL{
	db.DialectMySQL:    "SELECT * FROM dolt_status WHERE table_name = '%s';",
	db.DialectPostgres: `SELECT * FROM dolt_status WHERE table_name = 'public.%s';`,
}

// selectLastCommitHashSQL selects the most recent commit hash from dolt_log.
var selectLastCommitHashSQL = DialectSQL{
	db.DialectMySQL:    "SELECT commit_hash FROM dolt_log ORDER BY date DESC LIMIT 1;",
	db.DialectPostgres: "SELECT commit_hash FROM dolt_log ORDER BY date DESC LIMIT 1;",
}

type TableStatus struct {
	Status    string
	Staged    bool
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

	row := s.testDb.QueryRowContext(ctx, fmt.Sprintf(countRowsInTableSQL.Get(s.dialectType), tableName))

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
	rows, err := s.testDb.QueryContext(ctx, fmt.Sprintf(selectDoltStatusForTableSQL.Get(s.dialectType), tableName))
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
			Staged:    staged,
			Status:    status,
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

	row := s.testDb.QueryRowContext(ctx, selectLastCommitHashSQL.Get(s.dialectType))

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
