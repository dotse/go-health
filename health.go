// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	// ComponentTypeComponent is "component".
	ComponentTypeComponent = "component"
	// ComponentTypeDatastore is "datastore".
	ComponentTypeDatastore = "datastore"
	// ComponentTypeSystem is "system".
	ComponentTypeSystem = "system"

	port    = 9999
	timeout = 30 * time.Second
)

// CheckHealthConfig contains configuration for the CheckHealth() function.
type CheckHealthConfig struct {
	// The hostname. Defaults to 127.0.0.1.
	Host string

	// The port number. Defaults to 9,999.
	Port int

	// HTTP timeout. Defaults to 30 seconds.
	Timeout time.Duration
}

// CheckHealth gets a Response from an HTTP server.
func CheckHealth(config CheckHealthConfig) (*Response, error) {
	if config.Host == "" {
		config.Host = "127.0.0.1"
	}

	if config.Port == 0 {
		config.Port = port
	}

	if config.Timeout == 0 {
		config.Timeout = timeout
	}

	var (
		client = http.Client{
			Timeout: config.Timeout,
		}
		httpResp, err = client.Get(fmt.Sprintf("http://%s/", net.JoinHostPort(config.Host, strconv.Itoa(config.Port))))
	)

	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close() // nolint: errcheck

	resp, err := ReadResponse(httpResp.Body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CheckHealthCommand is a utility for services that exits the current process
// with 0 or 1 for a healthy or unhealthy state, respectively.
func CheckHealthCommand() {
	resp, err := CheckHealth(CheckHealthConfig{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}

	_, _ = resp.Write(os.Stdout)

	if resp.Good() {
		os.Exit(0)
	}

	os.Exit(1)
}

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
	s := getServer()

	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.checkers, string(r))
}

// Register registers a health checker. Can also make sure the server is
// started.
func Register(startServer bool, name string, checker Checker) Registered {
	if startServer {
		StartServer()
	}

	s := getServer()

	s.mtx.Lock()
	defer s.mtx.Unlock()

	name = insertUnique(s.checkers, name, checker)

	return Registered(name)
}

// RegisterFunc registers a health check function. Can also make sure the server
// is started.
func RegisterFunc(startServer bool, name string, f func() []Check) Registered {
	return Register(startServer, name, &checkFuncWrapper{
		Func: f,
	})
}

type checkFuncWrapper struct {
	Func func() []Check
}

func (wrapper *checkFuncWrapper) CheckHealth() []Check {
	return wrapper.Func()
}
