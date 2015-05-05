package log

import (
	"github.com/koding/logging"
	"os"
	"strings"
)

const (
	FATAL Level = iota
	ERROR
	WARNING
	INFO
	DEBUG
)

var (
	StdoutHandler = NewSyncWriterHandler(os.Stdout)
	StderrHandler = NewSyncWriterHandler(os.Stderr)
)

type Level int

type Logger interface {
	Fatal(format string, args ...interface{})
	Error(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Info(format string, args ...interface{})
	Debug(format string, args ...interface{})

	IsError() bool
	IsWarning() bool
	IsInfo() bool
	IsDebug() bool

	SetLevel(level Level)
	SetHandler(logging.Handler)
}

func getLogLevel() Level {
	switch strings.ToUpper(os.Getenv("LOOM_LOG_LEVEL")) {
	case "DEBUG":
		return DEBUG
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// convertLevel converst a level into logging level
func convertLevel(l Level) logging.Level {
	switch l {
	case DEBUG:
		return logging.DEBUG
	case WARNING:
		return logging.WARNING
	case ERROR:
		return logging.ERROR
	case FATAL:
		return logging.CRITICAL
	default:
		return logging.INFO
	}
}

type logger struct {
	logging.Logger
	Level Level
}

func New(name string) Logger {
	level := getLogLevel()
	l := &logger{
		logging.NewLogger(name),
		level,
	}
	l.SetLevel(level)
	logging.StderrHandler.Level = convertLevel(getLogLevel())
	logging.StdoutHandler.Level = convertLevel(getLogLevel())

	if os.Getenv("LOOM_LOG_NOCOLOR") != "" {
		logging.StdoutHandler.Colorize = false
		logging.StderrHandler.Colorize = false
	}

	return l
}

func NewWithSync(name string) Logger {
	l := New(name)
	l.SetHandler(StderrHandler)
	StderrHandler.Level = convertLevel(getLogLevel())

	if os.Getenv("LOOM_LOG_NOCOLOR") != "" {
		StdoutHandler.Colorize = false
		StderrHandler.Colorize = false
	}

	return l
}

func (l *logger) SetLevel(level Level) {
	l.Level = level
	l.Logger.SetLevel(convertLevel(level))
}

func (l *logger) isLevel(level Level) bool {
	if l.Level >= level {
		return true
	}
	return false
}

func (l *logger) IsError() bool {
	return l.isLevel(ERROR)
}

func (l *logger) IsWarning() bool {
	return l.isLevel(WARNING)
}

func (l *logger) IsInfo() bool {
	return l.isLevel(INFO)
}

func (l *logger) IsDebug() bool {
	return l.isLevel(DEBUG)
}
