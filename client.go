package health

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-http-utils/headers"
	sauté "gitlab.com/biffen/saute"
)

var _ Option = optionFunc(nil)

// CheckHealth gets a Response from an HTTP server.
func CheckHealth(ctx context.Context, options ...Option) (*Response, error) {
	ctx, span := sauté.TraceFunc(ctx, nil)
	defer span.End()

	c := config{
		Host:    "localhost",
		Port:    port(),
		Timeout: timeout,
	}

	for _, option := range options {
		if err := option.apply(&c); err != nil {
			return nil, fmt.Errorf("invalid option: %w", err)
		}
	}

	var (
		addr = fmt.Sprintf(
			"http://%s/",
			net.JoinHostPort(c.Host, strconv.FormatUint(uint64(c.Port), 10)),
		)
		client = http.Client{
			Timeout: c.Timeout,
		}
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Add(headers.Accept, ContentType)

	httpResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	defer httpResp.Body.Close()

	resp, err := ReadResponse(httpResp.Body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

const (
	// ExitErr is the exit code on failure.
	ExitErr = 1
	// ExitUser is the exit code when the user did something wrong.
	ExitUser = 2

	timeout = 30 * time.Second
)

// Main is a utility for services that exits the current process with 0 or 1 for
// a healthy or unhealthy state, respectively.
func Main(ctx context.Context) {
	var exit int
	defer func() {
		//nolint:revive // Exiting here is intentional.
		os.Exit(exit)
	}()

	ctx, span := sauté.TraceFunc(ctx, nil)
	defer span.End()

	resp, err := CheckHealth(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error",
			slog.Any("error", err),
		)

		exit = ExitErr

		return
	}

	_, _ = resp.Write(os.Stdout)

	if !resp.Good() {
		exit = ExitErr
	}
}

// Option is an optional configuration for [CheckHealth].
type Option interface {
	apply(*config) error
}

// WithHost is an [Option] for [CheckHealth] to specify the host.
func WithHost(host string) Option {
	return optionFunc(func(c *config) error {
		c.Host = host
		return nil
	})
}

// WithPort is an [Option] for [CheckHealth] to specify the port number.
func WithPort(port uint16) Option {
	return optionFunc(func(c *config) error {
		c.Port = port
		return nil
	})
}

// WithTimeout is an [Option] for [CheckHealth] to specify a timeout. See
// [net/http.Client.Timeout].
func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(c *config) error {
		c.Timeout = timeout
		return nil
	})
}

type config struct {
	Host    string
	Port    uint16
	Timeout time.Duration
}

type optionFunc func(*config) error

func (f optionFunc) apply(c *config) error {
	return f(c)
}
