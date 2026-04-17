package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"

	gosql "github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
)

// MySQLDialect implements Dialect for MySQL-compatible Dolt servers.
type MySQLDialect struct {
	unsupportedTools map[string]bool
}

var _ Dialect = &MySQLDialect{}

func NewMySQLDialect() *MySQLDialect {
	return &MySQLDialect{}
}

func (d *MySQLDialect) SupportsTool(toolName string) bool {
	if d.unsupportedTools == nil {
		return true
	}
	return !d.unsupportedTools[toolName]
}

func (d *MySQLDialect) DriverName() string {
	return "mysql"
}

func (d *MySQLDialect) FormatDSN(c Config) string {
	if c.DSN != "" {
		return c.DSN
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", c.User, c.Password, c.Host, c.Port)
	if c.DatabaseName != "" {
		dsn += c.DatabaseName
	}

	options := []string{}
	if c.ParseTime {
		options = append(options, "parseTime=true")
	}
	if c.MultiStatements {
		options = append(options, "multiStatements=true")
	}
	if c.TLS != "" {
		options = append(options, fmt.Sprintf("tls=%s", c.TLS))
	}
	if len(options) > 0 {
		dsn += "?" + strings.Join(options, "&")
	}

	return dsn
}

func (d *MySQLDialect) ConfigureTLS(c *Config) error {
	if c.TLSCAFile == "" {
		return nil
	}
	rootCertPool := x509.NewCertPool()
	pem, err := os.ReadFile(c.TLSCAFile)
	if err != nil {
		return fmt.Errorf("failed to read CA file %s: %w", c.TLSCAFile, err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return fmt.Errorf("failed to append CA certificate from %s", c.TLSCAFile)
	}
	tlsConfig := &tls.Config{
		RootCAs: rootCertPool,
	}
	if err := mysql.RegisterTLSConfig("custom", tlsConfig); err != nil {
		return fmt.Errorf("failed to register TLS config: %w", err)
	}
	c.TLS = "custom"
	return nil
}

func (d *MySQLDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf("`%s`", name)
}

func (d *MySQLDialect) CallProcedure(proc DoltProcedure, args ...string) string {
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = fmt.Sprintf("'%s'", arg)
	}
	return fmt.Sprintf("CALL %s(%s);", string(proc), strings.Join(quotedArgs, ", "))
}

func (d *MySQLDialect) UseDatabase(database string) string {
	return fmt.Sprintf("USE `%s`;", database)
}

// SQL validation using the Vitess MySQL parser.

func (d *MySQLDialect) parseSQLQuery(query string) (sqlparser.Statement, error) {
	sqlCtx := gosql.NewEmptyContext()
	sqlMode := gosql.LoadSqlMode(sqlCtx)
	return sqlparser.ParseWithOptions(sqlCtx, query, sqlMode.ParserOptions())
}

func (d *MySQLDialect) isReadOnlyStatement(stmt sqlparser.Statement) bool {
	switch stmt.(type) {
	case sqlparser.SelectStatement:
		return true
	case *sqlparser.Show:
		return true
	case *sqlparser.Explain, *sqlparser.OtherRead:
		return true
	default:
		return false
	}
}

func (d *MySQLDialect) ValidateReadQuery(query string) error {
	stmt, err := d.parseSQLQuery(query)
	if err != nil {
		return err
	}
	if d.isReadOnlyStatement(stmt) {
		return nil
	}
	return ErrInvalidSQLReadQuery
}

func (d *MySQLDialect) ValidateWriteQuery(query string) error {
	stmt, err := d.parseSQLQuery(query)
	if err != nil {
		return err
	}
	if d.isReadOnlyStatement(stmt) {
		return ErrInvalidSQLWriteQuery
	}
	return nil
}

func (d *MySQLDialect) ValidateCreateTableQuery(query string) error {
	stmt, err := d.parseSQLQuery(query)
	if err != nil {
		return err
	}
	switch stmt.(type) {
	case *sqlparser.DDL:
		return nil
	}
	return ErrInvalidCreateTableSQLQuery
}

func (d *MySQLDialect) ValidateAlterTableQuery(query string) error {
	stmt, err := d.parseSQLQuery(query)
	if err != nil {
		return err
	}
	switch stmt.(type) {
	case *sqlparser.AlterTable:
		return nil
	}
	return ErrInvalidAlterTableSQLQuery
}