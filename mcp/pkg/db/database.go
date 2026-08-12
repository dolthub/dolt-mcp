package db

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type ResultFormat int

const (
	ResultFormatUndefined = iota
	ResultFormatMarkdown
	ResultFormatCSV
)

var ErrUnsupportedResultFormat = errors.New("unsupported result format")
var ErrTransactionHasBeenCommittedOrRolledBack = errors.New("transaction has already been committed or rolled back")

type RowMap map[string]interface{}
type Columns []string

type DatabaseTransaction interface {
	QueryContext(ctx context.Context, query string, resultFormat ResultFormat) (string, error)
	ExecContext(ctx context.Context, query string) error
	Rollback(ctx context.Context) error
	Commit(ctx context.Context) error
}

type databaseTransactionImpl struct {
	db     *sql.DB
	config Config
	// keepOpen indicates the *sql.DB is a shared long-lived handle that must
	// not be closed when the transaction finishes (used by embedded DoltLite).
	keepOpen bool
	// release is invoked exactly once when the transaction finishes.
	release func()
	// handle is set for shared DoltLite transactions so a connection which
	// cannot be returned to a clean state can be evicted from the cache.
	handle *doltLiteHandle
}

var _ DatabaseTransaction = &databaseTransactionImpl{}

func NewDatabaseTransaction(ctx context.Context, config Config) (DatabaseTransaction, error) {
	if config.DialectType == DialectDoltLite {
		return newDoltLiteTransaction(ctx, config)
	}

	db, err := newDB(config)
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, "BEGIN;")
	if err != nil {
		db.Close()
		return nil, err
	}
	return &databaseTransactionImpl{
		db:     db,
		config: config,
	}, nil
}

// newDoltLiteTransaction begins a transaction on the shared DoltLite handle.
// DoltLite session state (checked-out branch, merge/rebase state, conflict
// catalogs) lives on the underlying connection, so all tool calls share one
// long-lived single-connection handle; a per-handle lock serializes
// transactions since BEGIN/COMMIT are issued as plain statements.
func newDoltLiteTransaction(ctx context.Context, config Config) (DatabaseTransaction, error) {
	var handle *doltLiteHandle
	var err error
	for {
		handle, err = getDoltLiteHandle(config)
		if err != nil {
			return nil, err
		}

		handle.mu.Lock()
		// CloseDatabase or a failed transaction may have retired this handle
		// after it was fetched from the map but before its lock was acquired.
		if !handle.closed {
			break
		}
		handle.mu.Unlock()
	}

	if config.CommitName != "" {
		if _, err = handle.db.ExecContext(ctx, "SELECT dolt_config('user.name', ?);", config.CommitName); err != nil {
			handle.mu.Unlock()
			return nil, fmt.Errorf("failed to configure DoltLite commit author name: %w", err)
		}
	}
	if config.CommitEmail != "" {
		if _, err = handle.db.ExecContext(ctx, "SELECT dolt_config('user.email', ?);", config.CommitEmail); err != nil {
			handle.mu.Unlock()
			return nil, fmt.Errorf("failed to configure DoltLite commit author email: %w", err)
		}
	}
	_, err = handle.db.ExecContext(ctx, "BEGIN;")
	if err != nil {
		closeErr := discardDoltLiteHandle(config, handle)
		handle.mu.Unlock()
		return nil, errors.Join(err, closeErr)
	}
	return &databaseTransactionImpl{
		db:       handle.db,
		config:   config,
		keepOpen: true,
		release:  handle.mu.Unlock,
		handle:   handle,
	}, nil
}

func (d *databaseTransactionImpl) QueryContext(ctx context.Context, query string, resultFormat ResultFormat) (string, error) {
	rowMap, columns, err := d.doQueryContext(ctx, query)
	if err != nil {
		return "", err
	}
	switch resultFormat {
	case ResultFormatMarkdown:
		return d.rowMapToMarkdown(rowMap, columns)
	case ResultFormatCSV:
		return d.rowMapToCSV(rowMap, columns)
	default:
		return "", ErrUnsupportedResultFormat
	}
}

