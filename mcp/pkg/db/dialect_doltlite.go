package db

import (
	"fmt"
	"regexp"
	"strings"
)

// doltLiteDriverName is the database/sql driver name registered when the
// server is built with the "doltlite" build tag. See driver_doltlite.go.
const doltLiteDriverName = "doltlite"

// DoltLiteDialect implements Dialect for DoltLite, an embedded SQLite fork
// with Dolt-style version control. DoltLite speaks the SQLite dialect: Dolt
// procedures are scalar functions invoked with SELECT, and version-control
// state (branch, log, diff) is exposed through virtual tables.
type DoltLiteDialect struct {
	unsupportedTools map[string]bool
}

var _ Dialect = &DoltLiteDialect{}

func NewDoltLiteDialect() *DoltLiteDialect {
	return &DoltLiteDialect{
		unsupportedTools: map[string]bool{
			// DoltLite is a single-file, single-database engine.
			"list_databases":  true,
			"create_database": true,
			"drop_database":   true,
			// dolt_clone clones into the currently open (empty) database
			// file; it cannot create a new named database.
			"clone_database": true,
			// There is no server process.
			"show_processlist": true,
			"kill_process":     true,
			// The dolt_merge_status system table does not exist in DoltLite.
			"get_dolt_merge_status": true,
			// The dolt_tests table and dolt_test_run() do not exist in DoltLite.
			"run_dolt_tests":   true,
			"add_dolt_test":    true,
			"remove_dolt_test": true,
		},
	}
}

func (d *DoltLiteDialect) SupportsTool(toolName string) bool {
	return !d.unsupportedTools[toolName]
}

func (d *DoltLiteDialect) DriverName() string {
	return doltLiteDriverName
}

func (d *DoltLiteDialect) FormatDSN(c Config) string {
	if c.DSN != "" {
		return c.DSN
	}
	return c.Path
}

func (d *DoltLiteDialect) ConfigureTLS(_ *Config) error {
	// DoltLite is embedded; there is no connection to secure.
	return nil
}

func (d *DoltLiteDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

// escapeStringLiteral escapes a value for inclusion in a single-quoted
// SQLite string literal.
func escapeStringLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func (d *DoltLiteDialect) CallProcedure(proc DoltProcedure, args ...string) string {
	fnName := strings.ToLower(string(proc))
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = fmt.Sprintf("'%s'", escapeStringLiteral(arg))
	}

	// DoltLite refuses to switch branches when the working set is dirty,
	// including a checkout of the branch the session is already on in some
	// merge states. Since tools check out their working branch on every
	// call, skip the checkout when the session is already on the target
	// branch. CASE evaluates lazily in SQLite, so dolt_checkout only runs
	// when the branch actually differs.
	if proc == DoltCheckout && len(args) == 1 {
		return fmt.Sprintf(
			"SELECT CASE WHEN active_branch() = %s THEN 0 ELSE dolt_checkout(%s) END;",
			quotedArgs[0], quotedArgs[0],
		)
	}

	// Full Dolt accepts DOLT_RESET('.') as "unstage everything". DoltLite
	// treats "." as a literal table name; its equivalent is dolt_reset()
	// with no arguments.
	if proc == DoltReset && len(args) == 1 && args[0] == "." {
		return "SELECT dolt_reset();"
	}

	// DoltLite rejects force-updating the currently checked-out branch. For
	// the create-from-HEAD tool that operation is already satisfied, so make
	// it an idempotent no-op while preserving force behavior for other names.
	if proc == DoltBranch && len(args) == 2 && args[0] == "-f" {
		return fmt.Sprintf(
			"SELECT CASE WHEN active_branch() = %s THEN 0 ELSE dolt_branch(%s, %s) END;",
			quotedArgs[1], quotedArgs[0], quotedArgs[1],
		)
	}

	// Full Dolt takes --force before the remote and branch. DoltLite's SQL
	// function takes it last: dolt_push(remote, branch, '--force').
	if proc == DoltPush && len(args) == 3 && args[0] == "--force" {
		return fmt.Sprintf("SELECT dolt_push(%s, %s, %s);", quotedArgs[1], quotedArgs[2], quotedArgs[0])
	}

	return fmt.Sprintf("SELECT %s(%s);", fnName, strings.Join(quotedArgs, ", "))
}

