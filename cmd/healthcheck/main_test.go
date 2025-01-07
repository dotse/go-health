package main_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dotse/go-health"
	main "github.com/dotse/go-health/cmd/healthcheck"
)

func TestMain(t *testing.T) {
	type T struct {
		Args           []string
		Check          *health.Check
		Env            map[string]string
		Exit           int
		Stderr, Stdout string
	}

	f := func(test T) {
		t.Helper()

		t.Run(fmt.Sprint(test.Args), func(t *testing.T) {
			t.Helper()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			for k, v := range test.Env {
				t.Setenv(k, v)
			}

			if test.Check != nil {
				r := health.RegisterFunc(ctx, "", func(context.Context) []health.Check {
					return []health.Check{*test.Check}
				})
				defer r.Deregister()
				health.StartServer(ctx)
			}

			var (
				dir       = t.TempDir()
				stderr, _ = os.CreateTemp(dir, "")
				stdout, _ = os.CreateTemp(dir, "")
			)

			os.Args = append([]string{"healthcheck"}, test.Args...)
			os.Stderr = stderr
			os.Stdout = stdout

			exit := main.Main(ctx)

			assert.Equal(t, test.Exit, exit, "the expected exit code")

			for _, file := range [...]struct {
				string
				*os.File
			}{
				{test.Stderr, stderr},
				{test.Stdout, stdout},
			} {
				if file.string != "" {
					file.Seek(0, 0)
					str, err := io.ReadAll(file)
					assert.NoError(t, err, "FIXME")
					assert.Regexp(t, file.string, string(str), "FIXME")
				}
			}

			if t.Failed() {
				stderr.Seek(0, 0)
				str, err := io.ReadAll(stderr)
				assert.NoError(t, err, "FIXME")
				t.Logf("STDERR:\n%s", string(str))
			}
		})
	}

	f(T{
		Args:   []string{"--version"},
		Stdout: `^healthcheck`,
	})

	f(T{
		Args:   []string{"--help"},
		Stdout: `^NAME`,
	})

	f(T{
		Args: []string{"--nope"},
		Exit: 2,
	})

	f(T{
		Check: &health.Check{
			Status: health.StatusPass,
		},
	})

	f(T{
		Args: []string{"--short"},
		Check: &health.Check{
			Status: health.StatusPass,
		},
	})

	f(T{
		Args: []string{"--port", "1234"},
		Check: &health.Check{
			Status: health.StatusPass,
		},
		Env: map[string]string{
			health.EnvHealthPort: "1234",
		},
	})

	f(T{
		Args:   []string{"too", "many", "operands"},
		Exit:   2,
		Stderr: `too many operands`,
	})
}
