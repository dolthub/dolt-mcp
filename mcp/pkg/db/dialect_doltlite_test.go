package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoltLiteDialectCallProcedure(t *testing.T) {
	d := NewDoltLiteDialect()

	require.Equal(t,
		"SELECT dolt_commit('-m', 'my message');",
		d.CallProcedure(DoltCommit, "-m", "my message"))

	require.Equal(t,
		"SELECT dolt_branch('feature');",
		d.CallProcedure(DoltBranch, "feature"))

	require.Equal(t,
		"SELECT CASE WHEN active_branch() = 'feature' THEN 0 ELSE dolt_checkout('feature') END;",
		d.CallProcedure(DoltCheckout, "feature"))

	require.Equal(t,
		"SELECT dolt_commit('-m', 'it''s a message');",
		d.CallProcedure(DoltCommit, "-m", "it's a message"))

	require.Equal(t, "SELECT dolt_reset();", d.CallProcedure(DoltReset, "."))
	require.Equal(t,
		"SELECT CASE WHEN active_branch() = 'main' THEN 0 ELSE dolt_branch('-f', 'main') END;",
		d.CallProcedure(DoltBranch, "-f", "main"))
	require.Equal(t,
		"SELECT dolt_push('origin', 'main', '--force');",
		d.CallProcedure(DoltPush, "--force", "origin", "main"))
}

func TestDoltLiteDialectSQLGeneration(t *testing.T) {
	d := NewDoltLiteDialect()

	require.Equal(t, `"my""table"`, d.QuoteIdentifier(`my"table`))
	require.Equal(t, "", d.UseDatabase("anything"))
	require.Contains(t, d.ShowTablesQuery(), "sqlite_schema")
	require.Contains(t, d.ShowCreateTableQuery("people"), "name = 'people'")
	require.Contains(t, d.DescribeTableQuery("people"), "pragma_table_info('people')")
	require.Contains(t, d.ShowCreateTableQuery("it's"), "name = 'it''s'")
}

func TestDiffTableIdentifiersAndRefsAreEscaped(t *testing.T) {
	fromExpr := "'from'"
	toExpr := "'to'"

	mysql := NewMySQLDialect()
	require.Equal(t,
		"SELECT * FROM `dolt_diff_odd``name` WHERE from_commit = 'from' AND to_commit = 'to';",
		mysql.ListTableDiffChangesQuery("odd`name", fromExpr, toExpr))
	require.Equal(t, "HASHOF('a''b')", mysql.HashOfFunction("a'b"))
	require.Equal(t, "CALL DOLT_BRANCH('a''b');", mysql.CallProcedure(DoltBranch, "a'b"))

	postgres := NewPostgresDialect()
	require.Equal(t,
		`SELECT * FROM "dolt_diff_odd""name" WHERE from_commit = 'from' AND to_commit = 'to';`,
		postgres.ListTableDiffChangesQuery(`odd"name`, fromExpr, toExpr))
	require.Equal(t, "HASHOF('a''b')", postgres.HashOfFunction("a'b"))
	require.Equal(t, "SELECT dolt_branch('a''b');", postgres.CallProcedure(DoltBranch, "a'b"))

	lite := NewDoltLiteDialect()
	require.Equal(t,
		`SELECT * FROM "dolt_diff_odd""name"('from', 'to');`,
		lite.ListTableDiffChangesQuery(`odd"name`, fromExpr, toExpr))
}

func TestDoltLiteDialectSupportsTool(t *testing.T) {
	d := NewDoltLiteDialect()

	unsupported := []string{
		"list_databases",
		"create_database",
		"drop_database",
		"clone_database",
		"show_processlist",
		"kill_process",
	}
	for _, name := range unsupported {
		require.False(t, d.SupportsTool(name), "expected %s to be unsupported", name)
	}

	supported := []string{
		"query",
		"exec",
		"show_tables",
		"create_dolt_commit",
		"list_dolt_branches",
		"merge_dolt_branch",
		"dolt_push_branch",
		"dolt_pull_branch",
		"add_dolt_remote",
		"run_dolt_tests",
		"add_dolt_test",
		"remove_dolt_test",
		"get_dolt_merge_status",
	}
	for _, name := range supported {
		require.True(t, d.SupportsTool(name), "expected %s to be supported", name)
	}
}

