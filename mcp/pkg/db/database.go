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

type RowMap map[string]interface{}
type Columns []string

type Database interface {
	QueryContext(ctx context.Context, query string, resultFormat ResultFormat) (string, error)
	ExecContext(ctx context.Context, query string) error
}

type databaseImpl struct {
	db     *sql.DB
	config Config
}

var _ Database = &databaseImpl{}

func NewDatabase(config Config) (Database, error) {
	db, err := newDB(config)
	if err != nil {
		return nil, err
	}
	return &databaseImpl{
		db:     db,
		config: config,
	}, nil
}

func (d *databaseImpl) QueryContext(ctx context.Context, query string, resultFormat ResultFormat) (string, error) {
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

func (d *databaseImpl) rowMapToMarkdown(rowMaps []RowMap, headers []string) (string, error) {
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

func (d *databaseImpl) rowMapToCSV(rowMaps []RowMap, headers []string) (string, error) {
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

func (d *databaseImpl) doQueryContext(ctx context.Context, query string) ([]RowMap, Columns, error) {
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

func (d *databaseImpl) doExecContext(ctx context.Context, query string) error {
	_, err := d.db.ExecContext(ctx, query)
	return err
}

func (d *databaseImpl) ExecContext(ctx context.Context, query string) error {
	return d.doExecContext(ctx, query)
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
