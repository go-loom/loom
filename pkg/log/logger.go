package log

import (
	stdlog "log"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

var Logger log.Logger

const loomLogLevel = "LOOM_LOG_LEVEL"

func init() {
	Logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	Logger = log.With(Logger, "ts", log.DefaultTimestampUTC)

	logLevel := level.AllowInfo()
	if os.Getenv(loomLogLevel) == "DEBUG" {
		logLevel = level.AllowAll()
	}

	Logger = level.NewFilter(Logger, logLevel)
	Logger = log.With(Logger, "caller", log.DefaultCaller)

	stdlog.SetFlags(0) // flags are handled by Go kit's logger
	stdlog.SetOutput(log.NewStdlibAdapter(Logger))
}

func With(l log.Logger, kv ...interface{}) log.Logger {
	return log.With(l, kv...)
}

func Info(l log.Logger) log.Logger {
	return level.Info(l)
}
func Debug(l log.Logger) log.Logger {
	return level.Debug(l)
}
func Error(l log.Logger) log.Logger {
	return level.Error(l)
}