func (d *databaseTransactionImpl) rowMapToMarkdown(rowMaps []RowMap, headers []string) (string, error) {
	var mdBuf strings.Builder

	// Write header row
	for i, header := range headers {
		if i > 0 {
			mdBuf.WriteString(" | ")
		}
		mdBuf.WriteString(header)
	}
	mdBuf.WriteString("\n")

	// Write separator row
	for i := range headers {
		if i > 0 {
			mdBuf.WriteString(" | ")
		}
		mdBuf.WriteString("---")
	}
	mdBuf.WriteString("\n")

	// Write data rows
	for _, rowMap := range rowMaps {
		for i, header := range headers {
			if i > 0 {
				mdBuf.WriteString(" | ")
			}
			value, exists := rowMap[header]
			if !exists {
				return "", fmt.Errorf("key '%s' not found in map", header)
			}
			mdBuf.WriteString(fmt.Sprintf("%v", value))
		}
		mdBuf.WriteString("\n")
	}

	return mdBuf.String(), nil
}

func (d *databaseTransactionImpl) rowMapToCSV(rowMaps []RowMap, headers []string) (string, error) {
	var csvBuf strings.Builder
	writer := csv.NewWriter(&csvBuf)

	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write headers: %v", err)
	}

	for _, rowMap := range rowMaps {
		row := make([]string, len(headers))
		for i, header := range headers {
			value, exists := rowMap[header]
			if !exists {
				return "", fmt.Errorf("key '%s' not found in map", header)
			}
			row[i] = fmt.Sprintf("%v", value)
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write row: %v", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("error flushing CSV writer: %v", err)
	}

	return csvBuf.String(), nil
}

func (d *databaseTransactionImpl) doQueryContext(ctx context.Context, query string) ([]RowMap, Columns, error) {
	if d.db == nil {
		return nil, nil, ErrTransactionHasBeenCommittedOrRolledBack
	}

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	rowMaps := []RowMap{}
	for rows.Next() {
		// Create a slice of interface{}'s to hold each column value
		values := make([]interface{}, len(columns))

		// Create a slice of pointers to each value in values
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for readability
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}

		rowMaps = append(rowMaps, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return rowMaps, columns, nil
}

func (d *databaseTransactionImpl) doExecContext(ctx context.Context, query string) error {
	if d.db == nil {
		return ErrTransactionHasBeenCommittedOrRolledBack
	}
	_, err := d.db.ExecContext(ctx, query)
	return err
}

func (d *databaseTransactionImpl) ExecContext(ctx context.Context, query string) error {
	return d.doExecContext(ctx, query)
}

// isNoActiveTransactionError reports whether err is SQLite's "cannot commit
// - no transaction is active" (or the rollback equivalent). DoltLite
// version-control functions (dolt_commit, dolt_merge, ...) seal the
// enclosing SQL transaction as a side effect, so the transaction this type
// opened may already be gone by the time Commit or Rollback runs; that is
// expected, not an error.
func isNoActiveTransactionError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no transaction is active")
}

const doltLiteTransactionCleanupTimeout = 5 * time.Second

// recoverSharedTransaction returns a shared DoltLite connection to a clean
// state after COMMIT or ROLLBACK failed with the request context. Cleanup is
// deliberately independent of that context because it is commonly already
// canceled. If cleanup also fails, the handle is evicted and closed so the
// next tool call opens a fresh connection instead of inheriting a transaction.
func (d *databaseTransactionImpl) recoverSharedTransaction(originalErr error) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), doltLiteTransactionCleanupTimeout)
	defer cancel()

	_, cleanupErr := d.db.ExecContext(cleanupCtx, "ROLLBACK;")
	if cleanupErr == nil || isNoActiveTransactionError(cleanupErr) {
		return originalErr
	}

	discardErr := discardDoltLiteHandle(d.config, d.handle)
	return errors.Join(
		originalErr,
		fmt.Errorf("failed to clean up DoltLite transaction: %w", cleanupErr),
		discardErr,
	)
}

// finish releases the transaction's resources: shared handles are unlocked
// but kept open, per-transaction handles are closed.
func (d *databaseTransactionImpl) finish(err error) error {
	if !d.keepOpen {
		cerr := d.db.Close()
		if err == nil {
			err = cerr
		}
	}
	d.db = nil
	if d.release != nil {
		d.release()
		d.release = nil
	}
	return err
}

func (d *databaseTransactionImpl) Rollback(ctx context.Context) (err error) {
	if d.db == nil {
		err = ErrTransactionHasBeenCommittedOrRolledBack
		return
	}

	defer func() {
		err = d.finish(err)
	}()

	err = d.doExecContext(ctx, "ROLLBACK;")
	if d.keepOpen && isNoActiveTransactionError(err) {
		err = nil
	} else if d.keepOpen && err != nil {
		err = d.recoverSharedTransaction(err)
	}
	if err != nil {
		return
	}

	return
}

