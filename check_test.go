package health_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dotse/go-health"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	var check health.Check

	assert.True(t, check.Good())
	check.Status = health.StatusWarn
	assert.True(t, check.Good())
	check.Status = health.StatusFail
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
