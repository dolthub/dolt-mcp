package db

import (
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	pganalyze "github.com/pganalyze/pg_query_go/v6"
	pg_query "github.com/wasilibs/go-pgquery"
)

// PostgresDialect implements Dialect for PostgreSQL-compatible DoltgreSQL servers.
type PostgresDialect struct {
	unsupportedTools map[string]bool
}

var _ Dialect = &PostgresDialect{}

func NewPostgresDialect() *PostgresDialect {
	return &PostgresDialect{
		unsupportedTools: map[string]bool{
			"add_dolt_test":    true,
			"remove_dolt_test": true,
			"run_dolt_tests":   true,
			"kill_process":     true,
			"show_processlist": true,
		},
	}
}

func (d *PostgresDialect) SupportsTool(toolName string) bool {
	return !d.unsupportedTools[toolName]
}

func (d *PostgresDialect) DriverName() string {
	return "pgx"
}

func (d *PostgresDialect) FormatDSN(c Config) string {
	if c.DSN != "" {
		return c.DSN
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d", c.User, c.Password, c.Host, c.Port)
	if c.DatabaseName != "" {
		dsn += "/" + c.DatabaseName
	}

	options := []string{}
	sslMode := d.mapTLSToSSLMode(c.TLS, c.TLSCAFile)
	if sslMode != "" {
		options = append(options, fmt.Sprintf("sslmode=%s", sslMode))
	}
	if c.TLSCAFile != "" {
		options = append(options, fmt.Sprintf("sslrootcert=%s", c.TLSCAFile))
	}
	if c.MultiStatements {
		// PostgreSQL's extended query protocol (pgx's default) handles only
		// one statement per call. Simple protocol supports ';'-separated
		// multi-statement queries, mirroring MySQL's multiStatements=true.
		options = append(options, "default_query_exec_mode=simple_protocol")
	}
	if len(options) > 0 {
		dsn += "?" + strings.Join(options, "&")
	}

	return dsn
}

func (d *PostgresDialect) mapTLSToSSLMode(tls, tlsCAFile string) string {
	switch tls {
	case "true":
		if tlsCAFile != "" {
			return "verify-ca"
		}
		return "require"
	case "false", "":
		return "disable"
	case "skip-verify":
		return "require"
	case "preferred":
		return "prefer"
	default:
		return "disable"
	}
}

func (d *PostgresDialect) ConfigureTLS(_ *Config) error {
	// PostgreSQL handles TLS via sslmode DSN parameter, configured in FormatDSN.
	return nil
}

func (d *PostgresDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func (d *PostgresDialect) CallProcedure(proc DoltProcedure, args ...string) string {
	pgName := strings.ToLower(string(proc))
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = fmt.Sprintf("'%s'", arg)
	}
	return fmt.Sprintf("SELECT %s(%s);", pgName, strings.Join(quotedArgs, ", "))
}

func (d *PostgresDialect) UseDatabase(database string) string {
	return fmt.Sprintf(`USE "%s";`, database)
}

// SQL validation using the PostgreSQL parser.

func (d *PostgresDialect) parseSQLQuery(query string) (*pganalyze.ParseResult, error) {
	return pg_query.Parse(query)
}

func (d *PostgresDialect) isReadOnlyStatement(result *pganalyze.ParseResult) bool {
	if len(result.Stmts) == 0 {
		return false
	}
	node := result.Stmts[0].Stmt
	return node.GetSelectStmt() != nil ||
		node.GetVariableShowStmt() != nil ||
		node.GetExplainStmt() != nil
}

func (d *PostgresDialect) ValidateReadQuery(query string) error {
	result, err := d.parseSQLQuery(query)
	if err != nil {
		return err
	}
	if d.isReadOnlyStatement(result) {
		return nil
	}
	return ErrInvalidSQLReadQuery
}

func (d *PostgresDialect) ValidateWriteQuery(query string) error {
	result, err := d.parseSQLQuery(query)
	if err != nil {
		return err
	}
	if d.isReadOnlyStatement(result) {
		return ErrInvalidSQLWriteQuery
	}
	return nil
}

func (d *PostgresDialect) ValidateCreateTableQuery(query string) error {
	result, err := d.parseSQLQuery(query)
	if err != nil {
		return err
	}
	if len(result.Stmts) > 0 && result.Stmts[0].Stmt.GetCreateStmt() != nil {
		return nil
	}
	return ErrInvalidCreateTableSQLQuery
}

func (d *PostgresDialect) ValidateAlterTableQuery(query string) error {
	result, err := d.parseSQLQuery(query)
	if err != nil {
		return err
	}
	if len(result.Stmts) > 0 && result.Stmts[0].Stmt.GetAlterTableStmt() != nil {
		return nil
	}
	return ErrInvalidAlterTableSQLQuery
}