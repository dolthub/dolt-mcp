package integration_tests

import (
	"database/sql"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"context"

	"errors"
	"os"
	"runtime"
	"syscall"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/dolthub/dolt-mcp/mcp/pkg/toolsets"
	"github.com/dolthub/dolt/go/performance/utils/benchmark_runner"
	"github.com/dolthub/dolt/go/store/constants"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	mcpTestDatabaseName         = "test"
	mcpTestRootUserName         = "root"
	mcpTestMCPServerSQLUser     = "mcp-client-1"
	mcpTestMCPServerSQLPassword = "passw0rd"
	mcpTestRootPassword         = ""
	doltServerHost              = "0.0.0.0"
	doltServerPort              = 3306
	mcpServerPort               = 8080
)

var ErrNoDatabaseConnection = errors.New("no database connection")

type testSuite struct {
	t                     *testing.T
	doltBinPath           string
	doltDatabaseParentDir string
	doltDatabaseDir       string
	dsn                   string
	testDb                *sql.DB
	doltServer            benchmark_runner.Server
	mcpServer             pkg.Server
	mcpErrGroup           *errgroup.Group
	mcpErrGroupCancelFunc context.CancelFunc
}

func (s *testSuite) Ping() error {
	if s.testDb == nil {
		return ErrNoDatabaseConnection
	}
	return s.testDb.Ping()
}

func (s *testSuite) GetMCPServerUrl() string {
	return "http://0.0.0.0:8080/mcp"
}

func (s *testSuite) checkoutBranch(branchName string) error {
	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database before checking out a branch: %s", err.Error())
	}
	_, err = s.testDb.Exec(fmt.Sprintf("CALL DOLT_CHECKOUT('%s');", branchName))
	return err
}

func (s *testSuite) createBranch(branchName string) error {
	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database before creating a branch: %s", err.Error())
	}
	_, err = s.testDb.Exec(fmt.Sprintf("CALL DOLT_BRANCH('-c', 'main', '%s');", branchName))
	return err
}

func (s *testSuite) deleteBranch(branchName string) error {
	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database before deleting a branch: %s", err.Error())
	}
	_, err = s.testDb.Exec(fmt.Sprintf("CALL DOLT_BRANCH('-D', '%s');", branchName))
	return err
}

func (s *testSuite) addAndCommitChanges(commitMessage string) error {
	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database before adding and committing changes: %s", err.Error())
	}
	_, err = s.testDb.Exec(fmt.Sprintf("CALL DOLT_COMMIT('-Am', '%s');", commitMessage))
	return err
}

func (s *testSuite) exec(sql string) error {
	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database before executing sql: %s", err.Error())
	}
	_, err = s.testDb.Exec(sql)
	return err
}

func (s *testSuite) Setup(newBranchName, setupSQL string, skipDoltCommit bool) {
	if newBranchName == "" {
		s.t.Fatalf("no new branch name provided")
	}

	err := s.exec(fmt.Sprintf("USE %s;", mcpTestDatabaseName))
	if err != nil {
		s.t.Fatalf("failed to use database during test setup: %s", err.Error())
	}

	err = s.checkoutBranch("main")
	if err != nil {
		s.t.Fatalf("failed checkout main branch during test setup: %s", err.Error())
	}

	err = s.createBranch(newBranchName)
	if err != nil {
		s.t.Fatalf("failed checkout generated branch during test setup: %s", err.Error())
	}

	err = s.checkoutBranch(newBranchName)
	if err != nil {
		s.t.Fatalf("failed checkout main branch during test setup: %s", err.Error())
	}

	if setupSQL != "" {
		err = s.exec(setupSQL)
		if err != nil {
			s.t.Fatalf("failed setup database with setup sql: %s", err.Error())
		}

		if !skipDoltCommit {
			err = s.addAndCommitChanges("add test setup changes")
			if err != nil {
				if !strings.Contains(err.Error(), "nothing to commit") {

					s.t.Fatalf("failed add and commit changes during test setup: %s", err.Error())
				}
			}
		}
	}
}

