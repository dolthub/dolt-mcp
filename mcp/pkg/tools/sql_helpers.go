package tools

import (
	"context"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
)

func ParseSQLQuery(query string) (sqlparser.Statement, error) {
	sqlCtx := sql.NewEmptyContext()
	sqlMode := sql.LoadSqlMode(sqlCtx)
	return sqlparser.ParseWithOptions(sqlCtx, query, sqlMode.ParserOptions())
}

func CommitTransactionOrRollbackOnError(ctx context.Context, tx db.DatabaseTransaction, err error) error {
	if err == nil {
		return tx.Commit(ctx)
	}
	tx.Rollback(ctx)
	return err
}

