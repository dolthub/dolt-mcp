package integration_tests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

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
	doltgresServerPort          = 5432
	doltgresRootUserName        = "postgres"
	doltgresRootPassword        = "password"
	mcpServerPort               = 8080
)

var ErrNoDatabaseConnection = errors.New("no database connection")

// DialectSQL maps a dialect type to the SQL for that dialect. The SQL may
// contain a single %s placeholder that will be substituted via fmt.Sprintf
// with the current test branch name when Setup/Teardown runs.
type DialectSQL map[db.DialectType]string

// Get returns the SQL for the given dialect, or "" if unset.
func (d DialectSQL) Get(dialectType db.DialectType) string {
	if d == nil {
		return ""
	}
	return d[dialectType]
}

// serverProcess abstracts the lifecycle of a database server process.
type serverProcess interface {
	Start() error
	Stop() error
}

type testSuite struct {
	t                     *testing.T
	dialect               db.Dialect
	dialectType           db.DialectType
	doltBinPath           string
	doltDatabaseParentDir string
	doltDatabaseDir       string
	dsn                   string
	testDb                *sql.DB
	doltServer            serverProcess
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

// formatBranchSQL substitutes each %s in sql with the branch name. If sql
// contains no %s, it is returned unchanged.
func formatBranchSQL(sql, branchName string) string {
	n := strings.Count(sql, "%s")
	if n == 0 {
		return sql
	}
	args := make([]any, n)
	for i := range args {
		args[i] = branchName
	}
	return fmt.Sprintf(sql, args...)
}

func (s *testSuite) checkoutBranch(branchName string) error {
	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database before checking out a branch: %s", err.Error())
	}
	_, err = s.testDb.Exec(s.dialect.CallProcedure(db.DoltCheckout, branchName))
	return err
}

func (s *testSuite) createBranch(branchName string) error {
	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database before creating a branch: %s", err.Error())
	}
	_, err = s.testDb.Exec(s.dialect.CallProcedure(db.DoltBranch, "-c", "main", branchName))
	return err
}

func (s *testSuite) deleteBranch(branchName string) error {
	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database before deleting a branch: %s", err.Error())
	}
	_, err = s.testDb.Exec(s.dialect.CallProcedure(db.DoltBranch, "-D", branchName))
	return err
}

func (s *testSuite) addAndCommitChanges(commitMessage string) error {
	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database before adding and committing changes: %s", err.Error())
	}
	_, err = s.testDb.Exec(s.dialect.CallProcedure(db.DoltCommit, "-Am", commitMessage))
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

// execStatements executes a multi-statement SQL string one statement at a time,
// which is required for the pgx driver (which does not support multiStatements).
func (s *testSuite) execStatements(sql string) error {
	for _, stmt := range splitSQLStatements(sql) {
		if err := s.exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// splitSQLStatements splits a SQL string into individual statements by ';'.
// Empty statements are dropped. This is a simple splitter that does not
// handle semicolons inside string literals - test SQL should avoid those.
func splitSQLStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	stmts := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			stmts = append(stmts, trimmed)
		}
	}
	return stmts
}