func (s *testSuite) Teardown(branchName, teardownSQL string, skipDoltCommit bool) {
	if branchName == "" {
		s.t.Fatalf("no new branch name provided")
	}

	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database: %s", err.Error())
	}

	err = s.exec(fmt.Sprintf("USE %s;", mcpTestDatabaseName))
	if err != nil {
		s.t.Fatalf("failed to use database during test setup: %s", err.Error())
	}

	if teardownSQL != "" {
		err = s.exec(teardownSQL)
		if err != nil {
			s.t.Fatalf("failed to execute teardown sql: %s", err.Error())
		}

		if !skipDoltCommit {
			err = s.addAndCommitChanges("teardown test changes")
			if err != nil {
				if !strings.Contains(err.Error(), "nothing to commit") {

					s.t.Fatalf("failed add and commit changes during test teardown: %s", err.Error())
				}
			}
		}
	}

	err = s.checkoutBranch("main")
	if err != nil {
		s.t.Fatalf("failed checkout main branch during test teardown: %s", err.Error())
	}

	err = s.deleteBranch(branchName)
	if err != nil {
		s.t.Fatalf("failed delete branch during test teardown: %s", err.Error())
	}
}

func createMCPDoltServerTestSuite(ctx context.Context, doltBinPath string) (*testSuite, error) {
	doltDatabaseParentDir, err := os.MkdirTemp("", "mcp-server-tests-*")
	if err != nil {
		return nil, err
	}

	doltDatabaseDir, err := benchmark_runner.InitDoltRepo(ctx, doltDatabaseParentDir, doltBinPath, constants.FormatDefaultString, mcpTestDatabaseName)
	if err != nil {
		return nil, err
	}

	serverArgs := []string{
		"-l",
		"debug",
	}

	doltServerConfig := benchmark_runner.NewDoltServerConfig(
		"",
		doltBinPath,
		mcpTestRootUserName,
		doltServerHost,
		"",
		"",
		benchmark_runner.CpuServerProfile,
		doltServerPort,
		serverArgs,
	)

	serverParams, err := doltServerConfig.GetServerArgs()
	if err != nil {
		return nil, err
	}

	doltServer := benchmark_runner.NewServer(ctx, doltDatabaseDir, doltServerConfig, syscall.SIGTERM, serverParams)
	err = doltServer.Start()
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true&parseTime=true", mcpTestRootUserName, mcpTestRootPassword, doltServerHost, doltServerPort, mcpTestDatabaseName)
	testDb, err := sql.Open("mysql", dsn)
	if err != nil {
		doltServer.Stop()
		return nil, err
	}

	err = testDb.PingContext(ctx)
	if err != nil {
		doltServer.Stop()
		testDb.Close()
		return nil, err
	}

	_, err = testDb.ExecContext(ctx, fmt.Sprintf("USE %s;", mcpTestDatabaseName))
	if err != nil {
		return nil, err
	}

	_, err = testDb.ExecContext(ctx, fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s';", mcpTestMCPServerSQLUser, "%", mcpTestMCPServerSQLPassword))
	if err != nil {
		return nil, err
	}

	_, err = testDb.ExecContext(ctx, fmt.Sprintf("GRANT ALL PRIVILEGES ON *.* TO '%s'@'%s';", mcpTestMCPServerSQLUser, "%"))
	if err != nil {
		return nil, err
	}

	seedSQLBytes, err := readSeedSQLFile()
	if err != nil {
		return nil, err
	}

	if len(seedSQLBytes) > 0 {
		_, err = testDb.ExecContext(ctx, string(seedSQLBytes))
		if err != nil {
			return nil, err
		}
		_, err = testDb.ExecContext(ctx, "CALL DOLT_COMMIT('-Am', 'seed test database');")
		if err != nil {
			return nil, err
		}
	}

	config := db.Config{
		Host:     doltServerHost,
		Port:     doltServerPort,
		User:     mcpTestMCPServerSQLUser,
		Password: mcpTestMCPServerSQLPassword,
	}

	logger := zap.NewNop()

	mcpServer, err := pkg.NewMCPHTTPServer(logger, config, mcpServerPort, toolsets.WithToolSet(&toolsets.PrimitiveToolSetV1{}))
	if err != nil {
		doltServer.Stop()
		testDb.Close()
		return nil, err
	}

	newCtx, cancelFunc := context.WithCancel(ctx)

	eg, egCtx := errgroup.WithContext(newCtx)

	eg.Go(func() error {
		mcpServer.ListenAndServe(egCtx)
		return nil
	})

	return &testSuite{
		doltBinPath:           doltBinPath,
		dsn:                   dsn,
		doltServer:            doltServer,
		doltDatabaseParentDir: doltDatabaseParentDir,
		doltDatabaseDir:       doltDatabaseDir,
		testDb:                testDb,
		mcpServer:             mcpServer,
		mcpErrGroup:           eg,
		mcpErrGroupCancelFunc: cancelFunc,
	}, nil
}

type FileRemoteDatabase struct {
	s                     *testSuite
	name                  string
	doltDatabaseDir       string
	doltDatabaseParentDir string
	doltServer            benchmark_runner.Server
	testDB                *sql.DB
	configFilePath        string
	remoteServePort       int
}

func NewFileRemoteDatabase(s *testSuite, name string) *FileRemoteDatabase {
	return &FileRemoteDatabase{
		s:               s,
		name:            name,
		remoteServePort: 2222,
	}
}

func (r *FileRemoteDatabase) Setup(ctx context.Context, setupSQL string) error {
	altServerPort := doltServerPort + 1
	configFilePath, err := generateFileRemoteDatabaseConfigFile(ctx, altServerPort, r.remoteServePort)
	if err != nil {
		return err
	}

	doltDatabaseParentDir, err := os.MkdirTemp("", "mcp-server-tests-remotes-*")
	if err != nil {
		return err
	}

	doltDatabaseDir, err := benchmark_runner.InitDoltRepo(ctx, doltDatabaseParentDir, r.s.doltBinPath, constants.FormatDefaultString, r.name)
	if err != nil {
		return err
	}

	serverArgs := []string{
		"--config",
		configFilePath,
	}

	doltServerConfig := benchmark_runner.NewDoltServerConfig(
		"",
		r.s.doltBinPath,
		mcpTestRootUserName,
		doltServerHost,
		"",
		"",
		benchmark_runner.CpuServerProfile,
		altServerPort,
		serverArgs,
	)

	serverParams, err := doltServerConfig.GetServerArgs()
	if err != nil {
		return err
	}

	doltServer := benchmark_runner.NewServer(ctx, doltDatabaseDir, doltServerConfig, syscall.SIGTERM, serverParams)
	err = doltServer.Start()
	if err != nil {
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true&parseTime=true", mcpTestRootUserName, mcpTestRootPassword, doltServerHost, altServerPort, r.name)

	testDb, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = testDb.PingContext(ctx)
	if err != nil {
		return err
	}

	if setupSQL != "" {
		_, err = testDb.ExecContext(ctx, setupSQL)
		if err != nil {
			return err
		}
	}

	r.configFilePath = configFilePath
	r.doltDatabaseDir = doltDatabaseDir
	r.doltDatabaseParentDir = doltDatabaseParentDir
	r.testDB = testDb
	r.doltServer = doltServer
	return nil
}

func (r *FileRemoteDatabase) Teardown(ctx context.Context) {
	if r.testDB != nil {
		r.testDB.Close()
		r.testDB = nil
	}
	if r.doltServer != nil {
		r.doltServer.Stop()
		r.doltServer = nil
	}
	if r.doltDatabaseParentDir != "" {
		os.RemoveAll(r.doltDatabaseParentDir)
		r.doltDatabaseParentDir = ""
	}
}

func (r *FileRemoteDatabase) GetRemoteURL() string {
	return fmt.Sprintf("http://localhost:%d/%s", r.remoteServePort, r.name)
}

func generateFileRemoteDatabaseConfigFile(ctx context.Context, dbPort, remoteServePort int) (configFilePath string, err error) {
	configBody := `log_level: debug 
log_format: text

behavior:
  read_only: false
  autocommit: true
  disable_client_multi_statements: false
  dolt_transaction_commit: false
  event_scheduler: "ON"
  auto_gc_behavior:
    enable: false
    archive_level: 0

listener:
  host: 0.0.0.0 
  port: ` + fmt.Sprintf("%d", dbPort) + ` 
remotesapi:
  port: ` + fmt.Sprintf("%d", remoteServePort) + ` 
`

	var file *os.File
	file, err = os.CreateTemp("", "config-*.yaml") // prefix "example-" and random suffix
	if err != nil {
		return
	}
	defer func() {
		rerr := file.Close()
		if err == nil {
			err = rerr
		}
	}()

	_, err = io.Copy(file, strings.NewReader(configBody))
	if err != nil {
		return
	}

	configFilePath = file.Name()
	return
}

func teardownMCPDoltServerTestSuite(s *testSuite) {
	if s == nil {
		return
	}

	defer func() {
		os.RemoveAll(s.doltDatabaseParentDir)
	}()

	if s.testDb != nil {
		s.testDb.Close()
		s.testDb = nil
	}

	if s.doltServer != nil {
		s.doltServer.Stop()
		s.doltServer = nil
	}

	if s.mcpErrGroup != nil && s.mcpErrGroupCancelFunc != nil {
		s.mcpErrGroupCancelFunc()
		s.mcpErrGroup.Wait()
	}
}

func readSeedSQLFile() ([]byte, error) {
	// Get the absolute path of the current source file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("could not determine caller path")
	}

	// Build path relative to this source file
	path := filepath.Join(filepath.Dir(filename), "testdata", "seed.sql")

	return os.ReadFile(path)
}