func (d *databaseTransactionImpl) Commit(ctx context.Context) (err error) {
	if d.db == nil {
		err = ErrTransactionHasBeenCommittedOrRolledBack
		return
	}

	defer func() {
		err = d.finish(err)
	}()

	err = d.doExecContext(ctx, "COMMIT;")
	if d.keepOpen && isNoActiveTransactionError(err) {
		err = nil
	} else if d.keepOpen && err != nil {
		err = d.recoverSharedTransaction(err)
	}
	if err != nil {
		return
	}

	return
}

// doltLiteHandle is a long-lived handle on an embedded DoltLite database
// file. The pool is capped at a single connection so DoltLite's
// per-connection session state survives across statements, and mu serializes
// whole transactions on that connection.
type doltLiteHandle struct {
	db     *sql.DB
	mu     sync.Mutex
	closed bool
}

var (
	doltLiteHandlesMu sync.Mutex
	doltLiteHandles   = map[string]*doltLiteHandle{}
)

// PrepareDatabase eagerly opens and verifies embedded databases. Networked
// dialects retain their existing lazy connection behavior.
func PrepareDatabase(config Config) error {
	if config.DialectType != DialectDoltLite {
		return nil
	}
	if err := config.Validate(); err != nil {
		return err
	}
	_, err := getDoltLiteHandle(config)
	return err
}

// CloseDatabase closes a shared embedded database handle. It is a no-op for
// networked dialects, whose per-tool pools already close with transactions.
func CloseDatabase(config Config) error {
	if config.DialectType != DialectDoltLite {
		return nil
	}

	dsn := NewDialect(config.DialectType).FormatDSN(config)
	doltLiteHandlesMu.Lock()
	handle, ok := doltLiteHandles[dsn]
	if ok {
		delete(doltLiteHandles, dsn)
	}
	doltLiteHandlesMu.Unlock()
	if !ok {
		return nil
	}

	handle.mu.Lock()
	defer handle.mu.Unlock()
	handle.closed = true
	return handle.db.Close()
}

// discardDoltLiteHandle removes and closes handle while its transaction lock
// is held. A goroutine which fetched the old handle before removal observes
// closed after acquiring that lock and retries with a fresh handle.
func discardDoltLiteHandle(config Config, handle *doltLiteHandle) error {
	if handle == nil || handle.closed {
		return nil
	}

	dsn := NewDialect(config.DialectType).FormatDSN(config)
	doltLiteHandlesMu.Lock()
	if current, ok := doltLiteHandles[dsn]; ok && current == handle {
		delete(doltLiteHandles, dsn)
	}
	doltLiteHandlesMu.Unlock()

	handle.closed = true
	if err := handle.db.Close(); err != nil {
		return fmt.Errorf("failed to close unusable DoltLite connection: %w", err)
	}
	return nil
}

func getDoltLiteHandle(config Config) (*doltLiteHandle, error) {
	dialect := NewDialect(config.DialectType)
	dsn := dialect.FormatDSN(config)

	doltLiteHandlesMu.Lock()
	defer doltLiteHandlesMu.Unlock()

	if handle, ok := doltLiteHandles[dsn]; ok {
		return handle, nil
	}

	// Registers the cgo-backed driver; fails when the binary was built
	// without the "doltlite" build tag.
	if err := registerDoltLiteDriver(config); err != nil {
		return nil, err
	}

	db, err := sql.Open(doltLiteDriverName, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to open DoltLite database %s (the file may have been written by an incompatible DoltLite version): %w", dsn, err)
	}

	// Verify the linked library is actually DoltLite and not stock SQLite.
	var engine string
	if err := db.QueryRow("SELECT doltlite_engine();").Scan(&engine); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to verify DoltLite engine (is the binary linked against libdoltlite?): %w", err)
	}
	if engine != "prolly" {
		db.Close()
		return nil, fmt.Errorf("unexpected DoltLite storage engine %q, expected \"prolly\"", engine)
	}

	handle := &doltLiteHandle{db: db}
	doltLiteHandles[dsn] = handle
	return handle, nil
}

func newDB(config Config) (*sql.DB, error) {
	dialect := NewDialect(config.DialectType)

	if err := dialect.ConfigureTLS(&config); err != nil {
		return nil, err
	}

	dsn := dialect.FormatDSN(config)

	db, err := sql.Open(dialect.DriverName(), dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
