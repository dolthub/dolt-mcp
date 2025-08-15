package pkg

import (
	"go.uber.org/zap"
)

type ZapErrorWriter struct {
	logger *zap.Logger
}

func (z *ZapErrorWriter) Write(p []byte) (n int, err error) {
	z.logger.Error(string(p))
	return len(p), nil
}

func NewZapErrorWriter(logger *zap.Logger) *ZapErrorWriter {
	return &ZapErrorWriter{
		logger: logger,
	}
}

// ZapUtilLogger adapts a zap logger to the minimal util.Logger
// interface used by github.com/mark3labs/mcp-go.
type ZapUtilLogger struct {
	sugar *zap.SugaredLogger
}

func NewZapUtilLogger(logger *zap.Logger) *ZapUtilLogger {
	return &ZapUtilLogger{sugar: logger.Sugar()}
}

func (l *ZapUtilLogger) Infof(format string, v ...any) {
	l.sugar.Infof(format, v...)
}

func (l *ZapUtilLogger) Errorf(format string, v ...any) {
	l.sugar.Errorf(format, v...)
}
