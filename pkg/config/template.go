package config

import (
	"bytes"
	"text/template"
)

type templateReader struct {
}

// Read is to execute Template context for each variables
func (tr *templateReader) Read(val string, ctx interface{}) (string, error) {
	t := template.New("")
	t, err := t.Parse(val)
	if err != nil {
		return val, err
	}

	var ret bytes.Buffer

	err = t.Execute(&ret, ctx)
	if err != nil {
		return val, err
	}

	return ret.String(), nil
}
