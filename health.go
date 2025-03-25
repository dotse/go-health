package health

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	sauté "gitlab.com/biffen/saute"

	"github.com/dotse/go-health/internal"
)

var (
	_ Checker        = CheckerFunc(nil)
	_ slog.LogValuer = CheckerFunc(nil)
)

// CheckNow returns the current (local) health status accumulated from all
// registered health checkers.
func CheckNow(ctx context.Context) (resp Response, err error) {
	ctx, span := sauté.TraceFunc(ctx, nil)
	defer span.End()

	checkersMu.RLock()
	defer checkersMu.RUnlock()

	span.AddEvent("lock")

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	resp.Checks = make(map[string][]Check, len(checkers))

	for name, checker := range checkers {
		if checker == nil {
			// Was deregistered
			continue
		}

		select {
		case <-ctx.Done():
			err = ctx.Err()
			return resp, err

		default:
			wg.Add(1)

			go func() {
				defer wg.Done()

				checks := checkOne(ctx, checker)

				mu.Lock()
				defer mu.Unlock()

				resp.AddChecks(name, checks...)
			}()
		}
	}

	wg.Wait()

	return resp, nil
}

func checkOne(ctx context.Context, checker Checker) (checks []Check) {
	ctx, span := sauté.TraceFunc(ctx, nil)
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			checks = []Check{
				{
					Status: StatusFail,
					Output: fmt.Sprintf("panic: %v", r),
				},
			}
		}
	}()

	return checker.CheckHealth(ctx)
}

const (
	// ComponentTypeComponent is ‘component’.
	ComponentTypeComponent = "component"
	// ComponentTypeDatastore is ‘datastore’.
	ComponentTypeDatastore = "datastore"
	// ComponentTypeSystem is ‘system’.
	ComponentTypeSystem = "system"
)

var (
	checkers     map[string]Checker
	checkersMu   sync.RWMutex
	logSubsystem = slog.String("subsystem", "health")
)

// DeregisterAll removes all previously registered health checkers.
func DeregisterAll() {
	checkersMu.Lock()
	defer checkersMu.Unlock()

	checkers = nil
}

// Checker can be implemented by anything whose health can be checked.
type Checker interface {
	CheckHealth(ctx context.Context) (checks []Check)
}

// CheckerFunc is a wrapper for a function that implements [Checker].
type CheckerFunc func(context.Context) []Check

// CheckHealth implements [Checker] by calling the [CheckerFunc].
func (f CheckerFunc) CheckHealth(ctx context.Context) []Check {
	return f(ctx)
}

func (f CheckerFunc) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("func(%v)", f))
}

// Registered is returned when registering a health check. It can be used to
// deregister that particular check at a later time, e.g. when closing whatever
// is being checked.
type Registered struct {
	string
}

// Register registers a health checker.
func Register(ctx context.Context, name string, checker Checker) Registered {
	checkersMu.Lock()
	defer checkersMu.Unlock()

	if checkers == nil {
		checkers = make(map[string]Checker)
	}

	name = internal.InsertUnique(checkers, name, checker)

	slog.DebugContext(ctx, "registered health checker",
		logSubsystem,
		slog.Any("name", name),
		slog.Any("checker", checker),
	)

	return Registered{name}
}

// RegisterFunc registers a health check function.
func RegisterFunc(
	ctx context.Context,
	name string,
	f func(context.Context) []Check,
) Registered {
	return Register(ctx, name, CheckerFunc(f))
}

// Deregister removes a previously registered health checker.
func (r Registered) Deregister() {
	checkersMu.Lock()
	defer checkersMu.Unlock()

	checkers[r.string] = nil
}
