# ⚕ `go-health`

[![License](https://img.shields.io/github/license/dotse/go-health)](https://opensource.org/licenses/MIT)
[![GoDoc](https://img.shields.io/badge/-Documentation-green?logo=go)](https://godoc.org/github.com/dotse/go-health)
[![Releases](https://img.shields.io/github/v/release/dotse/go-health?sort=semver)](https://github.com/dotse/go-health/releases)
[![Issues](https://img.shields.io/github/issues/dotse/go-health)](https://github.com/dotse/go-health/issues)

`go-health` is a Go library for setting up monitoring of anything within an
application. Anything that can have a health status can be registered, and then
an HTTP server can be started to serve a combined health status.

It follows the proposed standard [_Health Check Response Format for HTTP APIs_]
(but you don’t even have to know that, `go-health` takes care of that for you).

## Example

ℹ️ The code below is minimal. There’s more to `go-health`, but this is enough to
get something up and running.

### Setting up Health Checks

Let’s say your application has a database handle, and you want to monitor that
it can actually communicate with the database. First implement the `Checker`
interface:

```go
import (
    "github.com/dotse/go-health"
)

type MyApplication struct {
    conn Connection
    // …
}

func (app *MyApplication) CheckHealth(ctx context.Context) []health.Check {
    c := health.Check{}
    if err := app.conn.Ping(ctx); err != nil{
        c.Status = health.StatusFail
        c.Output = err.Error()
    }
    return []health.Check{ c }
}
```

Then whenever you create your application register it as a health check:

```go
app := NewMyApplication()
health.Register(ctx, "my-application", app)
health.StartServer(ctx)
```

Either like the above, e.g. in `main()`, or the application could even register
_itself_ on creation. You can register any number of checks. The reported health
status will be the ‘worst’ of all the registered checkers.

Then there will be an HTTP server listening on <http://127.0.0.1:9999/> and
serving a fresh health [response] on each request.

### Checking the Health

To then check the health, GET the response and look at its `status` field.

`go-health` has a function for this too:

```go
resp, err := health.CheckNow(ctx)
if err != nil {
    return err
}
fmt.Printf("Status: %s\n", resp.Status)
```

### Command-Line Tool

`go-health` provides a command-line tool for health checking. To install it run:

```sh
go install github.com/dotse/go-health/cmd/healthcheck@latest
```

For usage documentation see [`usage.txt`].

### Using as a Docker Health Check

A way to create a health check for a Docker image is to use the same binary as
your application to do the checking. E.g. if the application is invoked with the
first argument `healthcheck`:

```go
func main() {
    // …

    if os.Args[1] == "healthcheck" {
        health.Main(ctx)
    }

    // Your other code, including health.StartServer(ctx)…
}
```

`Main()` will GET the current health from the local HTTP server, parse the
response and `os.Exit()` either 0 or 1, depending on the health.

Then in your `Dockerfile` add:

```dockerfile
HEALTHCHECK --interval=10s --timeout=30s CMD ./app healthcheck
```

💁 Voilà! A bit of code and your Docker image has a built-in health check for
all the things you want monitored.

## Using as a Kubernetes Probe

If your Kubernetes pod runs a health check server, e.g. using
`health.StartServer(ctx)`, then [probing it][probe] can be as little as:

```yaml
spec:
  containers:
    - # …
      ports:
        - containerPort: 9999
          name: health
      livenessProbe:
        httpGet:
          port: health
      readinessProbe:
        httpGet:
          port: health
      startupProbe:
        httpGet:
          port: health
```

[_Health Check Response Format for HTTP APIs_]: https://inadarei.github.io/rfc-healthcheck/
[`usage.txt`]: ./cmd/healthcheck/usage.txt
[probe]: https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#Probe
[response]: https://inadarei.github.io/rfc-healthcheck/#rfc.section.3
