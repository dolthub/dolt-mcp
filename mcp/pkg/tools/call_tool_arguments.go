package tools

const (
	DatabaseCallToolArgumentName        = "database"
	WorkingBranchCallToolArgumentName   = "working_branch"
	WorkingDatabaseCallToolArgumentName = "working_database"
	TableCallToolArgumentName           = "table"
	MessageCallToolArgumentName         = "message"
	QueryCallToolArgumentName           = "query"
	IfNotExistsCallToolArgumentName     = "if_not_exists"
	IfExistsCallToolArgumentName        = "if_exists"
	BranchCallToolArgumentName          = "branch"
	OriginalBranchCallToolArgumentName  = "original_branch"
	NewBranchCallToolArgumentName       = "new_branch"
	OldNameCallToolArgumentName         = "old_name"
	NewNameCallToolArgumentName         = "new_name"
	CreateCallToolArgumentName          = "create"
	MoveCallToolArgumentName            = "move"
	DeleteCallToolArgumentName          = "delete"
	ForceCallToolArgumentName           = "force"
)

var DoltUseWorkingDatabaseSQLQueryFormatString = "USE %s;"
var WorkingDatabaseCallToolArgumentDescription = "The name of the database to use prior to making the tool call."

var DoltCheckoutWorkingBranchSQLQueryFormatString = "CALL DOLT_CHECKOUT('%s');"
var WorkingBranchCallToolArgumentDescription = "The name of the working branch to checkout prior to making the tool call."
