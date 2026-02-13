package tools

import "github.com/dolthub/vitess/go/vt/sqlparser"

// IsReadOnlyStatement returns true if the parsed SQL statement is considered read-only.
//
// This is intentionally conservative and is used to gate the `query` tool (read-only)
// vs the `exec` tool (non-read). It should not attempt to validate semantics, only
// classify by AST node type.
func IsReadOnlyStatement(stmt sqlparser.Statement) bool {
	switch stmt.(type) {
	// SELECT-like statements (including WITH/CTEs and set-ops like UNION/INTERSECT/EXCEPT)
	case sqlparser.SelectStatement:
		return true

	// SHOW ... statements (e.g. SHOW PROCESSLIST)
	case *sqlparser.Show:
		return true

	// EXPLAIN ... and other parser-indicated reads.
	case *sqlparser.Explain, *sqlparser.OtherRead:
		return true
	default:
		return false
	}
}