func (d *DoltLiteDialect) UseDatabase(_ string) string {
	// DoltLite has exactly one database per file; there is nothing to select.
	return ""
}

func (d *DoltLiteDialect) ShowTablesQuery() string {
	return "SELECT name FROM sqlite_schema WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name;"
}

func (d *DoltLiteDialect) ShowCreateTableQuery(table string) string {
	return fmt.Sprintf(
		"SELECT name AS \"Table\", sql AS \"Create Table\" FROM sqlite_schema WHERE type = 'table' AND name = '%s';",
		escapeStringLiteral(table),
	)
}

func (d *DoltLiteDialect) DescribeTableQuery(table string) string {
	return fmt.Sprintf(
		"SELECT name AS \"Field\", type AS \"Type\", CASE WHEN \"notnull\" = 0 THEN 'YES' ELSE 'NO' END AS \"Null\", CASE WHEN pk > 0 THEN 'PRI' ELSE '' END AS \"Key\", dflt_value AS \"Default\" FROM pragma_table_info('%s');",
		escapeStringLiteral(table),
	)
}

func (d *DoltLiteDialect) HashOfFunction(ref string) string {
	return fmt.Sprintf("dolt_hashof('%s')", escapeStringLiteral(ref))
}

func (d *DoltLiteDialect) ListTableDiffChangesQuery(table, fromExpr, toExpr string) string {
	// DoltLite's per-table diff vtable takes the from/to refs as
	// table-valued-function arguments; its from_commit/to_commit result
	// columns hold resolved hashes, so Dolt's WHERE form matches nothing.
	return fmt.Sprintf("SELECT * FROM dolt_diff_%s(%s, %s);", table, fromExpr, toExpr)
}

// SQL validation.
//
// There is no production-grade SQLite parser available in pure Go, so
// validation classifies statements by their leading keyword; the engine
// itself rejects anything syntactically invalid at prepare time.

// liteLeadingKeyword returns the first SQL keyword of the query in upper
// case, skipping whitespace and comments.
func liteLeadingKeyword(query string) string {
	s := query
	for {
		s = strings.TrimLeft(s, " \t\r\n;")
		switch {
		case strings.HasPrefix(s, "--"):
			idx := strings.Index(s, "\n")
			if idx < 0 {
				return ""
			}
			s = s[idx+1:]
		case strings.HasPrefix(s, "/*"):
			idx := strings.Index(s, "*/")
			if idx < 0 {
				return ""
			}
			s = s[idx+2:]
		default:
			end := strings.IndexFunc(s, func(r rune) bool {
				return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_')
			})
			if end < 0 {
				end = len(s)
			}
			return strings.ToUpper(s[:end])
		}
	}
}

// liteSecondKeyword returns the keyword following the leading keyword, used
// to distinguish e.g. CREATE TABLE from CREATE INDEX.
func liteSecondKeyword(query string) string {
	first := liteLeadingKeyword(query)
	if first == "" {
		return ""
	}
	s := strings.TrimLeft(query, " \t\r\n;")
	idx := strings.Index(strings.ToUpper(s), first)
	if idx < 0 {
		return ""
	}
	return liteLeadingKeyword(s[idx+len(first):])
}

var liteReadOnlyKeywords = map[string]bool{
	"SELECT":  true,
	"WITH":    true,
	"VALUES":  true,
	"EXPLAIN": true,
}

