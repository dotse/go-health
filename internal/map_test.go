package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dotse/go-health/internal"
)

func TestInsertUnique(t *testing.T) {
	t.Parallel()

	m := make(map[string]TestingType)

	unique := internal.InsertUnique(m, "", TestingType{})
	assert.Equal(t, "TestingType", unique)
	assert.Len(t, m, 1)

	unique = internal.InsertUnique(m, "", TestingType{})
	assert.Equal(t, "TestingType-1", unique)
	assert.Len(t, m, 2)

	unique = internal.InsertUnique(m, "foo", TestingType{})
	assert.Equal(t, "foo", unique)
	assert.Len(t, m, 3)

	unique = internal.InsertUnique(m, "foo", TestingType{})
	assert.Equal(t, "foo-1", unique)
	assert.Len(t, m, 4)

	unique = internal.InsertUnique(m, "foo", TestingType{})
	assert.Equal(t, "foo-2", unique)
	assert.Len(t, m, 5)

	unique = internal.InsertUnique(m, "bar", TestingType{})
	assert.Equal(t, "bar", unique)
	assert.Len(t, m, 6)

	unique = internal.InsertUnique(m, "foo", TestingType{})
	assert.Equal(t, "foo-3", unique)
	assert.Len(t, m, 7)
}

type TestingType struct{}
