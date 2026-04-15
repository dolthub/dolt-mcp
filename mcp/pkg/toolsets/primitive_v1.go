package toolsets

import (
	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
)

type PrimitiveToolSetV1 struct{}

var _ ToolSet = &PrimitiveToolSetV1{}

type registerFunc func(pkg.Server)

type toolRegistration struct {
	name     string
	register registerFunc
}

var toolRegistrations = []toolRegistration{
	{tools.ListDatabasesToolName, tools.RegisterListDatabasesTool},
	{tools.ListDoltBranchesToolName, tools.RegisterListDoltBranchesTool},
	{tools.CreateDatabaseToolName, tools.RegisterCreateDatabaseTool},
	{tools.DropDatabaseToolName, tools.RegisterDropDatabaseTool},
	{tools.ShowTablesToolName, tools.RegisterShowTablesTool},
	{tools.ShowProcesslistToolName, tools.RegisterShowProcesslistTool},
	{tools.ShowCreateTableToolName, tools.RegisterShowCreateTableTool},
	{tools.DescribeTableToolName, tools.RegisterDescribeTableTool},
	{tools.CreateTableToolName, tools.RegisterCreateTableTool},
	{tools.DropTableToolName, tools.RegisterDropTableTool},
	{tools.AlterTableToolName, tools.RegisterAlterTableTool},
	{tools.QueryToolName, tools.RegisterQueryTool},
	{tools.ExecToolName, tools.RegisterExecTool},
	{tools.KillProcessToolName, tools.RegisterKillProcessTool},
	{tools.SelectActiveBranchToolName, tools.RegisterSelectActiveBranchTool},
	{tools.SelectVersionToolName, tools.RegisterSelectVersionTool},
	{tools.CreateDoltBranchFromHeadToolName, tools.RegisterCreateDoltBranchFromHeadTool},
	{tools.CreateDoltBranchToolName, tools.RegisterCreateDoltBranchTool},
	{tools.MoveDoltBranchToolName, tools.RegisterMoveDoltBranchTool},
	{tools.DeleteDoltBranchToolName, tools.RegisterDeleteDoltBranchTool},
	{tools.StageTableForDoltCommitToolName, tools.RegisterStageTableForDoltCommitTool},
	{tools.StageAllTablesForDoltCommitToolName, tools.RegisterStageAllTablesForDoltCommitTool},
	{tools.UnstageTableToolName, tools.RegisterUnstageTableTool},
	{tools.UnstageAllTablesToolName, tools.RegisterUnstageAllTablesTool},
	{tools.CreateDoltCommitToolName, tools.RegisterCreateDoltCommitTool},
	{tools.DoltResetSoftToolName, tools.RegisterDoltResetSoftTool},
	{tools.DoltResetHardToolName, tools.RegisterDoltResetHardTool},
	{tools.ListDoltCommitsToolName, tools.RegisterListDoltCommitsTool},
	{tools.ListDoltDiffChangesInWorkingSetToolName, tools.RegisterListDoltDiffChangesInWorkingSetTool},
	{tools.ListDoltDiffChangesByTableNameToolName, tools.RegisterListDoltDiffChangesByTableNameTool},
	{tools.ListDoltDiffChangesInDateRangeToolName, tools.RegisterListDoltDiffChangesInDateRangeTool},
	{tools.GetDoltMergeStatusToolName, tools.RegisterGetDoltMergeStatusTool},
	{tools.MergeDoltBranchToolName, tools.RegisterMergeDoltBranchTool},
	{tools.MergeDoltBranchNoFastForwardToolName, tools.RegisterMergeDoltBranchNoFastForwardTool},
	{tools.ListDoltRemotesToolName, tools.RegisterListDoltRemotesTool},
	{tools.AddDoltRemoteToolName, tools.RegisterAddDoltRemoteTool},
	{tools.RemoveDoltRemoteToolName, tools.RegisterRemoveDoltRemoteTool},
	{tools.CloneDatabaseToolName, tools.RegisterCloneDatabaseTool},
	{tools.DoltFetchBranchToolName, tools.RegisterDoltFetchBranchTool},
	{tools.DoltFetchAllBranchesToolName, tools.RegisterDoltFetchAllBranchesTool},
	{tools.DoltPushBranchToolName, tools.RegisterDoltPushBranchTool},
	{tools.DoltPullBranchToolName, tools.RegisterDoltPullBranchTool},
	{tools.RunDoltTestsToolName, tools.RegisterRunDoltTestsTool},
	{tools.AddDoltTestToolName, tools.RegisterAddDoltTestTool},
	{tools.RemoveDoltTestToolName, tools.RegisterRemoveDoltTestTool},
}

func (v *PrimitiveToolSetV1) RegisterTools(server pkg.Server) {
	dialect := server.Dialect()
	for _, t := range toolRegistrations {
		if dialect.SupportsTool(t.name) {
			t.register(server)
		}
	}
}