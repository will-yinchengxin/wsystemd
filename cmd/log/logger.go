package log

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/promlog"
	"io"
	"os"
	"time"
)

var (
	Logger log.Logger
)

func InitLog() {
	promLogConfig := &promlog.Config{
		Level:  &promlog.AllowedLevel{},
		Format: &promlog.AllowedFormat{},
	}
	Logger = newLogger(promLogConfig.Level.String(), promLogConfig.Format.String(), os.Stderr)
}

func newLogger(logLevel, logFormat string, writer io.Writer) log.Logger {
	var (
		l  log.Logger
		le level.Option
	)
	if logFormat == "logfmt" {
		l = log.NewLogfmtLogger(writer)
	} else {
		l = log.NewJSONLogger(writer)
	}
	switch logLevel {
	case "debug":
		le = level.AllowDebug()
	case "info":
		le = level.AllowInfo()
	case "warn":
		le = level.AllowWarn()
	case "error":
		le = level.AllowError()
	default:
		le = level.AllowInfo()
	}
	l = level.NewFilter(l, le)

	l = log.With(l,
		"ts", log.TimestampFormat(func() time.Time { return time.Now().Local() }, "2006-01-02 15:04:05"),
		"caller", log.DefaultCaller)
	return l
}
