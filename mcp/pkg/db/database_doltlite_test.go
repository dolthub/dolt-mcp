//go:build doltlite

package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// A canceled request must not leave the shared single-connection handle in
// an open transaction. Both Commit and Rollback use an independent cleanup
// context before the handle is made available to the next tool call.
func TestDoltLiteCanceledTransactionDoesNotPoisonSharedHandle(t *testing.T) {
	config := Config{
		DialectType: DialectDoltLite,
		Path:        t.TempDir() + "/canceled-transaction.db",
	}
	require.NoError(t, PrepareDatabase(config))
	t.Cleanup(func() { require.NoError(t, CloseDatabase(config)) })

	finishers := []struct {
		name   string
		finish func(DatabaseTransaction, context.Context) error
	}{
		{name: "rollback", finish: func(tx DatabaseTransaction, ctx context.Context) error {
			return tx.Rollback(ctx)
		}},
		{name: "commit", finish: func(tx DatabaseTransaction, ctx context.Context) error {
			return tx.Commit(ctx)
		}},
	}

	for _, test := range finishers {
		t.Run(test.name, func(t *testing.T) {
			tx, err := NewDatabaseTransaction(context.Background(), config)
			require.NoError(t, err)

			canceledCtx, cancel := context.WithCancel(context.Background())
			cancel()
			require.ErrorIs(t, test.finish(tx, canceledCtx), context.Canceled)

			// Beginning and finishing another transaction proves the shared
			// connection was returned to autocommit mode after cancellation.
			next, err := NewDatabaseTransaction(context.Background(), config)
			require.NoError(t, err)
			require.NoError(t, next.Rollback(context.Background()))
		})
	}
}
