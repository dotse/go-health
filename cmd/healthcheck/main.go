// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/tidwall/pretty"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/dotse/go-health"
)

const (
	errorKey = "error"
)

type cmd struct {
	config     health.CheckHealthConfig
	continuous bool
	interval   time.Duration
	isatty     bool
	print      func(*health.Response)
	short      bool
	stats      map[string]uint64
	stop       bool
}

func newCmd() cmd {
	c := cmd{}

	var (
		docker  bool
		port    int
		timeout time.Duration
	)

	flag.BoolVar(&c.continuous, "c", false, "Run continuously (stop with Ctrl+C).")
	flag.BoolVar(&docker, "d", false, "Address is the name of a Docker container.")
	flag.DurationVar(&c.interval, "n", 0, "Interval between continuous checks (implies -c) (default: 2s).")
	flag.IntVar(&port, "p", 0, "Port.")
	flag.BoolVar(&c.short, "s", false, "Short output (just the status).")
	flag.DurationVar(&timeout, "t", 0, "HTTP timeout.")

	flag.Parse()

	// Setting interval implies continuous
	c.continuous = c.continuous || c.interval != 0
	if c.continuous && c.interval == 0 {
		c.interval = 2 * time.Second
	}

	var host string

	if docker {
		var err error
		if host, err = getContainerAddress(flag.Arg(0)); err != nil {
			log.Fatal(err)
		}
	} else {
		host = flag.Arg(0)
	}

	c.config = health.CheckHealthConfig{
		Port:    port,
		Host:    host,
		Timeout: timeout,
	}

	c.isatty = terminal.IsTerminal(int(os.Stdout.Fd()))
	c.print = c.makePrint()

	return c
}

func (c *cmd) exit() {
	if c.continuous && c.isatty {
		var str []string

		for status, count := range c.stats {
			str = append(str, fmt.Sprintf("\033[%dm%d %s\033[0m", map[string]int{
				health.StatusPass.String(): 32,
				health.StatusWarn.String(): 33,
				health.StatusFail.String(): 31,
				errorKey:                   91,
			}[status], count, status))
		}

		fmt.Printf("\n---\n%s\n", strings.Join(str, ", "))
	}

	for status, count := range c.stats {
		if status == health.StatusPass.String() {
			continue
		}

		if count > 0 {
			os.Exit(1)
		}
	}

	os.Exit(0)
}

func (c *cmd) makePrint() func(*health.Response) {
	switch {
	case c.continuous && c.isatty && c.short:
		return func(resp *health.Response) {
			fmt.Printf("%s %s\r", time.Now().Format(time.RFC3339), resp.Status)
		}

	case c.short:
		return func(resp *health.Response) {
			fmt.Println(resp.Status)
		}

	case c.isatty:
		return func(resp *health.Response) {
			var buffer bytes.Buffer
			_, _ = resp.Write(&buffer)
			fmt.Println(string(pretty.Color(pretty.Pretty(buffer.Bytes()), nil)))
		}

	default:
		return func(resp *health.Response) {
			_, _ = resp.Write(os.Stdout)
		}
	}
}

func (c *cmd) run() {
	c.stats = make(map[string]uint64)

	go c.wait()

	for !c.stop {
		resp, err := health.CheckHealth(c.config)
		if err == nil {
			c.stats[resp.Status.String()]++
			c.print(resp)
		} else {
			c.stats[errorKey]++
			log.Println(err)
		}

		if !c.continuous {
			c.exit()
		}

		time.Sleep(c.interval)
	}
}

func (c *cmd) wait() {
	channel := make(chan os.Signal, 1)

	signal.Notify(channel, os.Interrupt)

	<-channel

	c.stop = true

	c.exit()
}

func getContainerAddress(container string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}

	containerJSON, err := cli.ContainerInspect(context.Background(), container)
	if err != nil {
		return "", err
	}

	for _, network := range containerJSON.NetworkSettings.Networks {
		if network.IPAddress != "" {
			return network.IPAddress, nil
		}
	}

	return "", fmt.Errorf("couldn’t find address of %q", container)
}

func main() {
	c := newCmd()
	c.run()
}
