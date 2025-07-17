package pkg

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
	"github.com/mark3labs/mcp-go/server"
)

const (
	DoltMCPServerName    = "dolt-mcp"
	DoltMCPServerVersion = "0.1.0"
)

type Server interface {
	MCP() *server.MCPServer
	DB() db.Database
	ListenAndServe(ctx context.Context)
}

type Option func(*httpServerImpl)

func WithToolSet(ts ToolSet) Option {
	return func(s *httpServerImpl) {
		ts.RegisterTools(s)
	}
}

type httpServerImpl struct {
	mcp     *server.MCPServer
	handler http.Handler
	port    int
	db      db.Database
}

var _ Server = &httpServerImpl{}

func NewMCPHTTPServer(config db.Config, port int, opts ...Option) (Server, error) {
	db, err := db.NewDatabase(config)
	if err != nil {
		return nil, err
	}

	mcp := server.NewMCPServer(
		DoltMCPServerName,
		DoltMCPServerVersion,
		server.WithToolCapabilities(true),
	)

	srv := &httpServerImpl{
		mcp:     mcp,
		db:      db,
		port: port,
		handler: server.NewStreamableHTTPServer(mcp),
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv, nil
}

func (s *httpServerImpl) DB() db.Database {
	return s.db
}

func (s *httpServerImpl) MCP() *server.MCPServer {
	return s.mcp
}

func (s *httpServerImpl) ListenAndServe(ctx context.Context) {
	serve(ctx, s.handler, s.port)
}

func serve(ctx context.Context, handler http.Handler, port int) {
	portStr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    portStr,
		Handler: handler,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	shutdownOnce := sync.Once{}

	// Graceful shutdown logic shared by both signal and context
	shutdown := func(reason string) {
		shutdownOnce.Do(func() {
			fmt.Println("Shutting down Dolt MCP due to:", reason)
			ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctxTimeout); err != nil {
				fmt.Println("failed to shutdown server:", err.Error())
			}
		})
	}

	// Listen for OS signal
	go func() {
		<-quit
		shutdown("signal")
	}()

	// Listen for context cancellation
	go func() {
		<-ctx.Done()
		shutdown("context cancellation")
	}()

	// Start the server
	fmt.Println("Serving Dolt MCP on", portStr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("error serving Dolt MCP server:", err.Error())
	}

	fmt.Println("Successfully stopped Dolt MCP server.")
}

