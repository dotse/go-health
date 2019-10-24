// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
