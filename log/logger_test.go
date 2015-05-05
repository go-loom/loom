package log

import (
	"testing"
)

func TestLogger(t *testing.T) {
	log := New("test")
	log.SetLevel(INFO)

	if log.IsDebug() {
		t.Error("debug log can't logging..")
	}
}
