package toolsets

import (
	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
)

type PrimitiveToolSetV1 struct{}

var _ ToolSet = &PrimitiveToolSetV1{}

func (v *PrimitiveToolSetV1) RegisterTools(server pkg.Server) {
	tools.RegisterListDatabasesTool(server)
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
	// TODO: dolt_add
	// TODO: dolt_commit
	// TODO: dolt_remote
	// TODO: dolt_clone
	// TODO: dolt_push
	// TODO: dolt_pull
	// TODO: dolt_log
	// TODO: dolt_diff
	// TODO: dolt_merge
}

