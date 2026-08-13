package db

import (
	"fmt"
	"regexp"
	"strings"
)

const doltLiteDriverName = "doltlite"

// DoltLiteDialect implements Dialect for embedded DoltLite databases.
type DoltLiteDialect struct {
	unsupportedTools map[string]bool
}

var _ Dialect = &DoltLiteDialect{}

func NewDoltLiteDialect() *DoltLiteDialect {
	return &DoltLiteDialect{
		unsupportedTools: map[string]bool{
			"list_databases":   true,
			"create_database":  true,
			"drop_database":    true,
			"clone_database":   true,
			"show_processlist": true,
			"kill_process":     true,
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
	return nil
}

func (d *DoltLiteDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

func escapeStringLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func (d *DoltLiteDialect) CallProcedure(proc DoltProcedure, args ...string) string {
	fnName := strings.ToLower(string(proc))
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = fmt.Sprintf("'%s'", escapeStringLiteral(arg))
	}

	if proc == DoltCheckout && len(args) == 1 {
		return fmt.Sprintf(
			"SELECT CASE WHEN active_branch() = %s THEN 0 ELSE dolt_checkout(%s) END;",
			quotedArgs[0], quotedArgs[0],
		)
	}

	if proc == DoltReset && len(args) == 1 && args[0] == "." {
		return "SELECT dolt_reset();"
	}

	if proc == DoltBranch && len(args) == 2 && args[0] == "-f" {
		return fmt.Sprintf(
			"SELECT CASE WHEN active_branch() = %s THEN 0 ELSE dolt_branch(%s, %s) END;",
			quotedArgs[1], quotedArgs[0], quotedArgs[1],
		)
	}

	if proc == DoltPush && len(args) == 3 && args[0] == "--force" {
		return fmt.Sprintf("SELECT dolt_push(%s, %s, %s);", quotedArgs[1], quotedArgs[2], quotedArgs[0])
	}

	return fmt.Sprintf("SELECT %s(%s);", fnName, strings.Join(quotedArgs, ", "))
}

func (d *DoltLiteDialect) UseDatabase(_ string) string {
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
	diffTable := d.QuoteIdentifier("dolt_diff_" + table)
	return fmt.Sprintf("SELECT * FROM %s(%s, %s);", diffTable, fromExpr, toExpr)
}

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