// liteContainsMultipleStatements reports whether query contains a second
// non-empty SQL statement. The sqlite3 driver accepts statement batches, so
// leading-keyword validation alone is not enough for the read-only query
// tool (for example, "SELECT 1; DELETE FROM t" must be rejected). Semicolons
// inside quoted strings, identifiers, and comments do not terminate a
// statement.
func liteContainsMultipleStatements(query string) bool {
	const (
		liteScanNormal = iota
		liteScanSingleQuote
		liteScanDoubleQuote
		liteScanBacktick
		liteScanBracket
		liteScanLineComment
		liteScanBlockComment
	)

	state := liteScanNormal
	seenStatement := false
	endedStatement := false
	for i := 0; i < len(query); i++ {
		c := query[i]
		next := byte(0)
		if i+1 < len(query) {
			next = query[i+1]
		}

		switch state {
		case liteScanSingleQuote:
			if c == '\'' {
				if next == '\'' {
					i++
				} else {
					state = liteScanNormal
				}
			}
		case liteScanDoubleQuote:
			if c == '"' {
				if next == '"' {
					i++
				} else {
					state = liteScanNormal
				}
			}
		case liteScanBacktick:
			if c == '`' {
				if next == '`' {
					i++
				} else {
					state = liteScanNormal
				}
			}
		case liteScanBracket:
			if c == ']' {
				state = liteScanNormal
			}
		case liteScanLineComment:
			if c == '\n' || c == '\r' {
				state = liteScanNormal
			}
		case liteScanBlockComment:
			if c == '*' && next == '/' {
				state = liteScanNormal
				i++
			}
		default:
			switch {
			case c == '-' && next == '-':
				state = liteScanLineComment
				i++
			case c == '/' && next == '*':
				state = liteScanBlockComment
				i++
			case c == '\'':
				if endedStatement {
					return true
				}
				seenStatement = true
				state = liteScanSingleQuote
			case c == '"':
				if endedStatement {
					return true
				}
				seenStatement = true
				state = liteScanDoubleQuote
			case c == '`':
				if endedStatement {
					return true
				}
				seenStatement = true
				state = liteScanBacktick
			case c == '[':
				if endedStatement {
					return true
				}
				seenStatement = true
				state = liteScanBracket
			case c == ';':
				if seenStatement {
					endedStatement = true
				}
			case c == ' ' || c == '\t' || c == '\r' || c == '\n':
				// Trivia between or after statements.
			default:
				if endedStatement {
					return true
				}
				seenStatement = true
			}
		}
	}
	return false
}

// In DoltLite, version-control mutations are scalar functions invoked with
// SELECT, so a syntactically read-only statement can still mutate the
// database. Reject read queries that call any mutating dolt function.
// Read-only functions and virtual tables (active_branch, dolt_merge_base,
// dolt_hashof*, dolt_branches, dolt_log, ...) do not match this pattern.
var liteMutatingFunctionPattern = regexp.MustCompile(
	`(?i)\bdolt_(commit|add|reset|merge|branch|checkout|connect_branch|cherry_pick|revert|rebase|tag|remote|push|fetch|pull|clone|gc|config|creds|creds_new|conflicts_resolve)\s*\(`,
)

func (d *DoltLiteDialect) ValidateReadQuery(query string) error {
	if !liteReadOnlyKeywords[liteLeadingKeyword(query)] {
		return ErrInvalidSQLReadQuery
	}
	if liteContainsMultipleStatements(query) {
		return ErrInvalidSQLReadQuery
	}
	if liteMutatingFunctionPattern.MatchString(query) {
		return ErrInvalidSQLReadQuery
	}
	return nil
}

func (d *DoltLiteDialect) ValidateWriteQuery(query string) error {
	keyword := liteLeadingKeyword(query)
	if keyword == "" {
		return ErrInvalidSQLWriteQuery
	}
	if liteReadOnlyKeywords[keyword] {
		if liteMutatingFunctionPattern.MatchString(query) {
			return nil
		}
		return ErrInvalidSQLWriteQuery
	}
	return nil
}

func (d *DoltLiteDialect) ValidateCreateTableQuery(query string) error {
	if liteLeadingKeyword(query) == "CREATE" && liteSecondKeyword(query) == "TABLE" {
		return nil
	}
	return ErrInvalidCreateTableSQLQuery
}

func (d *DoltLiteDialect) ValidateAlterTableQuery(query string) error {
	if liteLeadingKeyword(query) == "ALTER" && liteSecondKeyword(query) == "TABLE" {
		return nil
	}
	return ErrInvalidAlterTableSQLQuery
}
