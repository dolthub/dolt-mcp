package tools

const (
	DatabaseCallToolArgumentName       = "database"
	WorkingBranchCallToolArgumentName  = "working_branch"
	TableCallToolArgumentName          = "table"
	QueryCallToolArgumentName          = "query"
	IfNotExistsCallToolArgumentName    = "if_not_exists"
	IfExistsCallToolArgumentName       = "if_exists"
	OriginalBranchCallToolArgumentName = "original_branch"
	NewBranchCallToolArgumentName      = "new_branch"
	CreateCallToolArgumentName         = "create"
	MoveCallToolArgumentName           = "move"
	DeleteCallToolArgumentName         = "delete"
	ForceCallToolArgumentName          = "force"
)

var WorkingBranchCallToolArgumentDescription = "The name of the working branch to checkout before prior to making the tool call."

