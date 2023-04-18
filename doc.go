// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

/*
Package health contains health checking utilities.

The most interesting part of the API is Register (and RegisterFunc), which is a
simple way to set up a health check for ‘anything’.

As soon as any health checks are registered a summary of them is served at
http://0.0.0.0:9999.

For services there is `HealthCheckCommand()` to put (early) in `main()`, e.g:

	if len(os.Args) >= 2 && os.Args[1] == "healthcheck" {
	    client.CheckHealthCommand()
	}

Docker images can then use the following:

	HEALTHCHECK --interval=10s --timeout=30s CMD ./app healthcheck

See <https://inadarei.github.io/rfc-healthcheck/>.
*/
package health
