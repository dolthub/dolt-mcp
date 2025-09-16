package integration_tests

import (
	"testing"
)

func TestTools(t *testing.T) {
	RunTest(t, "TestListDatabasesTool", testListDatabasesTool)
	t.Run("TestListDoltBranchesTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testListDoltBranchesToolInvalidArguments)
		RunTest(t, "TestSuccess", testListDoltBranchesToolSuccess)
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
	t.Run("TestMoveDoltBranchTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testMoveDoltBranchToolInvalidArguments)
		RunTestWithSetupAndTeardownSQL(t, "TestSuccess", testMoveDoltBranchSetupSQL, testMoveDoltBranchTeardownSQL, testMoveDoltBranchToolSuccess)
	})
	t.Run("TestDeleteDoltBranchTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDeleteDoltBranchToolInvalidArguments)
		RunTestWithSetupSQL(t, "TestSuccess", testDeleteDoltBranchSetupSQL, testDeleteDoltBranchToolSuccess)
	})
	t.Run("TestStageTableForDoltCommitTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testStageTableForDoltCommitToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testStageTableForDoltCommitSetupSQL, testStageTableForDoltCommitToolSuccess)
	})
	t.Run("TestStageAllTablesForDoltCommitTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testStageAllTablesForDoltCommitToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testStageAllTablesForDoltCommitSetupSQL, testStageAllTablesForDoltCommitToolSuccess)
	})
	t.Run("TestUnstageTableTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testUnstageTableToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testUnstageTableSetupSQL, testUnstageTableToolSuccess)
	})
	t.Run("TestUnstageAllTablesTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testUnstageAllTablesToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testUnstageAllTablesSetupSQL, testUnstageAllTablesToolSuccess)
	})
	t.Run("TestCreateDoltCommitTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testCreateDoltCommitToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testCreateDoltCommitSetupSQL, testCreateDoltCommitToolSuccess)
	})
	t.Run("TestDoltResetTableSoftTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDoltResetTableSoftToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testDoltResetTableSoftSetupSQL, testDoltResetTableSoftToolSuccess)
	})
	t.Run("TestDoltResetAllTablesSoftTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDoltResetAllTablesSoftToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testDoltResetAllTablesSoftSetupSQL, testDoltResetAllTablesSoftToolSuccess)
	})
	t.Run("TestDoltResetHardTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDoltResetHardToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testDoltResetHardSetupSQL, testDoltResetHardToolSuccess)
	})
	t.Run("TestListDoltCommitsTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testListDoltCommitsToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testListDoltCommitsSetupSQL, testListDoltCommitsToolSuccess)
	})
	t.Run("TestListDoltDiffChangesInWorkingSetTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testListDoltDiffChangesInWorkingSetToolInvalidArguments)
		RunTestWithSetupSQLSkipDoltCommit(t, "TestSuccess", testListDoltDiffChangesInWorkingSetSetupSQL, testListDoltDiffChangesInWorkingSetToolSuccess)
	})
	t.Run("TestListDoltDiffChangesByTableNameTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testListDoltDiffChangesByTableNameToolInvalidArguments)
		RunTestWithSetupSQL(t, "TestSuccess", testListDoltDiffChangesByTableNameSetupSQL, testListDoltDiffChangesByTableNameToolSuccess)
	})
	t.Run("TestListDoltDiffChangesInDateRangeTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testListDoltDiffChangesInDateRangeToolInvalidArguments)
		RunTestWithSetupSQL(t, "TestSuccess", testListDoltDiffChangesInDateRangeSetupSQL, testListDoltDiffChangesInDateRangeToolSuccess)
	})
	t.Run("TestGetDoltMergeStatusTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testGetDoltMergeStatusToolInvalidArguments)
		RunTest(t, "TestSuccess", testGetDoltMergeStatusToolSuccess)
	})
	t.Run("TestMergeDoltBranchTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testMergeDoltBranchToolInvalidArguments)
		RunTestWithSetupAndTeardownSQL(t, "TestSuccess", testMergeDoltBranchSetupSQL, testMergeDoltBranchTeardownSQL, testMergeDoltBranchToolSuccess)
	})
	t.Run("TestMergeDoltBranchNoFastForwardTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testMergeDoltBranchNoFastForwardToolInvalidArguments)
		RunTestWithSetupAndTeardownSQL(t, "TestSuccess", testMergeDoltBranchNoFastForwardSetupSQL, testMergeDoltBranchNoFastForwardTeardownSQL, testMergeDoltBranchNoFastForwardToolSuccess)
	})
	t.Run("TestListDoltRemotesTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testListDoltRemotesToolInvalidArguments)
		RunTestWithSetupAndTeardownSQL(t, "TestSuccess", testListDoltRemotesSetupSQL, testListDoltRemotesTeardownSQL, testListDoltRemotesToolSuccess)
	})
	t.Run("TestAddDoltRemoteTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testAddDoltRemoteToolInvalidArguments)
		RunTestWithTeardownSQL(t, "TestSuccess", testAddDoltRemoteTeardownSQL, testAddDoltRemoteToolSuccess)
	})
	t.Run("TestRemoveDoltRemoteTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testRemoveDoltRemoteToolInvalidArguments)
		RunTestWithSetupSQL(t, "TestSuccess", testRemoveDoltRemoteSetupSQL, testRemoveDoltRemoteToolSuccess)
	})
	t.Run("TestCloneDatabaseTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testCloneDatabaseToolInvalidArguments)
		RunTestWithTeardownSQL(t, "TestSuccess", testCloneDatabaseTeardownSQL, testCloneDatabaseToolSuccess)
	})
	t.Run("TestDoltFetchBranchTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDoltFetchBranchToolInvalidArguments)
		RunTestWithSetupAndTeardownSQL(t, "TestSuccess",testDoltFetchBranchSetupSQL, testDoltFetchBranchTeardownSQL, testDoltFetchBranchToolSuccess)
	})
	t.Run("TestDoltFetchAllBranchesTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDoltFetchAllBranchesToolInvalidArguments)
		RunTestWithSetupAndTeardownSQL(t, "TestSuccess", testDoltFetchAllBranchesSetupSQL, testDoltFetchAllBranchesTeardownSQL, testDoltFetchAllBranchesToolSuccess)
	})
	t.Run("TestDoltPushBranchTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDoltPushBranchToolInvalidArguments)
		RunTestWithSetupAndTeardownSQL(t, "TestSuccess", testDoltPushBranchSetupSQL, testDoltPushBranchTeardownSQL, testDoltPushBranchToolSuccess)
	})
	t.Run("TestDoltPullBranchTool", func(t *testing.T) {
		RunTest(t, "TestInvalidArguments", testDoltPullBranchToolInvalidArguments)
		RunTestWithTeardownSQL(t, "TestSuccess", testDoltPullBranchTeardownSQL, testDoltPullBranchToolSuccess)
	})

    // Dolt tests tools
    t.Run("TestRunDoltTestsTool", func(t *testing.T) {
        RunTest(t, "TestInvalidArguments", testRunDoltTestsToolInvalidArguments)
        RunTestWithSetupSQL(t, "TestSuccess", testRunDoltTestsSetupSQL, testRunDoltTestsToolSuccess)
    })
    t.Run("TestAddRemoveDoltTestTools", func(t *testing.T) {
        RunTest(t, "TestInvalidArguments", testAddDoltTestToolInvalidArguments)
        RunTest(t, "TestAddSuccess", testAddDoltTestToolSuccess)
        RunTest(t, "TestRemoveInvalidArguments", testRemoveDoltTestToolInvalidArguments)
        RunTestWithSetupSQL(t, "TestRemoveSuccess", testRemoveDoltTestSetupSQL, testRemoveDoltTestToolSuccess)
    })
}
