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
	t.Run("TestShowTablesTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testShowTablesToolInvalidArguments)
		RunTest(t, "TestSuccess", testShowTablesToolSuccess)
	})
	t.Run("TestDescribeTableTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDescribeTableToolInvalidArguments)
		RunTest(t, "TestSuccess", testDescribeTableToolSuccess)
	})
	t.Run("TestShowCreateTableTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testShowCreateTableToolInvalidArguments)
		RunTest(t, "TestSuccess", testShowCreateTableToolSuccess)
	})
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
	t.Run("TestSelectActiveBranchTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testSelectActiveBranchToolInvalidArguments)
		RunTest(t, "TestSuccess", testSelectActiveBranchToolSuccess)
	})
	t.Run("TestQueryTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testQueryToolInvalidArguments)
		RunTest(t, "TestSuccess", testQueryToolSuccess)
	})
	t.Run("TestExecTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testExecToolInvalidArguments)
		RunTest(t, "TestSuccess", testExecToolSuccess)
	})
	t.Run("TestCreateDoltBranchFromHeadTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testCreateDoltBranchFromHeadToolInvalidArguments)
		RunTestWithTeardownSQL(t, "TestSuccess", testCreateDoltBranchTeardownSQL, testCreateDoltBranchFromHeadToolSuccess)
	})
	t.Run("TestCreateDoltBranchTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testCreateDoltBranchToolInvalidArguments)
		RunTestWithTeardownSQL(t, "TestSuccess", testCreateDoltBranchTeardownSQL, testCreateDoltBranchToolSuccess)
	})
}

