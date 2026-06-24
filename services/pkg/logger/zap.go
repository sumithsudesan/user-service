package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger struct is a wrapper around zap.SugaredLogger
// that implements the Logger interface.
type zapLogger struct {
	log *zap.SugaredLogger
}

// New creates a JSON structured logger at the given level.
// Valid levels: debug, info, warn, error.
func New(level string) (Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", level, err)
	}

	//
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = "msg"

	//
	z, err := cfg.Build(zap.WithCaller(false))
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}

	return &zapLogger{log: z.Sugar()}, nil
}

// Debug logs a debug level message with optional key-value pairs.
func (l *zapLogger) Debug(msg string, keysAndValues ...any) {
	l.log.Debugw(msg, keysAndValues...)
}

// Info logs an info level message with optional key-value pairs.
func (l *zapLogger) Info(msg string, keysAndValues ...any) {
	l.log.Infow(msg, keysAndValues...)
}

// Warn logs a warning level message with optional key-value pairs.
func (l *zapLogger) Warn(msg string, keysAndValues ...any) {
	l.log.Warnw(msg, keysAndValues...)
}

// Error logs an error level message with optional key-value pairs.
func (l *zapLogger) Error(msg string, keysAndValues ...any) {
	l.log.Errorw(msg, keysAndValues...)
}

// With returns a child logger with the given key-value fields pre-attached.
func (l *zapLogger) With(keysAndValues ...any) Logger {
	return &zapLogger{log: l.sugar.With(keysAndValues...)}
}
