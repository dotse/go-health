package health_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dotse/go-health"
)

func TestStatus_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "pass", health.StatusPass.String())
	assert.Equal(t, "warn", health.StatusWarn.String())
	assert.Equal(t, "fail", health.StatusFail.String())
}

func TestWorstStatus(t *testing.T) {
	t.Parallel()

	assert.Equal(t, health.StatusPass, health.WorstStatus(health.StatusPass))
	assert.Equal(t, health.StatusWarn, health.WorstStatus(health.StatusWarn))
	assert.Equal(t, health.StatusFail, health.WorstStatus(health.StatusFail))

	assert.Equal(
		t,
		health.StatusWarn,
		health.WorstStatus(health.StatusPass, health.StatusWarn),
	)

	assert.Equal(
		t,
		health.StatusFail,
		health.WorstStatus(health.StatusWarn, health.StatusFail),
	)

	assert.Equal(
		t,
		health.StatusFail,
		health.WorstStatus(health.StatusPass, health.StatusFail),
	)

	assert.Equal(
		t,
		health.StatusWarn,
		health.WorstStatus(health.StatusWarn, health.StatusWarn),
	)

	assert.Equal(
		t,
		health.StatusFail,
		health.WorstStatus(health.StatusFail, health.StatusWarn),
	)

	assert.Equal(
		t,
		health.StatusFail,
		health.WorstStatus(
			health.StatusPass,
			health.StatusFail,
			health.StatusPass,
		),
	)
}
