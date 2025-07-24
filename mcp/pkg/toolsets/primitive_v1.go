package toolsets

import (
	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/tools"
)

type PrimitiveToolSetV1 struct{}

var _ ToolSet = &PrimitiveToolSetV1{}

func (v *PrimitiveToolSetV1) RegisterTools(server pkg.Server) {
	tools.RegisterListDatabasesTool(server)
	tools.RegisterUseDatabaseTool(server)
	tools.RegisterCreateDatabaseTool(server)
	tools.RegisterDropDatabaseTool(server)
	// TODO: show tables
	// TODO: show create table
	// TODO: describe table
	// TODO: create table
	// TODO: drop table
	// TODO: alter table
	// TODO: create index
	// TODO: drop index
	// TODO: select active branch
	// TODO: select version
	// TODO: query
	// TODO: exec
	// TODO: dolt_branch
	// TODO: dolt_checkout
	// TODO: dolt_add
	// TODO: dolt_commit
	// TODO: dolt_remote
	// TODO: dolt_push
	// TODO: dolt_pull
}

