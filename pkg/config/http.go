package config

import (
	"fmt"
	"os"
)

type HTTP struct {
	URL    string            `json:"url"`
	Method string            `json:"method"`
	Data   map[string]string `json:"data,omitempty"`
	Files  []*HTTPFile       `json:"files,omitempty"`
}

type HTTPFile struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

func (f *HTTPFile) Err() error {
	if f.Filename == "" || f.Path == "" {
		return fmt.Errorf("http file config is wrong")
	}

	fi, err := os.Stat(f.Path)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("http file is directory.")
	}

	return nil
}
