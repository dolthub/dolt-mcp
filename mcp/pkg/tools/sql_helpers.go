package tools

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
)

func ParseSQLQuery(query string) (sqlparser.Statement, error) {
	sqlCtx := sql.NewEmptyContext()
	sqlMode := sql.LoadSqlMode(sqlCtx)
	return sqlparser.ParseWithOptions(sqlCtx, query, sqlMode.ParserOptions())
}

