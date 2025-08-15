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
	mcp      *server.MCPServer
	handler  http.Handler
	port     int
	dbConfig db.Config
	logger   *zap.Logger
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

	baseHandler := server.NewStreamableHTTPServer(mcp, server.WithLogger(NewZapUtilLogger(logger)))
	// If debug logging is enabled, wrap with access log middleware
	var handler http.Handler = baseHandler
	if logger.Core().Enabled(zap.DebugLevel) {
		handler = withAccessLogging(baseHandler, logger)
	}

	srv := &httpServerImpl{
		logger:   logger,
		mcp:      mcp,
		dbConfig: config,
		port:     port,
		handler:  handler,
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

// withAccessLogging wraps an http.Handler to log HTTP requests at debug level
// including method, path, status code, and duration.
func withAccessLogging(next http.Handler, logger *zap.Logger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        lrw := &loggingResponseWriter{ResponseWriter: w}

        next.ServeHTTP(lrw, r)

        // Default to 200 if neither WriteHeader nor Write were called
        if lrw.statusCode == 0 {
            lrw.statusCode = http.StatusOK
        }

        duration := time.Since(start)

        // Access logs should go to stderr; zap production config defaults include stderr.
        logger.Debug("http request",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.Int("status", lrw.statusCode),
            zap.Int("bytes", lrw.bytesWritten),
            zap.String("remote", r.RemoteAddr),
            zap.Duration("duration", duration),
        )
    })
}

// loggingResponseWriter captures status code and bytes written for access logging
type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode   int
    bytesWritten int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
    w.statusCode = code
    w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
    if w.statusCode == 0 {
        w.statusCode = http.StatusOK
    }
    n, err := w.ResponseWriter.Write(b)
    w.bytesWritten += n
    return n, err
}

