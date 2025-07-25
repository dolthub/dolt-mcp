package db

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
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
}

var _ DatabaseTransaction = &databaseTransactionImpl{}

func NewDatabaseTransaction(ctx context.Context, config Config) (DatabaseTransaction, error) {
	db, err := newDB(config)
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, "BEGIN;")
	if err != nil {
		return nil, err
	}
	return &databaseTransactionImpl{
		db:     db,
		config: config,
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

func (d *databaseTransactionImpl) Rollback(ctx context.Context) (err error) {
	if d.db == nil {
		err = ErrTransactionHasBeenCommittedOrRolledBack
		return
	}

	defer func(){
		rerr := d.db.Close()
		if err == nil {
			err = rerr
		}
		d.db = nil
	}()

	err = d.doExecContext(ctx, "ROLLBACK;")
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

	defer func(){
		rerr := d.db.Close()
		if err == nil {
			err = rerr
		}
		d.db = nil
	}()

	err = d.doExecContext(ctx, "COMMIT;")
	if err != nil {
		return
	}

	return
}

func newDB(config Config) (*sql.DB, error) {
	dsn := config.GetDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

