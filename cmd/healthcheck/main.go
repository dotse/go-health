package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/dotse/slug"
	"github.com/logrusorgru/aurora"
	"github.com/tidwall/pretty"
	"gitlab.com/biffen/go-applause"
	"golang.org/x/term"

	"github.com/dotse/go-health"
)

const (
	errorKey        = "error"
	defaultInterval = 2 * time.Second
	name            = "healthcheck"
)

//go:embed usage.txt
var usage []byte

func Main(ctx context.Context) int {
	var (
		c        cmd
		help     bool
		isDocker bool
		level    = slog.LevelInfo
		parser   applause.Parser
		port     uint16
		timeout  time.Duration
		version  bool
	)

	parser.Add(
		applause.Option{
			Names:  []string{"V", "version"},
			Target: &version,
		},

		applause.Option{
			Names:  []string{"c", "continuous"},
			Target: &c.continuous,
		},

		applause.Option{
			Names:  []string{"d", "docker"},
			Target: &isDocker,
		},

		applause.Option{
			Names:  []string{"h", "?", "help"},
			Target: &help,
		},

		applause.Option{
			Names:  []string{"n", "interval"},
			Target: &c.interval,
		},

		applause.Option{
			Names:  []string{"p", "port"},
			Target: &port,
		},

		applause.Option{
			Names:  []string{"q", "quiet"},
			Target: &level,
			Add:    +4,
		},

		applause.Option{
			Names:  []string{"s", "short"},
			Target: &c.short,
		},

		applause.Option{
			Names:  []string{"t", "timeout"},
			Target: &timeout,
		},

		applause.Option{
			Names:  []string{"v", "verbose"},
			Target: &level,
			Add:    -4,
		},
	)

	operands, err := parser.Parse(ctx, os.Args[1:])
	if err != nil {
		_, _ = os.Stderr.Write(usage)

		return health.ExitUser
	}

	switch {
	case help:
		_, _ = os.Stdout.Write(usage)
		return 0

	case version:
		fmt.Print(name)

		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Print(" ", info.Main.Version)

			for _, bs := range info.Settings {
				switch bs.Key {
				case "vcs.revision", "vcs.time":
					fmt.Print(" ", bs.Value)
				}
			}

			fmt.Print(" ", info.GoVersion)
		}

		fmt.Println()

		return 0
	}

	slog.SetDefault(slog.New(slug.NewHandler(
		slug.HandlerOptions{
			HandlerOptions: slog.HandlerOptions{
				Level: level,
			},
		},
		os.Stderr,
	)))

	switch len(operands) {
	case 0:

	case 1:
		var host string

		if isDocker {
			if host, err = getContainerAddress(operands[0]); err != nil {
				return health.ExitUser
			}
		} else {
			host = operands[0]
		}

		c.options = append(c.options, health.WithHost(host))

	default:
		slog.ErrorContext(ctx, "too many operands")

		return health.ExitUser
	}

	if port != 0 {
		c.options = append(c.options, health.WithPort(port))
	}

	if timeout != 0 {
		c.options = append(c.options, health.WithTimeout(timeout))
	}

	// Setting interval implies continuous
	c.continuous = c.continuous || c.interval != 0
	if c.continuous && c.interval == 0 {
		c.interval = defaultInterval
	}

	c.isatty = term.IsTerminal(int(os.Stdout.Fd()))
	c.printFunc = c.makePrint()

	return c.run(ctx)
}

func getContainerAddress(container string) (string, error) {
	cli, err := docker.NewClientWithOpts(docker.FromEnv)
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
	var exit int
	defer func() {
		os.Exit(exit)
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	exit = Main(ctx)
}

type cmd struct {
	continuous bool
	interval   time.Duration
	isatty     bool
	options    []health.Option
	printFunc  func(*health.Response)
	short      bool
	stats      map[string]uint64
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
			fmt.Println(
				string(pretty.Color(pretty.Pretty(buffer.Bytes()), nil)),
			)
		}

	default:
		return func(resp *health.Response) {
			_, _ = resp.Write(os.Stdout)
		}
	}
}

func (c *cmd) run(ctx context.Context) int {
	c.stats = make(map[string]uint64)

	for {
		resp, err := health.CheckHealth(ctx, c.options...)
		if err == nil {
			c.stats[resp.Status.String()]++
			c.printFunc(resp)
		} else {
			c.stats[errorKey]++

			slog.ErrorContext(ctx, "error",
				slog.Any("error", err),
			)
		}

		if !c.continuous {
			for status, count := range c.stats {
				if status == health.StatusPass.String() {
					continue
				}

				if count > 0 {
					return health.ExitErr
				}
			}

			return 0
		}

		if c.isatty {
			var str []string

			for status, count := range c.stats {
				str = append(str, map[string]func(interface{}) aurora.Value{
					health.StatusPass.String(): aurora.Green,
					health.StatusWarn.String(): aurora.Yellow,
					health.StatusFail.String(): aurora.Red,
					errorKey:                   aurora.BrightRed,
				}[status](fmt.Sprintf("%d %s", count, status)).String())
			}

			fmt.Printf("\n---\n%s\n", strings.Join(str, ", "))
		}

		select {
		case <-ctx.Done():
			return 0

		case <-time.After(c.interval):
		}
	}
}
