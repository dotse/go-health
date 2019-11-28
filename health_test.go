// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MyTypeWithHealthCheck struct{}

func (*MyTypeWithHealthCheck) CheckHealth() []Check {
	return []Check{{}}
}

func Example() {
	// Register an instance of some type that implements HealthChecker:
	m := new(MyTypeWithHealthCheck)
	Register("mytype", m)

	// Register a function:
	RegisterFunc("func", func() (checks []Check) {
		// Checkers can return any number of checks.
		for i := 0; i < 3; i++ {
			var check Check
			// Make the relevant changes to `check` here, most importantly
			// `check.Status`.
			checks = append(checks, check)
		}
		return
	})
}

func TestInsertUnique(t *testing.T) {
	m := make(map[string]Checker)

	unique := insertUnique(m, "foo", nil)
	assert.Equal(t, "foo", unique)
	assert.Equal(t, 1, len(m))

	unique = insertUnique(m, "foo", nil)
	assert.Equal(t, "foo-1", unique)
	assert.Equal(t, 2, len(m))

	unique = insertUnique(m, "foo", nil)
	assert.Equal(t, "foo-2", unique)
	assert.Equal(t, 3, len(m))

	unique = insertUnique(m, "bar", nil)
	assert.Equal(t, "bar", unique)
	assert.Equal(t, 4, len(m))

	unique = insertUnique(m, "foo", nil)
	assert.Equal(t, "foo-3", unique)
	assert.Equal(t, 5, len(m))
}

func TestReadResponse(t *testing.T) {
	r := strings.NewReader(`{ "status": "pass" }`)

	resp, err := ReadResponse(r)
	assert.NoError(t, err)
	require.NotNil(t, resp)

	assert.EqualValues(t, StatusPass, resp.Status)
}

func TestResponse_Write(t *testing.T) {
	var (
		b    strings.Builder
		resp Response
	)

	_, err := resp.Write(&b)
	require.NoError(t, err)

	assert.EqualValues(t, `{"status":"pass"}`, b.String())
}
