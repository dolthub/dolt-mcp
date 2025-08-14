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

func TestMutationToolsAnnotations(t *testing.T) {
	cases := []struct {
		name string
		ctor func() mcp.Tool
		ro, destr, idem, open bool
	}{
		{"stage_table_for_dolt_commit", NewStageTableForDoltCommitTool, false, false, true, false},
		{"stage_all_tables_for_dolt_commit", NewStageAllTablesForDoltCommitTool, false, false, true, false},
		{"unstage_table", NewUnstageTableTool, false, false, true, false},
		{"unstage_all_tables", NewUnstageAllTablesTool, false, false, true, false},
		{"dolt_reset_table_soft", NewDoltResetTableSoftTool, false, false, true, false},
		{"dolt_reset_all_tables_soft", NewDoltResetAllTablesSoftTool, false, false, true, false},
		{"dolt_reset_hard", NewDoltResetHardTool, false, true, true, false},
		{"create_dolt_branch", NewCreateDoltBranchTool, false, false, false, false},
		{"create_dolt_branch_from_head", NewCreateDoltBranchFromHeadTool, false, false, false, false},
		{"move_dolt_branch", NewMoveDoltBranchTool, false, false, false, false},
		{"delete_dolt_branch", NewDeleteDoltBranchTool, false, true, true, false},
		{"merge_dolt_branch", NewMergeDoltBranchTool, false, true, false, false},
		{"merge_dolt_branch_no_fast_forward", NewMergeDoltBranchNoFastForwardTool, false, true, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tool := tc.ctor()
			ann := tool.Annotations
			if b(ann.ReadOnlyHint) != tc.ro { t.Fatalf("%s: readOnly got %v", tc.name, b(ann.ReadOnlyHint)) }
			if b(ann.DestructiveHint) != tc.destr { t.Fatalf("%s: destructive got %v", tc.name, b(ann.DestructiveHint)) }
			if b(ann.IdempotentHint) != tc.idem { t.Fatalf("%s: idempotent got %v", tc.name, b(ann.IdempotentHint)) }
			if b(ann.OpenWorldHint) != tc.open { t.Fatalf("%s: openWorld got %v", tc.name, b(ann.OpenWorldHint)) }
			if tool.Name != tc.name { t.Fatalf("expected name %s, got %s", tc.name, tool.Name) }
		})
	}
}

func TestWriteToolAnnotations(t *testing.T) {
	cases := []struct {
		name string
		ctor func() mcp.Tool
		ro, destr, idem, open bool
	}{
		{"exec", NewExecTool, false, false, false, false},
		{"create_table", NewCreateTableTool, false, false, false, false},
		{"alter_table", NewAlterTableTool, false, true, false, false},
		{"drop_table", NewDropTableTool, false, true, true, false},
		{"create_database", NewCreateDatabaseTool, false, false, true, false},
		{"drop_database", NewDropDatabaseTool, false, true, true, false},
		{"create_dolt_commit", NewCreateDoltCommitTool, false, false, true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tool := tc.ctor()
			ann := tool.Annotations
			if b(ann.ReadOnlyHint) != tc.ro { t.Fatalf("%s: readOnly got %v", tc.name, b(ann.ReadOnlyHint)) }
			if b(ann.DestructiveHint) != tc.destr { t.Fatalf("%s: destructive got %v", tc.name, b(ann.DestructiveHint)) }
			if b(ann.IdempotentHint) != tc.idem { t.Fatalf("%s: idempotent got %v", tc.name, b(ann.IdempotentHint)) }
			if b(ann.OpenWorldHint) != tc.open { t.Fatalf("%s: openWorld got %v", tc.name, b(ann.OpenWorldHint)) }
			if tool.Name != tc.name { t.Fatalf("expected name %s, got %s", tc.name, tool.Name) }
		})
	}
}

func TestOpenWorldToolAnnotations(t *testing.T) {
	cases := []struct {
		name string
		ctor func() mcp.Tool
		ro, destr, idem, open bool
	}{
		{"add_dolt_remote", NewAddDoltRemoteTool, false, false, true, false},
		{"remove_dolt_remote", NewRemoveDoltRemoteTool, false, false, true, false},
		{"clone_database", NewCloneDatabaseTool, false, false, true, true},
		{"dolt_fetch_branch", NewDoltFetchBranchTool, false, false, true, true},
		{"dolt_fetch_all_branches", NewDoltFetchAllBranchesTool, false, false, true, true},
		{"dolt_push_branch", NewDoltPushBranchTool, false, false, false, true},
		{"dolt_pull_branch", NewDoltPullBranchTool, false, false, false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tool := tc.ctor()
			ann := tool.Annotations
			if b(ann.ReadOnlyHint) != tc.ro { t.Fatalf("%s: readOnly got %v", tc.name, b(ann.ReadOnlyHint)) }
			if b(ann.DestructiveHint) != tc.destr { t.Fatalf("%s: destructive got %v", tc.name, b(ann.DestructiveHint)) }
			if b(ann.IdempotentHint) != tc.idem { t.Fatalf("%s: idempotent got %v", tc.name, b(ann.IdempotentHint)) }
			if b(ann.OpenWorldHint) != tc.open { t.Fatalf("%s: openWorld got %v", tc.name, b(ann.OpenWorldHint)) }
			if tool.Name != tc.name { t.Fatalf("expected name %s, got %s", tc.name, tool.Name) }
		})
	}
}

// Optional bool pointer -> bool
func b(p *bool) bool { if p == nil { return false }; return *p }
