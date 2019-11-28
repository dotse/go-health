// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"fmt"
	"sync"
)

const (
	// ComponentTypeComponent is "component".
	ComponentTypeComponent = "component"
	// ComponentTypeDatastore is "datastore".
	ComponentTypeDatastore = "datastore"
	// ComponentTypeSystem is "system".
	ComponentTypeSystem = "system"
)

//nolint: gochecknoglobals
var (
	checkers    map[string]Checker
	checkersMtx sync.RWMutex
)

// Checker can be implemented by anything whose health can be checked.
type Checker interface {
	CheckHealth() []Check
}

// Registered is returned when registering a health check. It can be used to
// deregister that particular check at a later time, e.g. when closing whatever
// is being checked.
type Registered string

// Deregister removes a previously registered health checker.
func (r Registered) Deregister() {
	checkersMtx.Lock()
	defer checkersMtx.Unlock()

	delete(checkers, string(r))
}

// CheckHealth returns the current (local) health status accumulated from all
// registered health checkers.
func CheckHealth() (resp Response) {
	checkersMtx.RLock()
	defer checkersMtx.RUnlock()

	for name, checker := range checkers {
		resp.AddChecks(name, getChecks(checker)...)
	}

	return
}

// Register registers a health checker.
func Register(name string, checker Checker) Registered {
	checkersMtx.Lock()
	defer checkersMtx.Unlock()

	if checkers == nil {
		checkers = make(map[string]Checker)
	}

	name = insertUnique(checkers, name, checker)

	return Registered(name)
}

// RegisterFunc registers a health check function.
func RegisterFunc(name string, f func() []Check) Registered {
	return Register(name, &checkFuncWrapper{
		Func: f,
	})
}

type checkFuncWrapper struct {
	Func func() []Check
}

func (wrapper *checkFuncWrapper) CheckHealth() []Check {
	return wrapper.Func()
}

func getChecks(checker Checker) (checks []Check) {
	defer func() {
		if r := recover(); r != nil {
			checks = []Check{
				{
					Status: StatusFail,
					Output: fmt.Sprintf("%v", r),
				},
			}
		}
	}()

	checks = checker.CheckHealth()

	return
}

func insertUnique(m map[string]Checker, name string, checker Checker) string {
	var (
		inc    uint64
		unique = name
	)

	for {
		if _, ok := m[unique]; !ok {
			break
		}

		inc++

		unique = fmt.Sprintf("%s-%d", name, inc)
	}

	m[unique] = checker

	return unique
}
