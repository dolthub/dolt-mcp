package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/dolthub/dolt-mcp/mcp/pkg/toolsets"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// New flag names (preferred).
	hostFlag     = "host"
	portFlag     = "port"
	userFlag     = "user"
	passwordFlag = "password"
	databaseFlag = "database"
	tlsFlag      = "tls"
	tlsCAFlag    = "tls-ca"
	doltFlag     = "dolt"
	doltgresFlag = "doltgres"

	// Deprecated flag names (kept for backwards compatibility).
	doltHostFlag     = "dolt-host"
	doltPortFlag     = "dolt-port"
	doltUserFlag     = "dolt-user"
	doltPasswordFlag = "dolt-password"
	doltDatabaseFlag = "dolt-database"
	doltTLSFlag      = "dolt-tls"
	doltTLSCAFlag    = "dolt-tls-ca"

	// Unchanged flags.
	mcpPortFlag    = "mcp-port"
	serveHTTPFlag  = "http"
	httpCertFlag   = "http-cert-file"
	httpKeyFlag    = "http-key-file"
	httpCAFlag     = "http-ca-file"
	serveStdioFlag = "stdio"
	logLevelFlag   = "log-level"
	helpFlag       = "help"
	versionFlag    = "version"
	jwkClaimsFlag  = "jwk-claims"
	jwkURLFlag     = "jwk-url"
)

// Default ports per dialect.
const (
	defaultDoltPort     = 3306
	defaultDoltgresPort = 5432
)

// New flags (preferred).
var (
	host     = flag.String(hostFlag, "", "The hostname for the database server.")
	port     = flag.Int(portFlag, 0, "The port for the database server. Defaults to 3306 for Dolt or 5432 for DoltgreSQL.")
	user     = flag.String(userFlag, "", "The username for connecting to the database server.")
	password = flag.String(passwordFlag, "", "The password for connecting to the database server.")
	database = flag.String(databaseFlag, "", "The database name for connecting to the server.")
	tlsMode  = flag.String(tlsFlag, "", "TLS mode for the database connection: 'true', 'false', 'skip-verify', or 'preferred'. Leave empty to disable TLS.")
	tlsCA    = flag.String(tlsCAFlag, "", "Path to CA certificate file for the database TLS connection. When provided, enables TLS with custom CA.")
	useDolt  = flag.Bool(doltFlag, false, "Use the Dolt (MySQL-compatible) dialect. This is the default when neither --dolt nor --doltgres is specified.")
	doltgres = flag.Bool(doltgresFlag, false, "Use the DoltgreSQL (PostgreSQL-compatible) dialect.")
)

// Deprecated flags (kept for backwards compatibility).
var (
	doltHost     = flag.String(doltHostFlag, "", "DEPRECATED: use --host instead.")
	doltPort     = flag.Int(doltPortFlag, 0, "DEPRECATED: use --port instead.")
	doltUser     = flag.String(doltUserFlag, "", "DEPRECATED: use --user instead.")
	doltPassword = flag.String(doltPasswordFlag, "", "DEPRECATED: use --password instead.")
	doltDatabase = flag.String(doltDatabaseFlag, "", "DEPRECATED: use --database instead.")
	doltTLS      = flag.String(doltTLSFlag, "", "DEPRECATED: use --tls instead.")
	doltTLSCA    = flag.String(doltTLSCAFlag, "", "DEPRECATED: use --tls-ca instead.")
)

// Unchanged flags.
var (
	mcpPort      = flag.Int(mcpPortFlag, 8080, "The HTTP port to serve Dolt MCP server on, default is 8080.")
	serveHTTP    = flag.Bool(serveHTTPFlag, false, "If true, serves Dolt MCP server over HTTP")
	serveStdio   = flag.Bool(serveStdioFlag, false, "If true, serves Dolt MCP server over stdio")
	logLevel     = flag.String(logLevelFlag, "info", "Log level: debug, info, warn, error. Default is info.")
	httpCertFile = flag.String(httpCertFlag, "", "Path to TLS certificate file for HTTPS. If provided, a key must also be provided.")
	httpKeyFile  = flag.String(httpKeyFlag, "", "Path to TLS private key file for HTTPS. If provided, a cert must also be be provided.")
	httpCAFile   = flag.String(httpCAFlag, "", "Path to TLS CA certificate file for HTTPS. If provided, all TLS parameters must be provided otherwise it will be ignored.")
	jwkClaims    = flag.String(jwkClaimsFlag, "", "A comma-separated list of key=value pairs for JWT claims for authentication.")
	jwkURL       = flag.String(jwkURLFlag, "", "The URL of the JWKS server for JWT authentication.")
	help         = flag.Bool(helpFlag, false, "If true, prints Dolt MCP server help information.")
	version      = flag.Bool(versionFlag, false, "If true, prints the Dolt MCP server version.")
)

