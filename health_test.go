package health_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dotse/go-health"
)

func Example() {
	ctx := context.Background()

	// Register an instance of some type that implements HealthChecker:
	m := new(MyTypeWithHealthCheck)
	health.Register(ctx, "mytype", m)

	// Register a function:
	r := health.RegisterFunc(ctx, "func", func(context.Context) (checks []health.Check) {
		// Checkers can return any number of checks.
		for i := 0; i < 3; i++ {
			var check health.Check
			// Make the relevant changes to `check` here, most importantly
			// `check.Status`.
			checks = append(checks, check)
		}

		return checks
	})
	defer r.Deregister()
}

func TestCheckerFunc_LogValue(t *testing.T) {
	t.Parallel()

	f := health.CheckerFunc(func(context.Context) []health.Check { return nil })

	assert.Regexp(t, `^func\(0x[\da-f]+\)`, f.LogValue().String())
}

func TestReadResponse(t *testing.T) {
	t.Parallel()

	r := strings.NewReader(`{ "status": "pass" }`)

	resp, err := health.ReadResponse(r)
	assert.NoError(t, err)
	require.NotNil(t, resp)

	assert.EqualValues(t, health.StatusPass, resp.Status)
}

func TestResponse_Write(t *testing.T) {
	t.Parallel()

	var (
		b    strings.Builder
		resp health.Response
	)

	_, err := resp.Write(&b)
	require.NoError(t, err)

	assert.JSONEq(t, `{"status":"pass"}`, b.String())
}

type MyTypeWithHealthCheck struct{}

func (*MyTypeWithHealthCheck) CheckHealth(context.Context) []health.Check {
	return []health.Check{{}}
}
