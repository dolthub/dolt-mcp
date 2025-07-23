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
}