// setFlags returns the set of flag names that were explicitly passed on the command line.
func setFlags() map[string]bool {
	s := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { s[f.Name] = true })
	return s
}

// coalesceStringFlag returns the value of the new flag if set, otherwise the
// value of the deprecated flag (with a deprecation warning), otherwise "".
func coalesceStringFlag(set map[string]bool, newName, newVal, oldName, oldVal string) string {
	if set[newName] {
		return newVal
	}
	if set[oldName] {
		fmt.Fprintf(os.Stderr, "warning: --%s is deprecated, use --%s instead\n", oldName, newName)
		return oldVal
	}
	return newVal
}

// coalesceIntFlag is the int version of coalesceStringFlag.
func coalesceIntFlag(set map[string]bool, newName string, newVal int, oldName string, oldVal int) int {
	if set[newName] {
		return newVal
	}
	if set[oldName] {
		fmt.Fprintf(os.Stderr, "warning: --%s is deprecated, use --%s instead\n", oldName, newName)
		return oldVal
	}
	return newVal
}

// resolveDialect determines the dialect from the --dolt/--doltgres flags.
// Defaults to Dolt (MySQL) if neither is set.
func resolveDialect(set map[string]bool) (db.DialectType, error) {
	if set[doltFlag] && set[doltgresFlag] {
		return "", errors.New("--dolt and --doltgres are mutually exclusive")
	}
	if set[doltgresFlag] && *doltgres {
		return db.DialectPostgres, nil
	}
	return db.DialectMySQL, nil
}

