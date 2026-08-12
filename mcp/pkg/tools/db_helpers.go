package tools

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
)

func CommitTransactionOrRollbackOnError(ctx context.Context, tx db.DatabaseTransaction, err error) error {
	if err == nil {
		return tx.Commit(ctx)
	}
	tx.Rollback(ctx)
	return err
}

func NewDatabaseTransactionOnBranch(ctx context.Context, config db.Config, dialect db.Dialect, branch string) (db.DatabaseTransaction, error) {
	tx, err := db.NewDatabaseTransaction(ctx, config)
	if err != nil {
		return nil, err
	}

	err = tx.ExecContext(ctx, dialect.CallProcedure(db.DoltCheckout, branch))
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	return tx, nil
}

func NewDatabaseTransactionUsingDatabase(ctx context.Context, config db.Config, dialect db.Dialect, database string) (db.DatabaseTransaction, error) {
	tx, err := db.NewDatabaseTransaction(ctx, config)
	if err != nil {
		return nil, err
	}

	// Single-database dialects have no USE statement.
	if useStmt := dialect.UseDatabase(database); useStmt != "" {
		err = tx.ExecContext(ctx, useStmt)
		if err != nil {
			tx.Rollback(ctx)
			return nil, err
		}
	}

	return tx, nil
}

func NewDatabaseTransactionUsingDatabaseOnBranch(ctx context.Context, config db.Config, dialect db.Dialect, database, branch string) (db.DatabaseTransaction, error) {
	tx, err := NewDatabaseTransactionUsingDatabase(ctx, config, dialect, database)
	if err != nil {
		return nil, err
	}

	err = tx.ExecContext(ctx, dialect.CallProcedure(db.DoltCheckout, branch))
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	return tx, nil
}
