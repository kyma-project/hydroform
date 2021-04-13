package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Interface describes logger API.
type Interface interface {
	// Info prints info message.
	Info(args ...interface{})

	// Infof prints formatted info message.
	Infof(template string, args ...interface{})

	// Warn prints warning message.
	Warn(args ...interface{})

	// Warnf prints formatted warning message.
	Warnf(template string, args ...interface{})

	// Error prints error message.
	Error(args ...interface{})

	// Errorf prints formatted error message.
	Errorf(template string, args ...interface{})

	// Fatal prints fatal message and calls os.Exit.
	Fatal(args ...interface{})

	// Fatalf prints formatted fatal message and calls os.Exit.
	Fatalf(template string, args ...interface{})
}

// Logger default implementation of logging.Interface.
type Logger struct {
	internalLogger *zap.SugaredLogger
}

// NewLogger instantiates logger instance that should be used.
// Depending on `verbose` flag it either prints everything
// or only high priority messages (of at least warning level).
func NewLogger(verbose bool) *Logger {
	zapLogger := newInternalLogger(verbose)
	return &Logger{
		internalLogger: zapLogger,
	}
}

func newInternalLogger(verbose bool) *zap.SugaredLogger {
	var logger *zap.Logger

	if verbose {
		logger = newVerboseLogger()
	} else {
		logger = newSilentLogger()
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error(fmt.Sprintf("Sync failed: %s", err))
		}
	}()
	return logger.Sugar()
}

func newVerboseLogger() *zap.Logger {
	logger, _ := getDevelopmentConfig().Build()
	return logger
}

func newSilentLogger() *zap.Logger {
	cfg := getDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	logger, _ := cfg.Build()
	return logger
}

func getDevelopmentConfig() zap.Config {
	return zap.NewDevelopmentConfig()
}

func (l *Logger) Info(args ...interface{}) {
	l.internalLogger.Info(args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.internalLogger.Infof(template, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.internalLogger.Warn(args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.internalLogger.Warnf(template, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.internalLogger.Error(args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.internalLogger.Errorf(template, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.internalLogger.Fatal(args...)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.internalLogger.Fatalf(template, args...)
}
