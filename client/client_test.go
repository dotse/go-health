// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package client

import (
	"fmt"
	"time"

	"github.com/dotse/go-health"
	"github.com/dotse/go-health/server"
)

func ExampleCheckHealth() {
	// The server can be started before registering…
	if err := server.Start(); err != nil {
		panic(err)
	}

	// Set up a checker so that there’s something to report.
	health.RegisterFunc("example", func() []health.Check {
		return []health.Check{{
			Status: health.StatusPass,
			Output: "all good",
		}}
	})

	// …or after. (Subsequent Start()s do nothing.)
	if err := server.Start(); err != nil {
		panic(err)
	}

	// Get the current health status of a server running at localhost. More
	// configuration is possible.
	resp, err := CheckHealth(Config{
		Timeout: time.Minute,
	})

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
