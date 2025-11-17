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
	mcpPortFlag      = "mcp-port"
	serveHTTPFlag    = "http"
	httpCertFlag     = "http-cert-file"
	httpKeyFlag      = "http-key-file"
	httpCAFlag       = "http-ca-file"
	serveStdioFlag   = "stdio"
	logLevelFlag     = "log-level"
	helpFlag         = "help"
	versionFlag      = "version"
)

var doltHost = flag.String(doltHostFlag, "", "The hostname for the Dolt server.")
var doltPort = flag.Int(doltPortFlag, 3306, "The port for the Dolt server, default is 3306.")
var doltUser = flag.String(doltUserFlag, "", "The username for connecting to the Dolt server.")
var doltPassword = flag.String(doltPasswordFlag, "", "The password for connecting to the Dolt server.")
var doltDatabase = flag.String(doltDatabaseFlag, "", "The database for connecting to the Dolt server.")

var mcpPort = flag.Int(mcpPortFlag, 8080, "The HTTP port to serve Dolt MCP server on, default is 8080.")
var serveHTTP = flag.Bool(serveHTTPFlag, false, "If true, serves Dolt MCP server over HTTP")
var serveStdio = flag.Bool(serveStdioFlag, false, "If true, serves Dolt MCP server over stdio")
var logLevel = flag.String(logLevelFlag, "info", "Log level: debug, info, warn, error. Default is info.")
var httpCertFile = flag.String(httpCertFlag, "", "Path to TLS certificate file for HTTPS. If provided, all TLS parameters must be provided.")
var httpKeyFile = flag.String(httpKeyFlag, "", "Path to TLS private key file for HTTPS. If provided, all TLS parameters must be provided.")
var httpCAFile = flag.String(httpCAFlag, "", "Path to TLS CA certificate file for HTTPS. If provided, all TLS parameters must be provided.")

var help = flag.Bool(helpFlag, false, "If true, prints Dolt MCP server help information.")
var version = flag.Bool(versionFlag, false, "If true, prints the Dolt MCP server version.")

func getTlsConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	if len(certFile) == 0 && len(keyFile) == 0 && len(caFile) == 0 {
		return nil, nil
	} else if len(certFile) == 0 || len(keyFile) == 0 || len(caFile) == 0 {
		return nil, fmt.Errorf("all of certFile, keyFile, and caFile must be provided for TLS configuration")
	}

	// Load server certificate and private key
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("error loading server certificate '%s' and key '%s': %w", certFile, keyFile, err)
	}

	// Load CA certificate for client authentication (if using mutual TLS)
	caCertPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading CA certificate '%s': %w", caFile, err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("failed to append CA certificate from '%s'", caFile)
	}

	// Create tls.Config
	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert, // For mutual TLS, require client cert
		MinVersion:   tls.VersionTLS12,               // Enforce a minimum TLS version
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
		Host:         "0.0.0.0",
		Port:         *doltPort,
		User:         *doltUser,
		Password:     finalPassword,
		DatabaseName: *doltDatabase,
	}

	tlsConfig, err := getTlsConfig(*httpCertFile, *httpKeyFile, *httpCAFile)
	if err != nil {
		logger.Fatal("failed to get TLS configuration", zap.Error(err))
	}

	if *serveHTTP {
		srv, err := pkg.NewMCPHTTPServer(
			logger,
			config,
			*mcpPort,
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
