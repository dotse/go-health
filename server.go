// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// nolint: gochecknoglobals
var (
	initOnce sync.Once
	s        server
)

// StartServer starts an HTTP server at 0.0.0.0:9999 serving health checks. Can
// be called multiple times but will only start one server.
//
// Doesn’t usually need to be called explicitly, as Register and RegisterFunc
// take care of that.
func StartServer() {
	getServer().once.Do(func() {
		go func() {
			_ = getServer().httpServer.ListenAndServe()
		}()
	})
}

type server struct {
	checkers   map[string]Checker
	httpServer http.Server
	mtx        sync.RWMutex
	once       sync.Once
}

func (s *server) handle(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	var resp Response

	s.mtx.RLock()
	defer s.mtx.RUnlock()

	for name, checker := range s.checkers {
		checks := checker.CheckHealth()
		resp.AddChecks(name, checks...)
	}

	if req.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/health+json")

		if _, err := resp.Write(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func getServer() *server {
	initOnce.Do(func() {
		s.checkers = make(map[string]Checker)

		s.httpServer = http.Server{
			Addr:           net.JoinHostPort("0.0.0.0", strconv.Itoa(port)),
			Handler:        http.HandlerFunc(s.handle),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
	})

	return &s
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