func getTLSConfig(cert, key, ca string) (*tls.Config, error) {
	if key == "" && cert == "" {
		return nil, nil
	}

	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("tls.LoadX509KeyPair(%v, %v) failed: %w", cert, key, err)
	}

	var caCertPool *x509.CertPool
	if ca != "" {
		caCertPEM, err := os.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("unable to read CA file at %s: %w", ca, err)
		}

		caCertPool = x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCertPEM); !ok {
			return nil, fmt.Errorf("unable to add CA cert to cert pool")
		}
	}

	return &tls.Config{
		Certificates: []tls.Certificate{c},
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	cfg := zap.NewProductionConfig()
	switch *logLevel {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn", "warning":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		// Fallback to info if unknown level provided
		fmt.Fprintf(os.Stderr, "Unknown log level '%s', defaulting to info\n", *logLevel)
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	// Keep timestamps readable
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// If running in stdio mode with debug, write logs to a file to avoid contaminating stdout
	if *serveStdio && *logLevel == "debug" {
		homeDir, herr := os.UserHomeDir()
		if herr == nil && homeDir != "" {
			logsDir := filepath.Join(homeDir, ".dolt-mcp-server", "logs")
			_ = os.MkdirAll(logsDir, os.ModePerm)
			ts := time.Now().Format("20060102-150405")
			logFile := filepath.Join(logsDir, fmt.Sprintf("%s.log", ts))
			cfg.OutputPaths = []string{logFile}
			cfg.ErrorOutputPaths = []string{logFile}
		}
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	if *version {
		logger.Info("Dolt MCP server", zap.String("version", pkg.DoltMCPServerVersion))
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Resolve flag values, preferring new flags over deprecated ones.
	set := setFlags()
	hostVal := coalesceStringFlag(set, hostFlag, *host, doltHostFlag, *doltHost)
	portVal := coalesceIntFlag(set, portFlag, *port, doltPortFlag, *doltPort)
	userVal := coalesceStringFlag(set, userFlag, *user, doltUserFlag, *doltUser)
	passwordVal := coalesceStringFlag(set, passwordFlag, *password, doltPasswordFlag, *doltPassword)
	databaseVal := coalesceStringFlag(set, databaseFlag, *database, doltDatabaseFlag, *doltDatabase)
	tlsVal := coalesceStringFlag(set, tlsFlag, *tlsMode, doltTLSFlag, *doltTLS)
	tlsCAVal := coalesceStringFlag(set, tlsCAFlag, *tlsCA, doltTLSCAFlag, *doltTLSCA)

	dialectType, err := resolveDialect(set)
	if err != nil {
		logger.Fatal("invalid dialect flags", zap.Error(err))
	}

	// Apply the dialect-appropriate default port if no port was explicitly set.
	if portVal == 0 {
		switch dialectType {
		case db.DialectPostgres:
			portVal = defaultDoltgresPort
		default:
			portVal = defaultDoltPort
		}
	}

	if err := validateArgs(hostVal, userVal, portVal); err != nil {
		logger.Fatal("invalid arguments", zap.Error(err))
	}

	// Password may come from the DOLT_PASSWORD environment variable.
	if passwordVal == "" {
		passwordVal = os.Getenv("DOLT_PASSWORD")
	}

	config := db.Config{
		Host:         hostVal,
		Port:         portVal,
		User:         userVal,
		Password:     passwordVal,
		DatabaseName: databaseVal,
		TLS:          tlsVal,
		TLSCAFile:    tlsCAVal,
		DialectType:  dialectType,
	}

	tlsConfig, err := getTLSConfig(*httpCertFile, *httpKeyFile, *httpCAFile)
	if err != nil {
		logger.Fatal("failed to get TLS configuration", zap.Error(err))
	}

	jwkClaimsMap, err := parseClaimsMap(jwkClaims)
	if err != nil {
		logger.Fatal("failed to parse JWK claims", zap.Stringp("jwk_claims", jwkClaims), zap.Error(err))
	}

	if *serveHTTP {
		srv, err := pkg.NewMCPHTTPServer(
			logger,
			config,
			*mcpPort,
			jwkClaimsMap,
			*jwkURL,
			tlsConfig,
			toolsets.WithToolSet(&toolsets.PrimitiveToolSetV1{}))
		if err != nil {
			logger.Fatal("failed to create Dolt MCP HTTP server", zap.Error(err))
		}

		srv.ListenAndServe(context.Background())
	} else if *serveStdio {
		srv, err := pkg.NewMCPStdioServer(
			logger,
			config,
			toolsets.WithToolSet(&toolsets.PrimitiveToolSetV1{}),
		)
		if err != nil {
			logger.Fatal("failed to create Dolt MCP stdio server", zap.Error(err))
		}
		srv.ServeStdio(context.Background())
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func mustSupplyError(flg string) error {
	return errors.New(fmt.Sprintf("must supply --%s", flg))
}

func validateArgs(host, user string, port int) error {
	if host == "" {
		return mustSupplyError(hostFlag)
	}
	if port == 0 {
		return mustSupplyError(portFlag)
	}
	if user == "" {
		return mustSupplyError(userFlag)
	}
	if *serveHTTP {
		if *mcpPort == 0 {
			return mustSupplyError(mcpPortFlag)
		}
	}
	return nil
}

func parseClaimsMap(jwkClaims *string) (map[string]string, error) {
	if jwkClaims == nil || *jwkClaims == "" {
		return nil, nil
	}

	claimsMap := make(map[string]string)
	items := splitAndTrim(*jwkClaims, ",")
	for _, item := range items {
		tup := splitAndTrim(item, "=")
		if len(tup) != 2 {
			return nil, errors.New("invalid format for jwk-claims, must be key=value pairs separated by commas")
		}
		claimsMap[tup[0]] = tup[1]
	}
	return claimsMap, nil
}

func splitAndTrim(s string, sep string) []string {
	parts := []string{}
	for _, part := range strings.Split(s, sep) {
		parts = append(parts, strings.TrimSpace(part))
	}

	return parts
}