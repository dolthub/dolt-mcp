package pkg

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type stdioServerImpl struct {
	mcp         *server.MCPServer
	stdioServer *server.StdioServer
	db          db.Database
}

type StdioServer interface {
	Server
	ServeStdio(ctx context.Context)
}

var _ StdioServer = &stdioServerImpl{}

func NewMCPStdioServer(logger *zap.Logger, config db.Config, opts ...Option) (StdioServer, error) {
	db, err := db.NewDatabase(config)
	if err != nil {
		return nil, err
	}

	errorWriter := NewZapErrorWriter(logger)
	errorLogger := log.New(errorWriter, "Dolt MCP server error:", 0)

	mcp := server.NewMCPServer(
		DoltMCPServerName,
		DoltMCPServerVersion,
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	stdioServer := server.NewStdioServer(mcp)
	stdioServer.SetErrorLogger(errorLogger)

	srv := &stdioServerImpl{
		mcp:         mcp,
		db:          db,
		stdioServer: stdioServer,
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv, nil
}

func (s *stdioServerImpl) DB() db.Database {
	return s.db
}

func (s *stdioServerImpl) MCP() *server.MCPServer {
	return s.mcp
}

func (s *stdioServerImpl) ServeStdio(ctx context.Context) {
	serveStdio(ctx, s.stdioServer)
}

func serveStdio(ctx context.Context, srv *server.StdioServer) {
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	defer func() {
		stdoutWriter.Close()
		stdinWriter.Close()
		stdinReader.Close()
		stdoutReader.Close()
	}()

	// Start the server
	fmt.Println("Serving Dolt MCP on Stdin")
	if err := srv.Listen(ctx, stdinReader, stdoutWriter); err != nil && err != io.EOF && err != context.Canceled {
		fmt.Println("error serving Dolt MCP server:", err.Error())
	}

	fmt.Println("Successfully stopped Dolt MCP server.")
}
