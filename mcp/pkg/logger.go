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

