package log

import (
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/logger/slog"
	"golang.org/x/xerrors"
)

var Logger *slog.Logger

func HertzLogInit(level string) (*slog.Logger, error) {
	var logLevel hlog.Level
	lowLevel := strings.ToLower(level)
	switch lowLevel {
	case "trace":
		logLevel = hlog.LevelTrace
	case "debug":
		logLevel = hlog.LevelDebug
	case "info":
		logLevel = hlog.LevelInfo
	case "warn":
		logLevel = hlog.LevelWarn
	case "error":
		logLevel = hlog.LevelError
	default:
		return nil, xerrors.New("unsupported log level: " + lowLevel)
	}
	logger := slog.NewLogger()
	logger.SetLevel(logLevel)
	hlog.SetLogger(logger)
	Logger = logger
	return logger, nil
}
