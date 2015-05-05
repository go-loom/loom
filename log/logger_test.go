package log

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	log := New("test")
	log.SetLevel(INFO)

	if log.IsDebug() {
		t.Error("debug log can't logging..")
	}

}

func TestLoggerWithEnv(t *testing.T) {
	os.Setenv("LOOM_LOG_LEVEL", "DEBUG")
	log := New("test")
	if !log.IsDebug() {
		t.Error("debug log should be logging.")
	}
}
