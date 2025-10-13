package toolsets

import (
	"github.com/dolthub/dolt-mcp/mcp/pkg"
)

func WithToolSet(ts ToolSet) pkg.Option {
	return func(s pkg.Server) {
		ts.RegisterTools(s)
	}
}

type ToolSet interface {
	RegisterTools(server pkg.Server)
}
