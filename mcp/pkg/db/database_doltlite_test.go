//go:build doltlite

package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newPreparedDoltLiteTestConfig(t *testing.T) Config {
	t.Helper()
	config := Config{
		DialectType: DialectDoltLite,
		Path:        t.TempDir() + "/database.db",
	}
	require.NoError(t, PrepareDatabase(&config))
	t.Cleanup(func() { require.NoError(t, CloseDatabase(config)) })
	return config
}

func TestDoltLiteCanceledTransactionDoesNotBlockNextOperation(t *testing.T) {
	config := newPreparedDoltLiteTestConfig(t)

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
			finishErr := test.finish(tx, canceledCtx)
			if finishErr != nil {
				require.ErrorIs(t, finishErr, context.Canceled)
			}

			next, err := NewDatabaseTransaction(context.Background(), config)
			require.NoError(t, err)
			require.NoError(t, next.Rollback(context.Background()))
		})
	}
}

func TestDoltLiteTransactionsUseIndependentHandles(t *testing.T) {
	config := newPreparedDoltLiteTestConfig(t)

	first, err := NewDatabaseTransaction(context.Background(), config)
	require.NoError(t, err)
	t.Cleanup(func() {
		if first != nil {
			_ = first.Rollback(context.Background())
		}
	})

	type transactionResult struct {
		tx  DatabaseTransaction
		err error
	}
	secondResult := make(chan transactionResult, 1)
	go func() {
		tx, txErr := NewDatabaseTransaction(context.Background(), config)
		secondResult <- transactionResult{tx: tx, err: txErr}
	}()

	var second DatabaseTransaction
	select {
	case result := <-secondResult:
		require.NoError(t, result.err)
		second = result.tx
	case <-time.After(2 * time.Second):
		t.Fatal("second DoltLite transaction blocked behind the first")
	}
	require.NoError(t, second.Rollback(context.Background()))
	require.NoError(t, first.Rollback(context.Background()))
	first = nil
}

func TestDoltLiteCoordinatesConcurrentWriters(t *testing.T) {
	config := newPreparedDoltLiteTestConfig(t)
	ctx := context.Background()

	setup, err := NewDatabaseTransaction(ctx, config)
	require.NoError(t, err)
	require.NoError(t, setup.ExecContext(ctx, "CREATE TABLE concurrent_writes (id INTEGER PRIMARY KEY);"))
	require.NoError(t, setup.Commit(ctx))

	first, err := NewDatabaseTransaction(ctx, config)
	require.NoError(t, err)
	require.NoError(t, first.ExecContext(ctx, "INSERT INTO concurrent_writes VALUES (1);"))

	second, err := NewDatabaseTransaction(ctx, config)
	require.NoError(t, err)
	secondWrite := make(chan error, 1)
	go func() {
		secondWrite <- second.ExecContext(ctx, "INSERT INTO concurrent_writes VALUES (2);")
	}()

	time.Sleep(100 * time.Millisecond)
	require.NoError(t, first.Commit(ctx))
	require.NoError(t, <-secondWrite)
	require.NoError(t, second.Commit(ctx))

	verify, err := NewDatabaseTransaction(ctx, config)
	require.NoError(t, err)
	result, err := verify.QueryContext(ctx, "SELECT count(*) AS count FROM concurrent_writes;", ResultFormatCSV)
	require.NoError(t, err)
	require.Equal(t, "count\n2\n", result)
	require.NoError(t, verify.Rollback(ctx))
}
