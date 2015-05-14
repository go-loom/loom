package worker

import (
	"github.com/seanpont/assert"
	"testing"
)

func TestParseExprs(t *testing.T) {
	a := assert.Assert(t)
	es, err := parseExprs([]string{"hello"})
	a.Nil(err)
	a.Equal(len(es), 1)
	a.Equal(es[0].key, "hello")
	a.Equal(es[0].operator, 0)
	a.Equal(es[0].value, TASK_STATE_DONE)

	es, err = parseExprs([]string{""})
	a.Nil(err)
	a.Equal(len(es), 1)
	a.Equal(es[0].key, "JOB")
	a.Equal(es[0].operator, 0)
	a.Equal(es[0].value, "START")

	es, err = parseExprs([]string{"hello==DONE"})
	a.Nil(err)
	a.Equal(len(es), 1)
	a.Equal(es[0].operator, 0)
	a.Equal(es[0].key, "hello")
	a.Equal(es[0].value, "DONE")

	es, err = parseExprs([]string{"hello ==DONE"})
	a.Nil(err)
	a.Equal(es[0].key, "hello")
}
