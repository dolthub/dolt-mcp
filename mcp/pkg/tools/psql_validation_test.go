package tools

import (
	"testing"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
)

func TestPostgresValidateReadQuery_AcceptsReadStatements(t *testing.T) {
	dialect := db.NewDialect(db.DialectPostgres)
	cases := []string{
		"SHOW TABLES",
		"SHOW DATABASES",
		"EXPLAIN SELECT 1",
		"SELECT 1 UNION SELECT 2",
		"SELECT * FROM foo",
		"SELECT dolt_version()",
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateReadQuery(sql); err != nil {
				t.Fatalf("expected read validation to pass, got err=%v", err)
			}
		})
	}
}

func TestPostgresValidateReadQuery_RejectsNonReadStatements(t *testing.T) {
	dialect := db.NewDialect(db.DialectPostgres)
	cases := []string{
		"INSERT INTO t VALUES (1)",
		"UPDATE t SET id = 2",
		"DELETE FROM t",
		`CREATE TABLE t (id int)`,
		`ALTER TABLE t ADD COLUMN c int`,
		`DROP TABLE t`,
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateReadQuery(sql); err == nil {
				t.Fatalf("expected read validation to fail, got nil")
			}
		})
	}
}

func TestPostgresValidateWriteQuery_AcceptsNonReadStatements(t *testing.T) {
	dialect := db.NewDialect(db.DialectPostgres)
	cases := []string{
		"INSERT INTO t VALUES (1)",
		"UPDATE t SET id = 2",
		"DELETE FROM t",
		`CREATE TABLE t (id int)`,
		`ALTER TABLE t ADD COLUMN c int`,
		`DROP TABLE t`,
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateWriteQuery(sql); err != nil {
				t.Fatalf("expected write validation to pass, got err=%v", err)
			}
		})
	}
}

func TestPostgresValidateWriteQuery_RejectsReadStatements(t *testing.T) {
	dialect := db.NewDialect(db.DialectPostgres)
	cases := []string{
		"SELECT 1",
		"SHOW TABLES",
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

func TestPostgresValidateCreateTableQuery_AcceptsCreateTable(t *testing.T) {
	dialect := db.NewDialect(db.DialectPostgres)
	cases := []string{
		"CREATE TABLE t (id int)",
		`CREATE TABLE "t" (id int PRIMARY KEY, name varchar(100))`,
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateCreateTableQuery(sql); err != nil {
				t.Fatalf("expected create table validation to pass, got err=%v", err)
			}
		})
	}
}

func TestPostgresValidateCreateTableQuery_RejectsNonCreateTable(t *testing.T) {
	dialect := db.NewDialect(db.DialectPostgres)
	cases := []string{
		"SELECT 1",
		"INSERT INTO t VALUES (1)",
		"ALTER TABLE t ADD COLUMN c int",
		"DROP TABLE t",
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateCreateTableQuery(sql); err == nil {
				t.Fatalf("expected create table validation to fail, got nil")
			}
		})
	}
}

func TestPostgresValidateAlterTableQuery_AcceptsAlterTable(t *testing.T) {
	dialect := db.NewDialect(db.DialectPostgres)
	cases := []string{
		"ALTER TABLE t ADD COLUMN c int",
		`ALTER TABLE "t" DROP COLUMN c`,
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateAlterTableQuery(sql); err != nil {
				t.Fatalf("expected alter table validation to pass, got err=%v", err)
			}
		})
	}
}

func TestPostgresValidateAlterTableQuery_RejectsNonAlterTable(t *testing.T) {
	dialect := db.NewDialect(db.DialectPostgres)
	cases := []string{
		"SELECT 1",
		"INSERT INTO t VALUES (1)",
		"CREATE TABLE t (id int)",
		"DROP TABLE t",
	}

	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			if err := dialect.ValidateAlterTableQuery(sql); err == nil {
				t.Fatalf("expected alter table validation to fail, got nil")
			}
		})
	}
}