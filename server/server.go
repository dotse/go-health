// Copyright © 2019, 2022 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package server

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-http-utils/headers"
	"github.com/go-http-utils/negotiator"

	"github.com/dotse/go-health"
)

const (
	// ContentType is the media (MIME) type returned by the server.
	ContentType = "application/health+json"

	// Port is the default port for checking health over HTTP.
	Port = 9_999
)

// nolint: gochecknoglobals
var (
	httpServer *http.Server
	initMtx    sync.Mutex
)

// Start starts an HTTP server at 0.0.0.0:9999 serving health checks. Can be
// called multiple times but will only start one server.
//
// Will block until the server is listening.
func Start() error {
	initMtx.Lock()
	defer initMtx.Unlock()

	if httpServer == nil {
		addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(Port))

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}

		httpServer = &http.Server{
			Addr:    addr,
			Handler: http.HandlerFunc(Handle),
		}

		go func() {
			_ = httpServer.Serve(listener)
		}()
	}

	return nil
}

// Handle serves a health response over HTTP. It supports the GET, HEAD and
// OPTIONS methods as well as content negotiation.
func Handle(w http.ResponseWriter, req *http.Request) {
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
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	negotiator := negotiator.New(req.Header)

	if negotiator.Charset("UTF-8") == "" {
		http.Error(w, "", http.StatusNotAcceptable)
		return
	}

	ct := negotiator.Type(ContentType, "application/json")
	if ct == "" {
		http.Error(w, "", http.StatusNotAcceptable)
		return
	}

	w.Header().Set(headers.ContentEncoding, "UTF-8")
	w.Header().Set(headers.ContentType, ct)

	if req.Method == http.MethodHead {
		w.Header().Set(headers.ContentLength, "0")
		return
	}

	resp, err := health.CheckHealthContext(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if _, err := resp.Write(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
