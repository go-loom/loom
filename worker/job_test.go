package worker

import (
	"testing"
)

func TestNewJob(t *testing.T) {
	job := NewJob("id", map[string]interface{}{"URL": "http://example.com"})

	if url, ok := job["URL"]; !ok {
		t.Error("job has no URL!")
	} else if url != "http://example.com" {
		t.Errorf("job's URL is wrong. %v", url)
	}

}
