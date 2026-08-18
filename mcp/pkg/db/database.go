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
	executor         sqlExecutor
	db               *sql.DB
	conn             *sql.Conn
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

func newDoltLiteTransaction(ctx context.Context, config Config) (DatabaseTransaction, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
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

	if _, err = conn.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout = %d;", config.BusyTimeout.Milliseconds())); err != nil {
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

func isNoActiveTransactionError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no transaction is active")
}

const doltLiteTransactionCleanupTimeout = 5 * time.Second

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

type doltLiteDatabase struct {
	db *sql.DB
}

// PrepareDatabase opens an embedded database when configured.
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

// CloseDatabase closes an embedded database when configured.
func CloseDatabase(config Config) error {
	if config.DialectType != DialectDoltLite || config.doltLiteDatabase == nil {
		return nil
	}
	return config.doltLiteDatabase.db.Close()
}

func openDoltLiteDatabase(config Config) (*doltLiteDatabase, error) {
	dialect := NewDialect(config.DialectType)
	dsn := dialect.FormatDSN(config)

	if err := registerDoltLiteDriver(config); err != nil {
		return nil, err
	}

	db, err := sql.Open(doltLiteDriverName, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(0)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to open DoltLite database %s (the file may have been written by an incompatible DoltLite version): %w", dsn, err)
	}

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
