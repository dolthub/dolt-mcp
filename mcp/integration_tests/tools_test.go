package integration_tests

import (
	"testing"
)

func TestTools(t *testing.T) {
	// RunTest(t, "TestListDatabasesTool", testListDatabasesTool)
	// t.Run("TestUseDatabaseTool", func(t *testing.T) {
	// 	RunTest(t, "TestInvalidArguments", testUseDatabaseToolInvalidArguments)
	// 	RunTest(t, "TestSuccess", testUseDatabaseToolSuccess)
	// })
	// t.Run("TestCreateDatabaseTool", func(t *testing.T) {
	// 	RunTest(t, "TestInvalidArguments", testCreateDatabaseToolInvalidArguments)
	// 	RunTestWithTeardownSQL(t, "TestSuccess", testCreateDatabaseTeardownSQL, testCreateDatabaseToolSuccess)
	// })
	// t.Run("TestDropDatabaseTool", func(t *testing.T) {
	// 	RunTest(t, "TestInvalidArguments", testDropDatabaseToolInvalidArguments)
	// 	RunTestWithSetupSQL(t, "TestSuccess", testDropDatabaseSetupSQL, testDropDatabaseToolSuccess)
	// })
	RunTest(t, "TestShowTablesTool", testShowTablesTool)
	t.Run("TestDescribeTableTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDescribeTableToolInvalidArguments)
		RunTest(t, "TestSuccess", testDescribeTableToolSuccess)
	})
}
