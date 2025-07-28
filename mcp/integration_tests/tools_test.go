package integration_tests

import (
	"testing"
)

func TestTools(t *testing.T) {
	RunTest(t, "TestListDatabasesTool", testListDatabasesTool)
	t.Run("TestUseDatabaseTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testUseDatabaseToolInvalidArguments)
		RunTest(t, "TestSuccess", testUseDatabaseToolSuccess)
	})
	t.Run("TestCreateDatabaseTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testCreateDatabaseToolInvalidArguments)
		RunTestWithTeardownSQL(t, "TestSuccess", testCreateDatabaseTeardownSQL, testCreateDatabaseToolSuccess)
	})
	t.Run("TestDropDatabaseTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDropDatabaseToolInvalidArguments)
		RunTestWithSetupSQL(t, "TestSuccess", testDropDatabaseSetupSQL, testDropDatabaseToolSuccess)
	})
	RunTest(t, "TestShowTablesTool", testShowTablesTool)
	t.Run("TestDescribeTableTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDescribeTableToolInvalidArguments)
		RunTest(t, "TestSuccess", testDescribeTableToolSuccess)
	})
	RunTest(t, "TestShowCreateTableTool", testShowCreateTableTool)
	t.Run("TestCreateTableTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testCreateTableToolInvalidArguments)
		RunTest(t, "TestSuccess", testCreateTableToolSuccess)
	})
	t.Run("TestAlterTableTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testAlterTableToolInvalidArguments)
		RunTest(t, "TestSuccess", testAlterTableToolSuccess)
	})
	t.Run("TestDropTableTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDropTableToolInvalidArguments)
		RunTestWithSetupSQL(t, "TestSuccess", testDropTableSetupSQL, testDropTableToolSuccess)
	})
	RunTest(t, "TestSelectVersionTool", testSelectVersionTool)
	RunTest(t, "TestSelectActiveBranchTool", testSelectActiveBranchTool)
}

