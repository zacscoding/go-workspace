// this logger is copied from https://github.com/google/exposure-notifications-server/blob/702c0796ecf6ab1a6816aaa2dfe422fed39fce9d/pkg/logging/logger.go
// and changed
package zaplogger

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

type contextKey = string

const loggerKey = contextKey("logger")

var (
	// defaultLogger is the default logger. It is initialized once per package
	// include upon calling DefaultLogger.
	defaultLogger     *zap.SugaredLogger
	defaultLoggerOnce sync.Once
)

// NewLogger creates a new logger with the given log level
func NewLogger(level zapcore.Level) *zap.SugaredLogger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg := zap.Config{
		Encoding:         "console",
		EncoderConfig:    ec,
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := cfg.Build()
	if err != nil {
		logger = zap.NewNop()
	}
	return logger.Sugar()
}

// DefaultLogger returns the default logger for the package.
func DefaultLogger() *zap.SugaredLogger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = NewLogger(zapcore.Level(-1))
	})
	return defaultLogger
}

// WithLogger creates a new context with the provided logger attached.
func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext returns the logger stored in the context. If no such logger
// exists, a default logger is returned.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return logger
	}
	fmt.Println("Not found logger in context")
	return DefaultLogger()
}

func WithGinLoggwer(ctx *gin.Context, logger *zap.SugaredLogger) {
	ctx.Set(loggerKey, logger)
}

func FromGinContext(ctx *gin.Context) *zap.SugaredLogger {
	logger, ok := ctx.Get(loggerKey)
	if !ok {
		return DefaultLogger()
	}
	return logger.(*zap.SugaredLogger)
}