func (s *testSuite) Setup(newBranchName string, setupSQL DialectSQL, skipDoltCommit bool) {
	if newBranchName == "" {
		s.t.Fatalf("no new branch name provided")
	}

	err := s.exec(s.dialect.UseDatabase(mcpTestDatabaseName))
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

	sqlText := formatBranchSQL(setupSQL.Get(s.dialectType), newBranchName)
	if sqlText != "" {
		err = s.execStatements(sqlText)
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

func (s *testSuite) Teardown(branchName string, teardownSQL DialectSQL, skipDoltCommit bool) {
	if branchName == "" {
		s.t.Fatalf("no new branch name provided")
	}

	err := s.Ping()
	if err != nil {
		s.t.Fatalf("failed to reach database: %s", err.Error())
	}

	err = s.exec(s.dialect.UseDatabase(mcpTestDatabaseName))
	if err != nil {
		s.t.Fatalf("failed to use database during test setup: %s", err.Error())
	}

	sqlText := formatBranchSQL(teardownSQL.Get(s.dialectType), branchName)
	if sqlText != "" {
		err = s.execStatements(sqlText)
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

func createMCPDoltServerTestSuite(ctx context.Context, doltBinPath string, dialectType db.DialectType) (*testSuite, error) {
	switch dialectType {
	case db.DialectPostgres:
		return createDoltgresTestSuite(ctx, doltBinPath)
	default:
		return createDoltTestSuite(ctx, doltBinPath)
	}
}

func createDoltTestSuite(ctx context.Context, doltBinPath string) (*testSuite, error) {
	dialectType := db.DialectMySQL
	dialect := db.NewDialect(dialectType)

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

	dsnConfig := db.Config{
		Host:            doltServerHost,
		Port:            doltServerPort,
		User:            mcpTestRootUserName,
		Password:        mcpTestRootPassword,
		DatabaseName:    mcpTestDatabaseName,
		ParseTime:       true,
		MultiStatements: true,
		DialectType:     dialectType,
	}
	dsn := dialect.FormatDSN(dsnConfig)
	testDb, err := sql.Open(dialect.DriverName(), dsn)
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

	_, err = testDb.ExecContext(ctx, dialect.UseDatabase(mcpTestDatabaseName))
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

	err = seedDatabase(ctx, testDb, dialect)
	if err != nil {
		return nil, err
	}

	mcpConfig := db.Config{
		Host:        doltServerHost,
		Port:        doltServerPort,
		User:        mcpTestMCPServerSQLUser,
		Password:    mcpTestMCPServerSQLPassword,
		DialectType: dialectType,
	}

	return finishTestSuite(ctx, dialect, dialectType, doltBinPath, doltDatabaseParentDir, doltDatabaseDir, dsn, testDb, doltServer, mcpConfig)
}

func createDoltgresTestSuite(ctx context.Context, doltgresBinPath string) (*testSuite, error) {
	dialectType := db.DialectPostgres
	dialect := db.NewDialect(dialectType)

	dataDir, err := os.MkdirTemp("", "mcp-server-tests-doltgres-*")
	if err != nil {
		return nil, err
	}

	server := &doltgresServer{
		binPath: doltgresBinPath,
		dataDir: dataDir,
	}
	err = server.Start()
	if err != nil {
		os.RemoveAll(dataDir)
		return nil, err
	}

	// Connect as the default postgres superuser to set up the test database
	adminDsnConfig := db.Config{
		Host:        doltServerHost,
		Port:        doltgresServerPort,
		User:        doltgresRootUserName,
		Password:    doltgresRootPassword,
		DialectType: dialectType,
	}
	adminDsn := dialect.FormatDSN(adminDsnConfig)
	adminDb, err := sql.Open(dialect.DriverName(), adminDsn)
	if err != nil {
		server.Stop()
		os.RemoveAll(dataDir)
		return nil, err
	}

	err = adminDb.PingContext(ctx)
	if err != nil {
		adminDb.Close()
		server.Stop()
		os.RemoveAll(dataDir)
		return nil, err
	}

	_, err = adminDb.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dialect.QuoteIdentifier(mcpTestDatabaseName)))
	if err != nil {
		adminDb.Close()
		server.Stop()
		os.RemoveAll(dataDir)
		return nil, err
	}
	adminDb.Close()

	testDsnConfig := db.Config{
		Host:         doltServerHost,
		Port:         doltgresServerPort,
		User:         doltgresRootUserName,
		Password:     doltgresRootPassword,
		DatabaseName: mcpTestDatabaseName,
		DialectType:  dialectType,
	}
	dsn := dialect.FormatDSN(testDsnConfig)
	testDb, err := sql.Open(dialect.DriverName(), dsn)
	if err != nil {
		server.Stop()
		os.RemoveAll(dataDir)
		return nil, err
	}

	err = testDb.PingContext(ctx)
	if err != nil {
		testDb.Close()
		server.Stop()
		os.RemoveAll(dataDir)
		return nil, err
	}

	err = seedDatabase(ctx, testDb, dialect)
	if err != nil {
		return nil, err
	}

	// MCP server connects as the same superuser for DoltgreSQL
	mcpConfig := db.Config{
		Host:        doltServerHost,
		Port:        doltgresServerPort,
		User:        doltgresRootUserName,
		Password:    doltgresRootPassword,
		DialectType: dialectType,
	}

	return finishTestSuite(ctx, dialect, dialectType, doltgresBinPath, dataDir, dataDir, dsn, testDb, server, mcpConfig)
}

// seedDatabase loads and executes the dialect-appropriate seed SQL, then commits.
func seedDatabase(ctx context.Context, testDb *sql.DB, dialect db.Dialect) error {
	seedSQLBytes, err := readSeedSQLFile()
	if err != nil {
		return err
	}

	if len(seedSQLBytes) == 0 {
		return nil
	}

	// Execute each statement individually to support both mysql and pgx drivers.
	for _, stmt := range splitSQLStatements(string(seedSQLBytes)) {
		if _, err := testDb.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute seed statement %q: %w", stmt, err)
		}
	}

	if _, err := testDb.ExecContext(ctx, dialect.CallProcedure(db.DoltCommit, "-Am", "seed test database")); err != nil {
		return err
	}
	return nil
}

func finishTestSuite(ctx context.Context, dialect db.Dialect, dialectType db.DialectType, binPath, databaseParentDir, databaseDir, dsn string, testDb *sql.DB, server serverProcess, mcpConfig db.Config) (*testSuite, error) {
	logger := zap.NewNop()

	mcpServer, err := pkg.NewMCPHTTPServer(logger, mcpConfig, mcpServerPort, nil, "", nil, toolsets.WithToolSet(&toolsets.PrimitiveToolSetV1{}))
	if err != nil {
		server.Stop()
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
		dialect:               dialect,
		dialectType:           dialectType,
		doltBinPath:           binPath,
		dsn:                   dsn,
		doltServer:            server,
		doltDatabaseParentDir: databaseParentDir,
		doltDatabaseDir:       databaseDir,
		testDb:                testDb,
		mcpServer:             mcpServer,
		mcpErrGroup:           eg,
		mcpErrGroupCancelFunc: cancelFunc,
	}, nil
}

// doltgresServer manages a DoltgreSQL server process.
type doltgresServer struct {
	binPath string
	dataDir string
	cmd     *exec.Cmd
}

func (s *doltgresServer) Start() error {
	s.cmd = exec.Command(s.binPath)
	s.cmd.Env = append(os.Environ(), fmt.Sprintf("DOLTGRES_DATA_DIR=%s", s.dataDir))
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start doltgres: %w", err)
	}

	// Poll the connection until the server is ready.
	dialect := db.NewPostgresDialect()
	dsn := dialect.FormatDSN(db.Config{
		Host:     doltServerHost,
		Port:     doltgresServerPort,
		User:     doltgresRootUserName,
		Password: doltgresRootPassword,
	})

	for i := 0; i < 30; i++ {
		testDb, err := sql.Open(dialect.DriverName(), dsn)
		if err == nil {
			if err = testDb.Ping(); err == nil {
				testDb.Close()
				fmt.Println("Successfully started database server")
				return nil
			}
			testDb.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}

	s.Stop()
	return fmt.Errorf("doltgres server failed to become ready")
}

func (s *doltgresServer) Stop() error {
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Signal(syscall.SIGTERM)
		s.cmd.Wait()
		fmt.Println("Successfully killed database server")
	}
	return nil
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

	remoteDsnConfig := db.Config{
		Host:            doltServerHost,
		Port:            altServerPort,
		User:            mcpTestRootUserName,
		Password:        mcpTestRootPassword,
		DatabaseName:    r.name,
		ParseTime:       true,
		MultiStatements: true,
	}
	mysqlDialect := db.NewMySQLDialect()
	dsn := mysqlDialect.FormatDSN(remoteDsnConfig)

	testDb, err := sql.Open(mysqlDialect.DriverName(), dsn)
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
	file, err = os.CreateTemp("", "config-*.yaml")
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
