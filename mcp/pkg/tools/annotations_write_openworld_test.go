package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestWriteToolAnnotations(t *testing.T) {
	cases := []struct {
		name string
		ctor func() mcp.Tool
		ro, destr, idem, open bool
	}{
		{"exec", NewExecTool, false, true, false, false},
		{"create_table", NewCreateTableTool, false, true, false, false},
		{"alter_table", NewAlterTableTool, false, true, false, false},
		{"drop_table", NewDropTableTool, false, true, true, false},
		{"create_database", NewCreateDatabaseTool, false, true, true, false},
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
		{"add_dolt_remote", NewAddDoltRemoteTool, false, false, true, true},
		{"remove_dolt_remote", NewRemoveDoltRemoteTool, false, false, true, true},
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
