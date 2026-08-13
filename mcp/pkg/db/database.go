package db

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
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

type sqlExecutor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type databaseTransactionImpl struct {
	executor sqlExecutor
	// db is set for the existing network-dialect implementation, which owns a
	// per-transaction *sql.DB and closes it when the transaction finishes.
	db *sql.DB
	// conn is a DoltLite handle pinned to this MCP operation. Every statement
	// from BEGIN through COMMIT / ROLLBACK runs on this exact sqlite3 handle.
	conn *sql.Conn
	// doltLiteDatabase is the pool that supplied conn. An unprepared Config
	// creates a temporary pool so direct callers retain the old API behavior.
	doltLiteDatabase *doltLiteDatabase
	closeDoltLiteDB  bool
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
		executor: db,
		db:       db,
	}, nil
}

const doltLiteBusyTimeout = 5 * time.Second

// newDoltLiteTransaction pins a dedicated DoltLite handle for one MCP tool
// operation. Concurrent operations use independent sqlite3 handles and rely
// on DoltLite's own reader/writer and snapshot locking for the shared file.
// Pinning is still required: branch selection and SQL transaction state live
// on a handle, so BEGIN, checkout, the tool SQL, and COMMIT must not move
// between database/sql pooled connections.
func newDoltLiteTransaction(ctx context.Context, config Config) (DatabaseTransaction, error) {
	database := config.doltLiteDatabase
	closeDatabase := false
	if database == nil {
		var err error
		database, err = openDoltLiteDatabase(config)
		if err != nil {
			return nil, err
		}
		closeDatabase = true
	}

	conn, err := database.db.Conn(ctx)
	if err != nil {
		if closeDatabase {
			_ = database.db.Close()
		}
		return nil, err
	}

	tx := &databaseTransactionImpl{
		executor:         conn,
		conn:             conn,
		doltLiteDatabase: database,
		closeDoltLiteDB:  closeDatabase,
	}

	// DoltLite uses its busy handler while another process or handle owns the
	// file's writer lock. This lets short Workbench/MCP overlaps resolve without
	// serializing all MCP operations in Go.
	if _, err = conn.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout = %d;", doltLiteBusyTimeout.Milliseconds())); err != nil {
		return nil, tx.finish(err)
	}
	if config.CommitName != "" {
		if _, err = conn.ExecContext(ctx, "SELECT dolt_config('user.name', ?);", config.CommitName); err != nil {
			return nil, tx.finish(fmt.Errorf("failed to configure DoltLite commit author name: %w", err))
		}
	}
	if config.CommitEmail != "" {
		if _, err = conn.ExecContext(ctx, "SELECT dolt_config('user.email', ?);", config.CommitEmail); err != nil {
			return nil, tx.finish(fmt.Errorf("failed to configure DoltLite commit author email: %w", err))
		}
	}
	_, err = conn.ExecContext(ctx, "BEGIN;")
	if err != nil {
		err = tx.recoverPinnedTransaction(err)
		return nil, tx.finish(err)
	}
	return tx, nil
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
	if d.executor == nil {
		return nil, nil, ErrTransactionHasBeenCommittedOrRolledBack
	}

	rows, err := d.executor.QueryContext(ctx, query)
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
	if d.executor == nil {
		return ErrTransactionHasBeenCommittedOrRolledBack
	}
	_, err := d.executor.ExecContext(ctx, query)
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

// recoverPinnedTransaction returns a DoltLite handle to a clean state after
// BEGIN, COMMIT, or ROLLBACK failed with the request context. Cleanup uses an
// independent context because the request is commonly already canceled. If
// rollback also fails, finish still closes this operation's physical handle;
// the pool retains no idle DoltLite handles, so it cannot poison a later call.
func (d *databaseTransactionImpl) recoverPinnedTransaction(originalErr error) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), doltLiteTransactionCleanupTimeout)
	defer cancel()

	_, cleanupErr := d.conn.ExecContext(cleanupCtx, "ROLLBACK;")
	if cleanupErr == nil || isNoActiveTransactionError(cleanupErr) {
		return originalErr
	}

	return errors.Join(
		originalErr,
		fmt.Errorf("failed to clean up DoltLite transaction: %w", cleanupErr),
	)
}

