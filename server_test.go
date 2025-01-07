package health_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"

	"github.com/dotse/go-health"
)

func TestHandle(t *testing.T) {
	var (
		ctx    = context.Background()
		status health.Status
	)

	r := health.RegisterFunc(ctx, t.Name(), func(context.Context) []health.Check {
		return []health.Check{
			{
				Status: status,
			},
		}
	})
	defer r.Deregister()

	test := func(
		name string,
		method string,
		requestHeaders map[string]string,
		expectedHeaders map[string]string,
		expectedStatus int,
	) {
		t.Helper()

		t.Run(name, func(t *testing.T) {
			t.Helper()

			if expectedHeaders == nil {
				expectedHeaders = make(map[string]string)
			}

			if _, ok := expectedHeaders[headers.ContentType]; !ok {
				expectedHeaders[headers.ContentType] = health.ContentType
			}

			req := httptest.NewRequest(method, "/", nil)

			for k, v := range requestHeaders {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			health.HandleHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, expectedStatus, resp.StatusCode, "correct status code")

			if expectedStatus != http.StatusOK {
				return
			}

			for k, v := range expectedHeaders {
				assert.Equal(t, v, resp.Header.Get(k), "%q header", k)
			}
		})
	}

	test(
		"bad type",
		http.MethodGet,
		map[string]string{headers.Accept: "text/html"},
		nil,
		http.StatusNotAcceptable,
	)

	test(
		"bad charset",
		http.MethodGet,
		map[string]string{headers.AcceptCharset: "iso-8859-1"},
		nil,
		http.StatusNotAcceptable,
	)

	test(
		"good type, bad charset",
		http.MethodGet,
		map[string]string{
			headers.Accept:        health.ContentType,
			headers.AcceptCharset: "iso-8859-1",
		},
		nil,
		http.StatusNotAcceptable,
	)

	test(
		"good charset, bad type",
		http.MethodGet,
		map[string]string{
			headers.Accept:        "text/html",
			headers.AcceptCharset: "UTF-8",
		},
		nil,
		http.StatusNotAcceptable,
	)

	test(
		http.MethodPost,
		http.MethodPost,
		nil,
		nil,
		http.StatusMethodNotAllowed,
	)

	test(
		http.MethodOptions,
		http.MethodOptions,
		nil,
		map[string]string{headers.Allow: "OPTIONS, GET, HEAD"},
		http.StatusNoContent,
	)

	test(
		http.MethodHead,
		http.MethodHead,
		nil,
		map[string]string{headers.ContentLength: "0"},
		http.StatusOK,
	)

	test(
		"no headers",
		http.MethodGet,
		nil,
		nil,
		http.StatusOK,
	)

	test(
		"alternative content-type good",
		http.MethodGet,
		map[string]string{headers.Accept: "application/json"},
		map[string]string{headers.ContentType: "application/json"},
		http.StatusOK,
	)

	test(
		"good charset, good type",
		http.MethodGet,
		map[string]string{
			headers.Accept:        health.ContentType,
			headers.AcceptCharset: "UTF-8",
		},
		nil,
		http.StatusOK,
	)

	status = health.StatusFail
	test(
		"bad status",
		http.MethodGet,
		nil,
		nil,
		http.StatusInternalServerError,
	)
}
