package db

import "errors"

// DialectType represents the type of SQL database dialect.
type DialectType string

const (
	DialectMySQL    DialectType = "mysql"
	DialectPostgres DialectType = "postgres"
	DialectDoltLite DialectType = "doltlite"
)

// Validation errors returned by Dialect validation methods.
var (
	ErrInvalidSQLReadQuery        = errors.New("invalid read query")
	ErrInvalidSQLWriteQuery       = errors.New("invalid write query")
	ErrInvalidCreateTableSQLQuery = errors.New("invalid create table statement")
	ErrInvalidAlterTableSQLQuery  = errors.New("invalid alter table statement")
)

// DoltProcedure represents a Dolt stored procedure name shared across dialects.
type DoltProcedure string

const (
	DoltCheckout DoltProcedure = "DOLT_CHECKOUT"
	DoltCommit   DoltProcedure = "DOLT_COMMIT"
	DoltBranch   DoltProcedure = "DOLT_BRANCH"
	DoltAdd      DoltProcedure = "DOLT_ADD"
	DoltReset    DoltProcedure = "DOLT_RESET"
	DoltMerge    DoltProcedure = "DOLT_MERGE"
	DoltRemote   DoltProcedure = "DOLT_REMOTE"
	DoltClone    DoltProcedure = "DOLT_CLONE"
	DoltFetch    DoltProcedure = "DOLT_FETCH"
	DoltPush     DoltProcedure = "DOLT_PUSH"
	DoltPull     DoltProcedure = "DOLT_PULL"
)

// Dialect encapsulates all SQL dialect differences between database engines.
type Dialect interface {
	// SupportsTool returns whether the given tool name is supported by this dialect.
	SupportsTool(toolName string) bool

	// Connection setup
	DriverName() string
	FormatDSN(c Config) string
	ConfigureTLS(c *Config) error

	// SQL generation
	QuoteIdentifier(name string) string
	CallProcedure(proc DoltProcedure, args ...string) string
	// UseDatabase returns a statement selecting the given database, or ""
	// when the dialect has no database selection (single-database engines).
	UseDatabase(database string) string

	// Schema inspection statements, which have no common syntax across engines.
	ShowTablesQuery() string
	ShowCreateTableQuery(table string) string
	DescribeTableQuery(table string) string

	// HashOfFunction returns a SQL expression resolving the given ref to a
	// commit hash.
	HashOfFunction(ref string) string
	// ListTableDiffChangesQuery returns a query listing dolt_diff_<table>
	// changes between two commits. fromExpr and toExpr are SQL expressions
	// (a quoted ref string or a HashOfFunction expression).
	ListTableDiffChangesQuery(table, fromExpr, toExpr string) string

	// SQL validation
	ValidateReadQuery(query string) error
	ValidateWriteQuery(query string) error
	ValidateCreateTableQuery(query string) error
	ValidateAlterTableQuery(query string) error
}

// NewDialect creates a Dialect for the given DialectType.
func NewDialect(dt DialectType) Dialect {
	switch dt {
	case DialectPostgres:
		return NewPostgresDialect()
	case DialectDoltLite:
		return NewDoltLiteDialect()
	default:
		return NewMySQLDialect()
	}
}