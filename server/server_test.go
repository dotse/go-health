// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
)

func TestHandle(t *testing.T) { // nolint: funlen
	for _, c := range [...]struct {
		Name            string
		Method          string
		Headers         map[string]string
		ExpectedHeaders map[string]string
		ExpectedStatus  int
	}{
		{
			Name:           "bad type",
			Headers:        map[string]string{headers.Accept: "text/html"},
			ExpectedStatus: http.StatusNotAcceptable,
		},

		{
			Name:           "bad charset",
			Headers:        map[string]string{headers.AcceptCharset: "iso-8859-1"},
			ExpectedStatus: http.StatusNotAcceptable,
		},

		{
			Name: "good type, bad charset",
			Headers: map[string]string{
				headers.Accept:        ContentType,
				headers.AcceptCharset: "iso-8859-1",
			},
			ExpectedStatus: http.StatusNotAcceptable,
		},

		{
			Name: "good charset, bad type",
			Headers: map[string]string{
				headers.Accept:        "text/html",
				headers.AcceptCharset: "UTF-8",
			},
			ExpectedStatus: http.StatusNotAcceptable,
		},

		{
			Name:           http.MethodPost,
			Method:         http.MethodPost,
			ExpectedStatus: http.StatusMethodNotAllowed,
		},

		{
			Name:            http.MethodOptions,
			Method:          http.MethodOptions,
			ExpectedStatus:  http.StatusNoContent,
			ExpectedHeaders: map[string]string{headers.Allow: "OPTIONS, GET, HEAD"},
		},

		{
			Name:            http.MethodHead,
			Method:          http.MethodHead,
			ExpectedHeaders: map[string]string{headers.ContentLength: "0"},
		},

		{
			Name: "no headers",
		},

		{
			Name:            "alternative content-type good",
			Headers:         map[string]string{headers.Accept: "application/json"},
			ExpectedHeaders: map[string]string{headers.ContentType: "application/json"},
		},

		{
			Name: "good charset, good type",
			Headers: map[string]string{
				headers.Accept:        ContentType,
				headers.AcceptCharset: "UTF-8",
			},
		},
	} {
		if c.Method == "" {
			c.Method = http.MethodGet
		}

		if c.ExpectedHeaders == nil {
			c.ExpectedHeaders = make(map[string]string)
		}

		if _, ok := c.ExpectedHeaders[headers.ContentType]; !ok {
			c.ExpectedHeaders[headers.ContentType] = ContentType
		}

		if c.ExpectedStatus == 0 {
			c.ExpectedStatus = http.StatusOK
		}

		//nolint: scopelint
		t.Run(c.Name, func(t *testing.T) {
			req := httptest.NewRequest(c.Method, "/", nil)

			for k, v := range c.Headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			Handle(w, req)
			resp := w.Result()
			defer resp.Body.Close() //nolint: errcheck

			assert.Equal(t, c.ExpectedStatus, resp.StatusCode, "correct status code")

			if c.ExpectedStatus != http.StatusOK {
				return
			}

			for k, v := range c.ExpectedHeaders {
				assert.Equal(t, v, resp.Header.Get(k), "%q header", k)
			}
		})
	}
}
