package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

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
