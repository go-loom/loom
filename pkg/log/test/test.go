package log

import (
	"testing"

	"github.com/go-kit/kit/log"
)

type writer struct {
	t *testing.T
}

func (w writer) Write(p []byte) (n int, err error) {
	w.t.Logf("%s", p)
	return len(p), nil
}

func TestLogger(t *testing.T) log.Logger {
	w := &writer{t}
	l := log.NewLogfmtLogger(w)
	l = log.With(l, "caller", log.DefaultCaller)

	return l
}
