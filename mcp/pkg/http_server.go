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
	"go.uber.org/zap"
)

type httpServerImpl struct {
	mcp     *server.MCPServer
	handler http.Handler
	port    int
	dbConfig db.Config
	logger  *zap.Logger
}

type HTTPServer interface {
	Server
	ListenAndServe(ctx context.Context)
}

var _ HTTPServer = &httpServerImpl{}

func NewMCPHTTPServer(logger *zap.Logger, config db.Config, port int, opts ...Option) (HTTPServer, error) {
	mcp := server.NewMCPServer(
		DoltMCPServerName,
		DoltMCPServerVersion,
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	srv := &httpServerImpl{
		logger:  logger,
		mcp:     mcp,
		dbConfig: config,
		port:    port,
		handler: server.NewStreamableHTTPServer(mcp),
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv, nil
}

func (s *httpServerImpl) MCP() *server.MCPServer {
	return s.mcp
}

func (s *httpServerImpl) DBConfig() db.Config {
	return s.dbConfig
}

func (s *httpServerImpl) ListenAndServe(ctx context.Context) {
	serve(ctx, s.logger, s.handler, s.port)
}

func serve(ctx context.Context, logger *zap.Logger, handler http.Handler, port int) {
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
				logger.Error("failed to shutdown server", zap.Error(err))
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
		logger.Error("error serving Dolt MCP server", zap.Error(err))
	}

	logger.Info("Successfully stopped Dolt MCP server.")
}
