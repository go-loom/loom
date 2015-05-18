package worker

import (
	"os"
)

func init() {
	if os.Getenv("LOOM_LOG_LEVEL") == "" {
		os.Setenv("LOOM_LOG_LEVEL", "ERROR")
	}
}
