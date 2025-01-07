package health_test

import (
	"context"
	"fmt"
	"time"

	"github.com/dotse/go-health"
)

func ExampleCheckHealth() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The server can be started before registering…
	if err := health.StartServer(ctx); err != nil {
		panic(err)
	}

	// Set up a checker so that there’s something to report.
	r := health.RegisterFunc(ctx, "example", func(context.Context) []health.Check {
		return []health.Check{{
			Status: health.StatusPass,
			Output: "all good",
		}}
	})
	defer r.Deregister()

	// …or after. (Subsequent StartServer()s do nothing.)
	if err := health.StartServer(ctx); err != nil {
		panic(err)
	}

	// Get the current health status of a server running at localhost. More
	// configuration is possible.
	resp, err := health.CheckHealth(ctx, health.WithTimeout(time.Minute))
	if resp == nil || err != nil {
		panic(err)
	}

	fmt.Printf(
		`
resp.Status: %q
resp.Checks["example"][0].Status: %q
resp.Checks["example"][0].Output: %q
err: %v
`,
		resp.Status,
		resp.Checks["example"][0].Status,
		resp.Checks["example"][0].Output,
		err,
	)

	// Output:
	// resp.Status: "pass"
	// resp.Checks["example"][0].Status: "pass"
	// resp.Checks["example"][0].Output: "all good"
	// err: <nil>
}
