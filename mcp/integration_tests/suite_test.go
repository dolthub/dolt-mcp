package integration_tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/google/uuid"
)

var suite *testSuite

func TestMain(m *testing.M) {
	ctx := context.Background()

	dialectType := db.DialectMySQL
	binName := "dolt"
	if d := os.Getenv("DOLT_DIALECT"); d != "" {
		switch strings.ToLower(d) {
		case "postgres", "postgresql", "pg", "doltgres", "doltgresql":
			dialectType = db.DialectPostgres
			binName = "doltgres"
		}
	}

	doltBinPath, err := exec.LookPath(binName)
	if err != nil {
		fmt.Printf("%s binary not found in PATH, skipping mcp test\n", binName)
		os.Exit(0)
	}

	suite, err = createMCPDoltServerTestSuite(ctx, doltBinPath, dialectType)
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
		suite.Setup(generatedTestBranchName, nil, false)
		defer suite.Teardown(generatedTestBranchName, nil, false)
		testFunc(suite, generatedTestBranchName)
	})
}

func RunTestWithSetupSQL(t *testing.T, testName string, setupSQL DialectSQL, testFunc func(s *testSuite, testBranchName string)) {
	t.Run(testName, func(t *testing.T) {
		if suite == nil {
			t.Fatalf("no test suite")
		}
		suite.t = t
		generatedTestBranchName := generateTestBranchName()
		suite.Setup(generatedTestBranchName, setupSQL, false)
		defer suite.Teardown(generatedTestBranchName, nil, false)
		testFunc(suite, generatedTestBranchName)
	})
}

func RunTestWithSetupSQLSkipDoltCommit(t *testing.T, testName string, setupSQL DialectSQL, testFunc func(s *testSuite, testBranchName string)) {
	t.Run(testName, func(t *testing.T) {
		if suite == nil {
			t.Fatalf("no test suite")
		}
		suite.t = t
		generatedTestBranchName := generateTestBranchName()
		suite.Setup(generatedTestBranchName, setupSQL, true)
		defer suite.Teardown(generatedTestBranchName, nil, false)
		testFunc(suite, generatedTestBranchName)
	})
}

func RunTestWithTeardownSQL(t *testing.T, testName string, teardownSQL DialectSQL, testFunc func(s *testSuite, testBranchName string)) {
	t.Run(testName, func(t *testing.T) {
		if suite == nil {
			t.Fatalf("no test suite")
		}
		suite.t = t
		generatedTestBranchName := generateTestBranchName()
		suite.Setup(generatedTestBranchName, nil, false)
		defer suite.Teardown(generatedTestBranchName, teardownSQL, false)
		testFunc(suite, generatedTestBranchName)
	})
}

func RunTestWithSetupAndTeardownSQL(t *testing.T, testName string, setupSQL, teardownSQL DialectSQL, testFunc func(s *testSuite, testBranchName string)) {
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

func RunTestWithSetupAndTeardownSQLSkipDoltCommit(t *testing.T, testName string, setupSQL, teardownSQL DialectSQL, testFunc func(s *testSuite, testBranchName string)) {
	t.Run(testName, func(t *testing.T) {
		if suite == nil {
			t.Fatalf("no test suite")
		}
		suite.t = t
		generatedTestBranchName := generateTestBranchName()
		suite.Setup(generatedTestBranchName, setupSQL, true)
		defer suite.Teardown(generatedTestBranchName, teardownSQL, true)
		testFunc(suite, generatedTestBranchName)
	})
}