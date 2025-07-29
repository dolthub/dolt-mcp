package integration_tests

import (
	"fmt"
	"testing"

	"context"

	"os"
	"os/exec"

	"github.com/google/uuid"
)

var suite *testSuite

func TestMain(m *testing.M) {
	ctx := context.Background()

	doltBinPath, err := exec.LookPath("dolt")
	if err != nil {
		fmt.Println("dolt binary not found in PATH, skipping mcp test")
		os.Exit(0)
	}

	suite, err = createMCPDoltServerTestSuite(ctx, doltBinPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create dolt server test suite: %v\n", err)
		teardownMCPDoltServerTestSuite(suite)
		os.Exit(1)
	}

	code := m.Run()

	teardownMCPDoltServerTestSuite(suite)

	os.Exit(code)
}

func generateTestBranchName() string {
	return uuid.NewString()
}

func RunTest(t *testing.T, testName string, testFunc func(s *testSuite, testBranchName string)) {
	t.Run(testName, func(t *testing.T) {
		if suite == nil {
			t.Fatalf("no test suite")
		}
		suite.t = t
		generatedTestBranchName := generateTestBranchName()
		suite.Setup(generatedTestBranchName, "", false)
		defer suite.Teardown(generatedTestBranchName, "", false)
		testFunc(suite, generatedTestBranchName)
	})
}

func RunTestWithSetupSQL(t *testing.T, testName, setupSQL string, testFunc func(s *testSuite, testBranchName string)) {
	t.Run(testName, func(t *testing.T) {
		if suite == nil {
			t.Fatalf("no test suite")
		}
		suite.t = t
		generatedTestBranchName := generateTestBranchName()
		suite.Setup(generatedTestBranchName, setupSQL, false)
		defer suite.Teardown(generatedTestBranchName, "", false)
		testFunc(suite, generatedTestBranchName)
	})
}

func RunTestWithSetupSQLSkipDoltCommit(t *testing.T, testName, setupSQL string, testFunc func(s *testSuite, testBranchName string)) {
	t.Run(testName, func(t *testing.T) {
		if suite == nil {
			t.Fatalf("no test suite")
		}
		suite.t = t
		generatedTestBranchName := generateTestBranchName()
		suite.Setup(generatedTestBranchName, setupSQL, true)
		defer suite.Teardown(generatedTestBranchName, "", false)
		testFunc(suite, generatedTestBranchName)
	})
}

func RunTestWithTeardownSQL(t *testing.T, testName, teardownSQL string, testFunc func(s *testSuite, testBranchName string)) {
	t.Run(testName, func(t *testing.T) {
		if suite == nil {
			t.Fatalf("no test suite")
		}
		suite.t = t
		generatedTestBranchName := generateTestBranchName()
		suite.Setup(generatedTestBranchName, "", false)
		defer suite.Teardown(generatedTestBranchName, teardownSQL, false)
		testFunc(suite, generatedTestBranchName)
	})
}

func RunTestWithSetupAndTeardownSQL(t *testing.T, testName, setupSQL, teardownSQL string, testFunc func(s *testSuite, testBranchName string)) {
	t.Run(testName, func(t *testing.T) {
		if suite == nil {
			t.Fatalf("no test suite")
		}
		suite.t = t
		generatedTestBranchName := generateTestBranchName()
		suite.Setup(generatedTestBranchName, setupSQL, false)
		defer suite.Teardown(generatedTestBranchName, teardownSQL, false)
		testFunc(suite, generatedTestBranchName)
	})
}

