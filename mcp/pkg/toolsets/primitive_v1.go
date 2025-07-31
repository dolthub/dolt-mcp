package toolsets

import (
	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
)

type PrimitiveToolSetV1 struct{}

var _ ToolSet = &PrimitiveToolSetV1{}

func (v *PrimitiveToolSetV1) RegisterTools(server pkg.Server) {
	tools.RegisterListDatabasesTool(server)
	tools.RegisterListDoltBranchesTool(server)
	tools.RegisterCreateDatabaseTool(server)
	tools.RegisterDropDatabaseTool(server)
	tools.RegisterShowTablesTool(server)
	tools.RegisterShowCreateTableTool(server)
	tools.RegisterDescribeTableTool(server)
	tools.RegisterCreateTableTool(server)
	tools.RegisterDropTableTool(server)
	tools.RegisterAlterTableTool(server)
	tools.RegisterQueryTool(server)
	tools.RegisterExecTool(server)
	tools.RegisterSelectActiveBranchTool(server)
	tools.RegisterSelectVersionTool(server)
	tools.RegisterCreateDoltBranchFromHeadTool(server)
	tools.RegisterCreateDoltBranchTool(server)
	tools.RegisterMoveDoltBranchTool(server)
	tools.RegisterDeleteDoltBranchTool(server)
	tools.RegisterStageTableForDoltCommitTool(server)
	tools.RegisterStageAllTablesForDoltCommitTool(server)
	tools.RegisterUnstageTableTool(server)
	tools.RegisterUnstageAllTablesTool(server)
	tools.RegisterCreateDoltCommitTool(server)
	tools.RegisterDoltResetTableSoftTool(server)
	tools.RegisterDoltResetAllTablesSoftTool(server)
	tools.RegisterDoltResetHardTool(server)
	// TODO: dolt_log
	// TODO: dolt_diff
	// TODO: dolt_merge
	// TODO: dolt_remote
	// TODO: dolt_clone
	// TODO: dolt_push
	// TODO: dolt_pull
}

