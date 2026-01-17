package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"base/api/internal/domain/user"
	"base/api/internal/middleware"
	"base/api/internal/session"

	"github.com/go-chi/chi/v5"
)

// NewJSONRequest creates an HTTP request with JSON body.
func NewJSONRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// WithAuthContext adds user and session to the request context.
func WithAuthContext(req *http.Request, usr *user.User, sess *session.Session) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserContextKey, usr)
	ctx = context.WithValue(ctx, middleware.SessionContextKey, sess)
	return req.WithContext(ctx)
}

// WithURLParams adds chi URL parameters to the request context.
func WithURLParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

// WithCookie adds a cookie to the request.
func WithCookie(req *http.Request, name, value string) *http.Request {
	req.AddCookie(&http.Cookie{
		Name:  name,
		Value: value,
	})
	return req
}

// ParseJSONResponse parses the response body as JSON into the given destination.
func ParseJSONResponse(w *httptest.ResponseRecorder, dst any) error {
	return json.Unmarshal(w.Body.Bytes(), dst)
}

// ResponseRecorder is a helper to create a new httptest.ResponseRecorder.
func ResponseRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// ExecuteRequest is a helper that executes a handler and returns the response.
func ExecuteRequest(handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}
