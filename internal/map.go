package internal

import (
	"fmt"
	"math/big"
	"reflect"
)

// InsertUnique inserts an element into a map making sure it gets a unique name
// by appending -<NUMBER> as necessary.
//
// It is not safe for concurrent use of the map; the caller is responsible for
// locking.
func InsertUnique[T any](m map[string]T, name string, checker T) string {
	if name == "" {
		name = reflect.TypeOf(checker).Name()
	}

	var (
		inc    = big.NewInt(0)
		unique = name
	)

	for {
		if _, ok := m[unique]; !ok {
			break
		}

		inc.Add(inc, big.NewInt(1))

		unique = fmt.Sprintf("%s-%s", name, inc)
	}

	m[unique] = checker

	return unique
}