// finish releases the transaction's resources. DoltLite handles return to
// the pool (or are closed when MaxIdleConns is zero); network dialects retain
// their existing per-transaction *sql.DB lifecycle.
func (d *databaseTransactionImpl) finish(err error) error {
	if d.conn != nil {
		cerr := d.conn.Close()
		if err == nil {
			err = cerr
		}
	}
	if d.db != nil {
		cerr := d.db.Close()
		if err == nil {
			err = cerr
		}
	}
	if d.closeDoltLiteDB && d.doltLiteDatabase != nil {
		cerr := d.doltLiteDatabase.db.Close()
		if err == nil {
			err = cerr
		}
	}
	d.executor = nil
	d.conn = nil
	d.db = nil
	d.doltLiteDatabase = nil
	return err
}

func (d *databaseTransactionImpl) Rollback(ctx context.Context) (err error) {
	if d.executor == nil {
		err = ErrTransactionHasBeenCommittedOrRolledBack
		return
	}

	defer func() {
		err = d.finish(err)
	}()

	err = d.doExecContext(ctx, "ROLLBACK;")
	if d.conn != nil && isNoActiveTransactionError(err) {
		err = nil
	} else if d.conn != nil && err != nil {
		err = d.recoverPinnedTransaction(err)
	}
	if err != nil {
		return
	}

	return
}

func (d *databaseTransactionImpl) Commit(ctx context.Context) (err error) {
	if d.executor == nil {
		err = ErrTransactionHasBeenCommittedOrRolledBack
		return
	}

	defer func() {
		err = d.finish(err)
	}()

	err = d.doExecContext(ctx, "COMMIT;")
	if d.conn != nil && isNoActiveTransactionError(err) {
		err = nil
	} else if d.conn != nil && err != nil {
		err = d.recoverPinnedTransaction(err)
	}
	if err != nil {
		return
	}

	return
}

// doltLiteDatabase owns the database/sql pool for one MCP server. Each tool
// operation pins one independent sqlite3 handle from this pool.
type doltLiteDatabase struct {
	db *sql.DB
}

// PrepareDatabase eagerly opens and verifies embedded databases. Networked
// dialects retain their existing lazy connection behavior.
func PrepareDatabase(config *Config) error {
	if config == nil {
		return errors.New("database config is nil")
	}
	if config.DialectType != DialectDoltLite {
		return nil
	}
	if err := config.Validate(); err != nil {
		return err
	}
	if config.doltLiteDatabase != nil {
		return nil
	}
	database, err := openDoltLiteDatabase(*config)
	if err != nil {
		return err
	}
	config.doltLiteDatabase = database
	return nil
}

// CloseDatabase closes the server-owned embedded database pool. It is a no-op
// for networked dialects, whose per-tool pools already close with transactions.
func CloseDatabase(config Config) error {
	if config.DialectType != DialectDoltLite || config.doltLiteDatabase == nil {
		return nil
	}
	return config.doltLiteDatabase.db.Close()
}

func openDoltLiteDatabase(config Config) (*doltLiteDatabase, error) {
	dialect := NewDialect(config.DialectType)
	dsn := dialect.FormatDSN(config)

	// Registers the cgo-backed driver; fails when the binary was built
	// without the "doltlite" build tag.
	if err := registerDoltLiteDriver(config); err != nil {
		return nil, err
	}

	db, err := sql.Open(doltLiteDriverName, dsn)
	if err != nil {
		return nil, err
	}
	// Do not retain idle handles. A fresh sqlite3 handle per MCP operation
	// avoids carrying connection-local branch/configuration or dynamically
	// registered DoltLite virtual-table state into a later operation. Handles
	// used concurrently remain independent and DoltLite coordinates the file.
	db.SetMaxIdleConns(0)

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

	return &doltLiteDatabase{db: db}, nil
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
