package config

import (
	"fmt"
	"os"
)

type HTTP struct {
	URL    string
	Method string
	Data   map[string]string
	Files  []*HTTPFile
}

type HTTPFile struct {
	Filename string
	Path     string
}

func (f *HTTPFile) Err() error {
	if f.Filename == "" || f.Path == "" {
		return fmt.Errorf("http file config is wrong")
	}

	fi, err := os.Stat(f.Path)
	if err == nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("http file is directory.")
	}

	return nil
}
