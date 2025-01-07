package health

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/go-http-utils/negotiator"
	sauté "gitlab.com/biffen/saute"
	"go.opentelemetry.io/otel/codes"
)

const (
	// ContentType is the media (MIME) type returned by the server.
	ContentType = "application/health+json"
	// EnvHealthPort is ‘HEALTH_PORT’; the environment variable to configure the
	// health check HTTP port.
	EnvHealthPort = "HEALTH_PORT"
)

var (
	server   *http.Server
	serverMu sync.Mutex
)

// HandleHTTP serves a health response over HTTP. It supports the GET, HEAD and
// OPTIONS methods as well as content negotiation.
func HandleHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, span := sauté.TraceFunc(req.Context(), nil)
	defer span.End()

	errorStatus := func(err error, status int) {
		if err == nil {
			http.Error(w, "", status)
			span.SetStatus(codes.Error, http.StatusText(status))

			return
		}

		http.Error(w, err.Error(), status)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	switch req.Method {
	case http.MethodOptions:
		w.Header().Set(headers.Allow, strings.Join([]string{
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
		}, ", "))
		w.WriteHeader(http.StatusNoContent)

		return

	case http.MethodGet, http.MethodHead:

	default:
		errorStatus(nil, http.StatusMethodNotAllowed)
		return
	}

	negotiator := negotiator.New(req.Header)

	if negotiator.Charset("UTF-8") == "" {
		http.Error(w, "", http.StatusNotAcceptable)
		return
	}

	ct := negotiator.Type(ContentType, "application/json")
	if ct == "" {
		errorStatus(nil, http.StatusNotAcceptable)
		return
	}

	w.Header().Set(headers.ContentEncoding, "UTF-8")
	w.Header().Set(headers.ContentType, ct)

	resp, err := CheckNow(ctx)
	if err != nil {
		errorStatus(err, http.StatusInternalServerError)
	}

	if resp.Status == StatusFail {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if req.Method == http.MethodHead {
		w.Header().Set(headers.ContentLength, "0")
		return
	}

	if _, err := resp.Write(w); err != nil {
		errorStatus(err, http.StatusInternalServerError)
	}
}

// StartServer starts an HTTP server at 0.0.0.0:${HEALTH_PORT:-9999} serving
// health checks. Can be called multiple times but will only start one server.
//
// Will block until the server is listening.
//
// The server will be stopped when the passed [context.Context] is cancelled.
func StartServer(ctx context.Context) error {
	serverMu.Lock()
	defer serverMu.Unlock()

	if server == nil {
		addr := netip.AddrPortFrom(netip.IPv4Unspecified(), port())

		listener, err := net.Listen("tcp", addr.String())
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /{$}", HandleHTTP)

		server = &http.Server{
			Addr:              addr.String(),
			Handler:           mux,
			ReadHeaderTimeout: 30 * time.Second,
		}

		go func() {
			if err := server.Serve(listener); err != nil &&
				!errors.Is(err, http.ErrServerClosed) {
				slog.ErrorContext(ctx, "health server error",
					slog.Any("error", err),
				)
			}

			serverMu.Lock()
			defer serverMu.Unlock()

			server = nil
		}()

		go func() {
			<-ctx.Done()

			if err := server.Shutdown(context.Background()); err != nil {
				slog.ErrorContext(ctx, "error when stopping health server",
					slog.Any("error", err),
				)
			}
		}()
	}

	return nil
}

func port() uint16 {
	if str := os.Getenv(EnvHealthPort); str != "" {
		u, err := strconv.ParseUint(str, 0, 16)
		if err == nil {
			return uint16(u)
		}

		slog.Error("failed to parse "+EnvHealthPort,
			logSubsystem,
			slog.Any("error", err),
		)
	}

	return 9_999
}
