package logger

import (
	"go.uber.org/zap"
)

//NewLogger instantiates logger instance that should be used.
//Depending on `verbose` flag it either prints everything
//or only high priority messages (of at least warning level).
func NewLogger(verbose bool) *zap.SugaredLogger {
	var logger *zap.Logger

	if verbose {
		logger = newVerboseLogger()
	} else {
		logger = newSilentLogger()
	}

	defer logger.Sync()
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
