package integration_tests

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestDoltLiteFileRemoteOperations(t *testing.T) {
	if suite == nil || suite.dialectType != db.DialectDoltLite {
		t.Skip("DoltLite-specific file remote test")
	}

	RunTest(t, "TestAddPushFetchPullListRemove", func(s *testSuite, branch string) {
		ctx := context.Background()
		client, err := NewMCPHTTPTestClient(testSuiteHTTPURL)
		require.NoError(t, err)
		_, err = client.Initialize(ctx)
		require.NoError(t, err)

		call := func(name string, arguments map[string]any) string {
			result, callErr := client.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{
				Name:      name,
				Arguments: arguments,
			}})
			require.NoError(t, callErr)
			require.NotNil(t, result)
			text, textErr := resultToString(result)
			require.NoError(t, textErr, "tool %s returned an error", name)
			return text
		}

		working := map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
		}

		call(tools.QueryToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.WorkingBranchCallToolArgumentName:   branch,
			tools.QueryCallToolArgumentName:           "SELECT active_branch();",
		})

		remotePath := filepath.Join(t.TempDir(), "remote.db")
		remoteURL := "file://" + remotePath
		call(tools.AddDoltRemoteToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.RemoteNameCallToolArgumentName:      "origin",
			tools.RemoteURLCallToolArgumentName:       remoteURL,
		})
		defer call(tools.RemoveDoltRemoteToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.RemoteNameCallToolArgumentName:      "origin",
		})

		call(tools.DoltPushBranchToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: working[tools.WorkingDatabaseCallToolArgumentName],
			tools.RemoteNameCallToolArgumentName:      "origin",
			tools.BranchCallToolArgumentName:          branch,
		})
		call(tools.DoltPushBranchToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.RemoteNameCallToolArgumentName:      "origin",
			tools.BranchCallToolArgumentName:          branch,
			tools.ForceCallToolArgumentName:           true,
		})

		remoteDB, err := sql.Open(s.dialect.DriverName(), remotePath)
		require.NoError(t, err)
		remoteDB.SetMaxOpenConns(1)
		_, err = remoteDB.ExecContext(ctx, "SELECT dolt_config('user.name', 'remote-test');")
		require.NoError(t, err)
		_, err = remoteDB.ExecContext(ctx, "SELECT dolt_config('user.email', 'remote-test@dolthub.com');")
		require.NoError(t, err)
		_, err = remoteDB.ExecContext(ctx, "SELECT dolt_checkout(?);", branch)
		require.NoError(t, err)
		_, err = remoteDB.ExecContext(ctx, "CREATE TABLE remote_only (id INTEGER PRIMARY KEY);")
		require.NoError(t, err)
		_, err = remoteDB.ExecContext(ctx, "INSERT INTO remote_only VALUES (1);")
		require.NoError(t, err)
		_, err = remoteDB.ExecContext(ctx, "SELECT dolt_commit('-Am', 'advance remote');")
		require.NoError(t, err)
		require.NoError(t, remoteDB.Close())

		call(tools.DoltFetchBranchToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.RemoteNameCallToolArgumentName:      "origin",
			tools.BranchCallToolArgumentName:          branch,
		})
		call(tools.DoltFetchAllBranchesToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.RemoteNameCallToolArgumentName:      "origin",
		})
		call(tools.DoltPullBranchToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.RemoteNameCallToolArgumentName:      "origin",
			tools.BranchCallToolArgumentName:          branch,
		})

		queryResult := call(tools.QueryToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.WorkingBranchCallToolArgumentName:   branch,
			tools.QueryCallToolArgumentName:           "SELECT COUNT(*) AS count FROM remote_only;",
		})
		require.Contains(t, queryResult, "1")

		listResult := call(tools.ListDoltRemotesToolName, map[string]any{
			tools.WorkingDatabaseCallToolArgumentName: mcpTestDatabaseName,
			tools.WorkingBranchCallToolArgumentName:   branch,
		})
		require.Contains(t, listResult, "origin")
		require.Contains(t, listResult, remoteURL)
	})
}
