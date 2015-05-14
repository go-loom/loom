package worker

import (
	"strings"
)

const (
	EQ = iota
	NOTEQ
)

var OPERATORS = []string{"==", "!="}

type expr struct {
	key      string
	operator int
	value    string
}

func parseExprs(env []string) ([]expr, error) {
	exprs := []expr{}
	for _, e := range env {
		found := false
		for i, op := range OPERATORS {
			if strings.Contains(e, op) {
				//split with the op
				parts := strings.SplitN(e, op, 2)
				//TODO: Validate

				if len(parts) == 2 {
					exprs = append(exprs, expr{
						key:      strings.TrimSpace(parts[0]),
						operator: i,
						value:    strings.TrimSpace(parts[1]),
					})

				} else {
					exprs = append(exprs, expr{
						key:      strings.TrimSpace(parts[0]),
						operator: i,
					})
				}
				found = true
				break
			}
		}

		if !found {

			var key, value string
			e := strings.TrimSpace(e)
			if e == "" {
				key = "JOB"
				value = "START"
			} else {
				key = e
				value = TASK_STATE_DONE
			}

			exprs = append(exprs, expr{
				key:      key,
				operator: EQ,
				value:    value,
			})

		}

	}

	return exprs, nil
}
