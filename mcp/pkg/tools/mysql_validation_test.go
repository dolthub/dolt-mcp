package tools

import (
	"testing"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
)

func TestValidateReadQuery_AcceptsReadStatements(t *testing.T) {
	dialect := db.NewDialect(db.DialectMySQL)
	cases := []string{
		"SHOW PROCESSLIST",
		"SHOW TABLES",
		"EXPLAIN SELECT 1",
		"SELECT 1 UNION SELECT 2",
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateReadQuery(sql); err != nil {
				t.Fatalf("expected read validation to pass, got err=%v", err)
			}
		})
	}
}

func TestValidateReadQuery_RejectsNonReadStatements(t *testing.T) {
	dialect := db.NewDialect(db.DialectMySQL)
	cases := []string{
		"KILL 123",
		"INSERT INTO t VALUES (1)",
		"CREATE TABLE t (id int)",
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateReadQuery(sql); err == nil {
				t.Fatalf("expected read validation to fail, got nil")
			}
		})
	}
}

func TestValidateWriteQuery_AcceptsNonReadStatements(t *testing.T) {
	dialect := db.NewDialect(db.DialectMySQL)
	cases := []string{
		"KILL 123",
		"INSERT INTO t VALUES (1)",
		"UPDATE t SET id = 2",
		"DELETE FROM t",
		"CREATE TABLE t (id int)",
		"ALTER TABLE t ADD COLUMN c int",
		"DROP TABLE t",
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateWriteQuery(sql); err != nil {
				t.Fatalf("expected write validation to pass, got err=%v", err)
			}
		})
	}
}

func TestValidateWriteQuery_RejectsReadStatements(t *testing.T) {
	dialect := db.NewDialect(db.DialectMySQL)
	cases := []string{
		"SELECT 1",
		"SHOW PROCESSLIST",
		"EXPLAIN SELECT 1",
		"SELECT 1 UNION SELECT 2",
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateWriteQuery(sql); err == nil {
				t.Fatalf("expected write validation to fail, got nil")
			}
		})
	}
}
