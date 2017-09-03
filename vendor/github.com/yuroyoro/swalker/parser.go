package swalker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reProperty = regexp.MustCompile(`^(\w+)$`)
var reIndexing = regexp.MustCompile(`^(\w+)\[(\d+)\]$`)

// MustParse : panic if given expressions could not be parsed
func MustParse(exp string) Expressions {
	exps, err := Parse(exp)
	if err != nil {
		panic(err)
	}
	return exps
}

// Parse : parse given expression
func Parse(exp string) (Expressions, error) {

	tokens := strings.Split(exp, ".")

	var exps Expressions
	for _, t := range tokens {
		xs := reIndexing.FindStringSubmatch(t)
		if len(xs) > 1 {
			name := xs[1]
			ind, err := strconv.Atoi(xs[2])
			if err != nil {
				return nil, fmt.Errorf(`Invalid expression "%s" : %s`, t, err)
			}
			exps = append(exps, &Expression{Type: Indexing, Name: name, Index: ind})
			continue
		}

		ys := reProperty.FindStringSubmatch(t)
		if len(ys) > 0 {
			exps = append(exps, &Expression{Type: Property, Name: ys[1]})
			continue
		}

		return nil, fmt.Errorf(`Invalid expression "%s"`, t)
	}

	return exps, nil
}
