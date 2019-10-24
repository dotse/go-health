# ⚕ `go-health`

[![License](https://img.shields.io/github/license/dotse/go-health)](https://opensource.org/licenses/MIT)
[![GoDoc](https://img.shields.io/badge/-Documentation-green?logo=go)](https://godoc.org/github.com/dotse/go-health)
[![Actions](https://github.com/dotse/go-health/workflows/Build/badge.svg?branch=master)](https://github.com/dotse/go-health/actions)
[![Releases](https://img.shields.io/github/v/release/dotse/go-health?sort=semver)](https://github.com/dotse/go-health/releases)
[![Issues](https://img.shields.io/github/issues/dotse/go-health)](https://github.com/dotse/go-health/issues)

`go-health` is a Go library for easily setting up monitoring of anything within
an application. Anything that can have a health status can be registered, and
then, as if by magic 🧙, an HTTP server is running and serving a combined health
status.

It follows the proposed standard [_Health Check Response Format for HTTP APIs_]
(but you don’t even have to know that, `go-health` takes care of that for you).

## Example

ℹ️ The code below is minimal. There’s more to `go-health`, but this is enough to
get something up and running.

### Setting up Health Checks

Let‘s say your application has a database handle, and you want to monitor that
it can actually communicate with the database. Simply implement the `Checker`
interface:

```go
import (
    "github.com/dotse/go-health"
)

type MyApplication struct {
    db sql.DB

    // ...
}

func (app *MyApplication) CheckHealth() []health.Check {
    c := health.Check{}
    if err := app.db.Ping(); err != nil{
        c.Status = health.StatusFail
        c.Output = err.Error()
    }
    return []health.Check{ c }
}
```

Then whenever you create your application register it as a health check:

```go
app := NewMyApplication()
health.Register(true, "my-application", app)
```

Either like the above, e.g. in `main()`, or the application could even register
_itself_ on creation. You can register as many times as you want. The reported
health status will be the ‘worst’ of all the registered checkers.

Then there will be an HTTP server listening on <http://127.0.0.1:9999/> and
serving a fresh health [response] on each request.

### Checking the Health

To then check the health, GET the response and look at its `status` field.

`go-health` has a function for this too:

```go
resp, err := health.CheckHealth(c.config)
if err == nil {
    fmt.Printf("Status: %s\n", resp.Status)
} else {
    fmt.Printf("ERROR: %v\n", err)
}
```

### Using as a Docker Health Check

An easy way to create a health check for a Docker image is to use the same
binary as your application to do the checking. E.g. if the application is
invoked with the first argument `healthcheck`:

```go
func main() {
    if os.Args[1] == "healthcheck" {
        health.CheckHealthCommand()
    }

    // Your other code...
}
```

`CheckHealthCommand()` will GET the current health from the local HTTP server,
parse the response and `os.Exit()` either 0 or 1, depending on the health.

Then in your `Dockerfile` add:

```dockerfile
HEALTHCHECK --interval=10s --timeout=30s CMD ./app healthcheck
```

💁 Voilà! A few lines of code and your Docker image has a built-in health check
for all the things you want monitored.

[_health check response format for http apis_]: https://inadarei.github.io/rfc-healthcheck/
[response]: https://inadarei.github.io/rfc-healthcheck/#rfc.section.3