func TestDoltLiteDialectValidateReadQuery(t *testing.T) {
	d := NewDoltLiteDialect()

	valid := []string{
		"SELECT * FROM people;",
		"  select 1",
		"WITH t AS (SELECT 1) SELECT * FROM t;",
		"EXPLAIN SELECT 1;",
		"-- comment\nSELECT 1;",
		"/* comment */ SELECT 1;",
		"SELECT * FROM dolt_log;",
		"SELECT * FROM dolt_branches;",
		"SELECT active_branch();",
		"SELECT dolt_merge_base('main', 'feature');",
		"SELECT dolt_hashof('HEAD');",
	}
	for _, q := range valid {
		require.NoError(t, d.ValidateReadQuery(q), "expected valid read query: %s", q)
	}

	invalid := []string{
		"INSERT INTO people VALUES (1);",
		"UPDATE people SET name = 'x';",
		"DELETE FROM people;",
		"DROP TABLE people;",
		"",
		"SELECT dolt_commit('-m', 'sneaky');",
		"SELECT dolt_checkout('main');",
		"SELECT dolt_reset('--hard');",
		"select DOLT_MERGE('feature');",
		"SELECT 1; DELETE FROM people;",
		"SELECT 1; SELECT 2;",
	}
	for _, q := range invalid {
		require.Error(t, d.ValidateReadQuery(q), "expected invalid read query: %s", q)
	}
}

func TestDoltLiteDialectReadQueryAllowsSemicolonsInTrivia(t *testing.T) {
	d := NewDoltLiteDialect()

	require.NoError(t, d.ValidateReadQuery("SELECT ';' AS value;"))
	require.NoError(t, d.ValidateReadQuery("SELECT 1; -- trailing ; comment\n"))
	require.NoError(t, d.ValidateReadQuery("SELECT 'it''s; fine'; /* ; */"))
}

func TestDoltLiteDialectValidateWriteQuery(t *testing.T) {
	d := NewDoltLiteDialect()

	valid := []string{
		"INSERT INTO people VALUES (1);",
		"UPDATE people SET name = 'x';",
		"DELETE FROM people;",
		"REPLACE INTO people VALUES (1);",
		"SELECT dolt_commit('-Am', 'commit through exec');",
		"SELECT dolt_creds_new();",
	}
	for _, q := range valid {
		require.NoError(t, d.ValidateWriteQuery(q), "expected valid write query: %s", q)
	}

	invalid := []string{
		"SELECT * FROM people;",
		"WITH t AS (SELECT 1) SELECT * FROM t;",
		"",
	}
	for _, q := range invalid {
		require.Error(t, d.ValidateWriteQuery(q), "expected invalid write query: %s", q)
	}
}

func TestDoltLiteDialectValidateDDLQueries(t *testing.T) {
	d := NewDoltLiteDialect()

	require.NoError(t, d.ValidateCreateTableQuery("CREATE TABLE t (id INTEGER PRIMARY KEY);"))
	require.Error(t, d.ValidateCreateTableQuery("SELECT 1;"))
	require.Error(t, d.ValidateCreateTableQuery("DROP TABLE t;"))
	require.Error(t, d.ValidateCreateTableQuery("CREATE INDEX idx ON t(id);"))

	require.NoError(t, d.ValidateAlterTableQuery("ALTER TABLE t ADD COLUMN c TEXT;"))
	require.Error(t, d.ValidateAlterTableQuery("CREATE TABLE t (id INTEGER);"))
	require.Error(t, d.ValidateAlterTableQuery("SELECT 1;"))
}
