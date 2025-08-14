package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestReadOnlyToolAnnotations(t *testing.T) {
	cases := []struct {
		name string
		ctor func() mcp.Tool
	}{
		{"query", NewQueryTool},
		{"show_tables", NewShowTablesTool},
		{"show_create_table", NewShowCreateTableTool},
		{"describe_table", NewDescribeTableTool},
		{"list_databases", NewListDatabasesTool},
		{"list_dolt_branches", NewListDoltBranchesTool},
		{"list_dolt_commits", NewListDoltCommitsTool},
		{"list_dolt_remotes", NewListDoltRemotesTool},
		{"list_dolt_diff_changes_in_working_set", NewListDoltDiffChangesInWorkingSetTool},
		{"list_dolt_diff_changes_in_date_range", NewListDoltDiffChangesInDateRangeTool},
		{"list_dolt_diff_changes_by_table_name", NewListDoltDiffChangesByTableNameTool},
		{"get_dolt_merge_status", NewGetDoltMergeStatusTool},
		{"select_active_branch", NewSelectActiveBranchTool},
		{"select_version", NewSelectVersionTool},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tool := tc.ctor()
			ann := tool.Annotations
			if b(ann.ReadOnlyHint) != true {
				t.Fatalf("%s: expected readOnly=true", tc.name)
			}
			if b(ann.DestructiveHint) != false {
				t.Fatalf("%s: expected destructive=false", tc.name)
			}
			if b(ann.IdempotentHint) != true {
				t.Fatalf("%s: expected idempotent=true", tc.name)
			}
			if b(ann.OpenWorldHint) != false {
				t.Fatalf("%s: expected openWorld=false", tc.name)
			}
			if tool.Name != tc.name {
				t.Fatalf("expected tool name %s, got %s", tc.name, tool.Name)
			}
		})
	}
}

// Optional bool pointer -> bool
func b(p *bool) bool { if p == nil { return false }; return *p }
