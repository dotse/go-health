// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package client

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-http-utils/headers"

	"github.com/dotse/go-health"
	"github.com/dotse/go-health/server"
)

const (
	// ErrExit is the exit code on failure.
	ErrExit = 1

	timeout = 30 * time.Second
)

// Config contains configuration for the CheckHealth() function.
type Config struct {
	// The hostname. Defaults to 127.0.0.1.
	Host string

	// The port number. Defaults to 9999.
	Port int

	// HTTP timeout. Defaults to 30 seconds.
	Timeout time.Duration
}

// CheckHealth gets a Response from an HTTP server.
func CheckHealth(config Config) (*health.Response, error) {
	if config.Host == "" {
		config.Host = "127.0.0.1"
	}

	if config.Port == 0 {
		config.Port = server.Port
	}

	if config.Timeout == 0 {
		config.Timeout = timeout
	}

	var (
		addr   = fmt.Sprintf("http://%s/", net.JoinHostPort(config.Host, strconv.Itoa(config.Port)))
		client = http.Client{
			Timeout: config.Timeout,
		}
	)

	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Add(headers.Accept, server.ContentType)

	httpResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	defer httpResp.Body.Close() // nolint: errcheck

	resp, err := health.ReadResponse(httpResp.Body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CheckHealthCommand is a utility for services that exits the current process
// with 0 or 1 for a healthy or unhealthy state, respectively.
func CheckHealthCommand() {
	resp, err := CheckHealth(Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(ErrExit)
	}

	_, _ = resp.Write(os.Stdout)

	if resp.Good() {
		os.Exit(0)
	}

	os.Exit(ErrExit)
}
