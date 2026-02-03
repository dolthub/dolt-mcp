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
	doltHostFlag     = "dolt-host"
	doltPortFlag     = "dolt-port"
	doltUserFlag     = "dolt-user"
	doltPasswordFlag = "dolt-password"
	doltDatabaseFlag = "dolt-database"
	doltTLSFlag      = "dolt-tls"
	doltTLSCAFlag    = "dolt-tls-ca"
	mcpPortFlag      = "mcp-port"
	serveHTTPFlag    = "http"
	httpCertFlag     = "http-cert-file"
	httpKeyFlag      = "http-key-file"
	httpCAFlag       = "http-ca-file"
	serveStdioFlag   = "stdio"
	logLevelFlag     = "log-level"
	helpFlag         = "help"
	versionFlag      = "version"
	jwkClaimsFlag    = "jwk-claims"
	jwkURLFlag       = "jwk-url"
)

var doltHost = flag.String(doltHostFlag, "", "The hostname for the Dolt server.")
var doltPort = flag.Int(doltPortFlag, 3306, "The port for the Dolt server, default is 3306.")
var doltUser = flag.String(doltUserFlag, "", "The username for connecting to the Dolt server.")
var doltPassword = flag.String(doltPasswordFlag, "", "The password for connecting to the Dolt server.")
var doltDatabase = flag.String(doltDatabaseFlag, "", "The database for connecting to the Dolt server.")
var doltTLS = flag.String(doltTLSFlag, "", "TLS mode for Dolt server connection: 'true', 'false', 'skip-verify', or 'preferred'. Leave empty to disable TLS.")
var doltTLSCA = flag.String(doltTLSCAFlag, "", "Path to CA certificate file for Dolt server TLS connection. When provided, enables TLS with custom CA.")

var mcpPort = flag.Int(mcpPortFlag, 8080, "The HTTP port to serve Dolt MCP server on, default is 8080.")
var serveHTTP = flag.Bool(serveHTTPFlag, false, "If true, serves Dolt MCP server over HTTP")
var serveStdio = flag.Bool(serveStdioFlag, false, "If true, serves Dolt MCP server over stdio")
var logLevel = flag.String(logLevelFlag, "info", "Log level: debug, info, warn, error. Default is info.")
var httpCertFile = flag.String(httpCertFlag, "", "Path to TLS certificate file for HTTPS. If provided, a key must also be provided.")
var httpKeyFile = flag.String(httpKeyFlag, "", "Path to TLS private key file for HTTPS. If provided, a cert must also be be provided.")
var httpCAFile = flag.String(httpCAFlag, "", "Path to TLS CA certificate file for HTTPS. If provided, all TLS parameters must be provided otherwise it will be ignored.")
var jwkClaims = flag.String(jwkClaimsFlag, "", "A comma-separated list of key=value pairs for JWT claims for authentication.")
var jwkURL = flag.String(jwkURLFlag, "", "The URL of the JWKS server for JWT authentication.")

var help = flag.Bool(helpFlag, false, "If true, prints Dolt MCP server help information.")
var version = flag.Bool(versionFlag, false, "If true, prints the Dolt MCP server version.")

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

	err = validateArgs()
	if err != nil {
		logger.Fatal("invalid arguments", zap.Error(err))
	}

	envDoltPassword := os.Getenv("DOLT_PASSWORD")
	var finalPassword string
	if *doltPassword != "" {
		finalPassword = *doltPassword
	} else {
		finalPassword = envDoltPassword
	}

	config := db.Config{
		Host:         *doltHost,
		Port:         *doltPort,
		User:         *doltUser,
		Password:     finalPassword,
		DatabaseName: *doltDatabase,
		TLS:          *doltTLS,
		TLSCAFile:    *doltTLSCA,
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
	return errors.New(fmt.Sprintf("must supply %s", flg))
}

func validateArgs() error {
	if *doltHost == "" {
		return mustSupplyError(doltHostFlag)
	}
	if *doltPort == 0 {
		return mustSupplyError(doltPortFlag)
	}
	if *doltUser == "" {
		return mustSupplyError(doltUserFlag)
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
