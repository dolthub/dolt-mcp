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
		return nil, err
	}

	return tx, nil
}

func NewDatabaseTransactionUsingDatabase(ctx context.Context, config db.Config, dialect db.Dialect, database string) (db.DatabaseTransaction, error) {
	tx, err := db.NewDatabaseTransaction(ctx, config)
	if err != nil {
		return nil, err
	}

	err = tx.ExecContext(ctx, dialect.UseDatabase(database))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func NewDatabaseTransactionUsingDatabaseOnBranch(ctx context.Context, config db.Config, dialect db.Dialect, database, branch string) (db.DatabaseTransaction, error) {
	tx, err := db.NewDatabaseTransaction(ctx, config)
	if err != nil {
		return nil, err
	}

	err = tx.ExecContext(ctx, dialect.UseDatabase(database))
	if err != nil {
		return nil, err
	}

	err = tx.ExecContext(ctx, dialect.CallProcedure(db.DoltCheckout, branch))
	if err != nil {
		return nil, err
	}

	return tx, nil
}