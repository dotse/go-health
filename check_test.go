// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	var check Check

	assert.True(t, check.Good())
	check.Status = StatusWarn
	assert.True(t, check.Good())
	check.Status = StatusFail
	assert.False(t, check.Good())

	check.SetObservedTime(123*time.Microsecond + 456*time.Nanosecond)

	check.AffectedEndpoints = []string{"https://example.test/1", "https://example.test/2"}
	check.Output = "test output"
	check.Links = []string{"https://example.test/about"}

	j, err := json.Marshal(check)
	assert.NoError(t, err)
	assert.JSONEq(t, `
{
  "affectedEndpoints": [ "https://example.test/1", "https://example.test/2" ],
  "links": [ "https://example.test/about" ],
  "observedUnit": "ns",
  "observedValue": 123456,
  "output": "test output",
  "status": "fail"
}
`, string(j))
}
