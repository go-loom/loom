// +build ignore

package log

import (
	"errors"
	"testing"
)

func TestInfoWithSlack(t *testing.T) {
	l := WithSlack(Logger)
	l = With(l, "env", "test")

	Info(l, "module", "dummy", "status", "running")
}

func TestInfoWithoutSlack(t *testing.T) {
	l := With(Logger, "env", "test")

	Info(l, "module", "dummy", "status", "running")
}

func TestWarnWithSlack(t *testing.T) {
	l := WithSlack(Logger)
	l = With(l, "env", "test")

	Warn(l, "module", "dummy", "status", "stopped")
}

func TestErrorWithSlack(t *testing.T) {
	l := WithSlack(Logger)
	l = With(l, "env", "test")

	Error(l, "module", "dummy", "status", "panic")
}

func TestErrorWithSentry(t *testing.T) {
	l := WithSentry(Logger)
	l = With(l, "env", "test")

	Error(l, "module", "dummy", "status", "panic", "err", errors.New("test error"))
}
