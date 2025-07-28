package tools

const (
	DatabaseCallToolArgumentName        = "database"
	WorkingBranchCallToolArgumentName   = "working_branch"
	WorkingDatabaseCallToolArgumentName = "working_database"
	TableCallToolArgumentName           = "table"
	QueryCallToolArgumentName           = "query"
	IfNotExistsCallToolArgumentName     = "if_not_exists"
	IfExistsCallToolArgumentName        = "if_exists"
	OriginalBranchCallToolArgumentName  = "original_branch"
	NewBranchCallToolArgumentName       = "new_branch"
	CreateCallToolArgumentName          = "create"
	MoveCallToolArgumentName            = "move"
	DeleteCallToolArgumentName          = "delete"
	ForceCallToolArgumentName           = "force"
)

var DoltUseWorkingDatabaseSQLQueryFormatString = "USE %s;"
var WorkingDatabaseCallToolArgumentDescription = "The name of the database to use prior to making the tool call."

var DoltCheckoutWorkingBranchSQLQueryFormatString = "CALL DOLT_CHECKOUT('%s');"
var WorkingBranchCallToolArgumentDescription = "The name of the working branch to checkout prior to making the tool call."
