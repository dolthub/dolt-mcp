package tools

import (
	"context"
	"fmt"
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

func NewDatabaseTransactionOnBranch(ctx context.Context, config db.Config, branch string) (db.DatabaseTransaction, error) {
	tx, err := db.NewDatabaseTransaction(ctx, config)
	if err != nil {
		return nil, err
	}

	err = tx.ExecContext(ctx, fmt.Sprintf(DoltCheckoutWorkingBranchSQLQueryFormatString, branch))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func NewDatabaseTransactionUsingDatabase(ctx context.Context, config db.Config, database string) (db.DatabaseTransaction, error) {
	tx, err := db.NewDatabaseTransaction(ctx, config)
	if err != nil {
		return nil, err
	}

	err = tx.ExecContext(ctx, fmt.Sprintf(DoltUseWorkingDatabaseSQLQueryFormatString, database))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func NewDatabaseTransactionUsingDatabaseOnBranch(ctx context.Context, config db.Config, database, branch string) (db.DatabaseTransaction, error) {
	tx, err := db.NewDatabaseTransaction(ctx, config)
	if err != nil {
		fmt.Println("DUSTIN: NewDatabaseTransactionOnBranchUsingDatabase: NewDatabaseTransaction: error:", err.Error())
		return nil, err
	}

	err = tx.ExecContext(ctx, fmt.Sprintf(DoltUseWorkingDatabaseSQLQueryFormatString, database))
	if err != nil {
		fmt.Println("DUSTIN: NewDatabaseTransactionOnBranchUsingDatabase: Use Database: error:", err.Error())
		return nil, err
	}

	err = tx.ExecContext(ctx, fmt.Sprintf(DoltCheckoutWorkingBranchSQLQueryFormatString, branch))
	if err != nil {
		fmt.Println("DUSTIN: DSN:", config.GetDSN())
		fmt.Println("DUSTIN: NewDatabaseTransactionOnBranchUsingDatabase: CheckoutWorkingBranch: error:", err.Error())
		return nil, err
	}

	return tx, nil
}
