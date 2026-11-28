package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edro08/go-utils/logger"
)

func TestClient_RESTMethodsAndJSON(t *testing.T) {
	type userReq struct {
		Name string `json:"name"`
	}
	type userResp struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Method string `json:"method"`
	}
	type apiError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users":
			if r.Header.Get("Authorization") != "Bearer secret-tok" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if r.Header.Get("X-Custom") != "val-1" || r.Header.Get("X-Extra") != "val-2" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if r.URL.Query().Get("active") != "true" || r.URL.Query().Get("page") != "2" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var in userReq
			if r.Body != nil && r.ContentLength > 0 {
				_ = json.NewDecoder(r.Body).Decode(&in)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(userResp{
				ID:     "u-1",
				Name:   in.Name,
				Method: r.Method,
			})

		case "/basic":
			user, pass, ok := r.BasicAuth()
			if !ok || user != "admin" || pass != "pass123" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("authenticated"))

		case "/error":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(apiError{
				Code:    "INVALID_INPUT",
				Message: "name is required",
			})

		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(r.Method))
		}
	}))
	defer server.Close()

	var client IClient = New(Options{Timeout: 5 * time.Second})

	t.Run("POST JSON with headers, bearer token, query params and SetResult", func(t *testing.T) {
		var out userResp
		resp, err := client.R(t.Context()).
			SetURL(server.URL+"/users").
			SetBearerToken("secret-tok").
			SetHeader("X-Custom", "val-1").
			SetHeaders(map[string]string{"X-Extra": "val-2"}).
			SetQueryParam("active", "true").
			SetQueryParams(map[string]string{"page": "2"}).
			SetBody(userReq{Name: "Edro"}).
			SetResult(&out).
			Post()

		if err != nil {
			t.Fatalf("Post() error = %v", err)
		}
		if !resp.IsSuccess() || resp.IsError() {
			t.Errorf("expected IsSuccess=true, IsError=false; status=%d", resp.StatusCode)
		}
		if out.ID != "u-1" || out.Name != "Edro" || out.Method != http.MethodPost {
			t.Errorf("unexpected decoded result: %+v", out)
		}
	})

	t.Run("GET with BasicAuth and String()", func(t *testing.T) {
		resp, err := client.R(t.Context()).
			SetURL(server.URL+"/basic").
			SetBasicAuth("admin", "pass123").
			Get()

		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK || resp.String() != "authenticated" {
			t.Errorf("unexpected response: status=%d body=%q", resp.StatusCode, resp.String())
		}
	})

	t.Run("PUT, PATCH, DELETE methods", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			call   func(*Request) (*Response, error)
			expect string
		}{
			{"PUT", (*Request).Put, http.MethodPut},
			{"PATCH", (*Request).Patch, http.MethodPatch},
			{"DELETE", (*Request).Delete, http.MethodDelete},
		} {
			resp, err := tc.call(client.R(t.Context()).SetURL(server.URL + "/verb"))
			if err != nil {
				t.Fatalf("%s error = %v", tc.name, err)
			}
			if resp.String() != tc.expect {
				t.Errorf("%s body = %q, want %q", tc.name, resp.String(), tc.expect)
			}
		}
	})

	t.Run("SetError decodes 4xx JSON body", func(t *testing.T) {
		var apiErr apiError
		resp, err := client.R(t.Context()).
			SetURL(server.URL + "/error").
			SetError(&apiErr).
			Get()

		if err != nil {
			t.Fatalf("Get(/error) unexpected transport error = %v", err)
		}
		if !resp.IsError() || resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 error response, got %d", resp.StatusCode)
		}
		if apiErr.Code != "INVALID_INPUT" || apiErr.Message != "name is required" {
			t.Errorf("unexpected decoded apiErr: %+v", apiErr)
		}
	})
}

func TestClient_FormAndMultipartAndRawBodies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/form":
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
			_ = r.ParseForm()
			_, _ = w.Write([]byte(r.FormValue("grant_type") + ":" + r.FormValue("scope")))

		case "/upload":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			file, header, err := r.FormFile("document")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer func() { _ = file.Close() }()
			content, _ := io.ReadAll(file)
			_, _ = w.Write([]byte(r.FormValue("folder") + "|" + header.Filename + "|" + string(content)))

		case "/raw":
			data, _ := io.ReadAll(r.Body)
			_, _ = w.Write(data)
		}
	}))
	defer server.Close()

	client := New(Options{})

	t.Run("SetFormData encodes url-encoded form", func(t *testing.T) {
		resp, err := client.R(t.Context()).
			SetURL(server.URL + "/form").
			SetFormData(map[string]string{
				"grant_type": "client_credentials",
				"scope":      "read",
			}).
			Post()

		if err != nil {
			t.Fatalf("SetFormData Post() error = %v", err)
		}
		if resp.String() != "client_credentials:read" {
			t.Errorf("body = %q, want %q", resp.String(), "client_credentials:read")
		}
	})

	t.Run("SetMultipartFormData and SetFile upload file and fields", func(t *testing.T) {
		resp, err := client.R(t.Context()).
			SetURL(server.URL+"/upload").
			SetMultipartFormData(map[string]string{"folder": "invoices"}).
			SetFile("document", "sample.txt", strings.NewReader("hello-file")).
			Post()

		if err != nil {
			t.Fatalf("Multipart Post() error = %v", err)
		}
		if resp.String() != "invoices|sample.txt|hello-file" {
			t.Errorf("body = %q, want %q", resp.String(), "invoices|sample.txt|hello-file")
		}
	})

	t.Run("Raw body supports []byte, string and io.Reader", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			body any
			want string
		}{
			{"bytes", []byte("raw-bytes"), "raw-bytes"},
			{"string", "raw-string", "raw-string"},
			{"reader", bytes.NewBufferString("raw-reader"), "raw-reader"},
		} {
			resp, err := client.R(t.Context()).
				SetURL(server.URL+"/raw").
				SetHeader("Content-Type", "text/plain").
				SetBody(tc.body).
				Post()

			if err != nil {
				t.Fatalf("%s Post() error = %v", tc.name, err)
			}
			if resp.String() != tc.want {
				t.Errorf("%s body = %q, want %q", tc.name, resp.String(), tc.want)
			}
		}
	})
}

func TestClient_MiddlewaresAndLogger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Header.Get("X-Order")))
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	type ctxKey string
	const traceKey ctxKey = "traceID"

	log, err := logger.New(logger.Options{
		MinLevel: logger.INFO,
		Format:   logger.FormatText,
		Writer:   &logBuf,
		ContextExtractor: func(ctx context.Context) []slog.Attr {
			if v, ok := ctx.Value(traceKey).(string); ok {
				return []slog.Attr{slog.String("traceID", v)}
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	var mu sync.Mutex
	var order []string

	mw := func(tag string) Middleware {
		return func(next http.RoundTripper) http.RoundTripper {
			return roundTripFunc(func(req *http.Request) (*http.Response, error) {
				mu.Lock()
				order = append(order, tag)
				mu.Unlock()

				cloned := req.Clone(req.Context())
				prev := cloned.Header.Get("X-Order")
				if prev != "" {
					prev += ","
				}
				cloned.Header.Set("X-Order", prev+tag)

				return next.RoundTrip(cloned)
			})
		}
	}

	client := New(Options{
		Timeout:     5 * time.Second,
		Logger:      log,
		Middlewares: []Middleware{mw("m1"), mw("m2")},
	})

	ctx := context.WithValue(t.Context(), traceKey, "tr-999")
	resp, err := client.R(ctx).
		SetURL(server.URL + "/ping?secret=should-not-log").
		Get()

	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.String() != "m1,m2" {
		t.Errorf("middleware header order = %q, want %q", resp.String(), "m1,m2")
	}
	if strings.Join(order, ",") != "m1,m2" {
		t.Errorf("execution order = %v, want [m1 m2]", order)
	}

	logged := logBuf.String()
	for _, want := range []string{"HTTP_CLIENT", "method=GET", "path=/ping", "status=200", "traceID=tr-999"} {
		if !strings.Contains(logged, want) {
			t.Errorf("log output %q missing %q", logged, want)
		}
	}
	if strings.Contains(logged, "should-not-log") {
		t.Errorf("log output leaked query parameter: %q", logged)
	}
}

func TestClient_ErrorsAndLimits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/large":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("0123456789ABCDEF"))
		case "/bad-json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{not-json"))
		}
	}))
	defer server.Close()

	client := New(Options{
		Timeout:     2 * time.Second,
		MaxBodySize: 8,
	})

	t.Run("Empty URL returns ErrEmptyURL", func(t *testing.T) {
		_, err := client.R(t.Context()).Get()
		if !errors.Is(err, ErrEmptyURL) {
			t.Errorf("error = %v, want %v", err, ErrEmptyURL)
		}
	})

	t.Run("Invalid URL returns ErrInvalidURL", func(t *testing.T) {
		_, err := client.R(t.Context()).SetURL("/relative-only").Get()
		if !errors.Is(err, ErrInvalidURL) {
			t.Errorf("error = %v, want %v", err, ErrInvalidURL)
		}
	})

	t.Run("Conflicting body modes returns ErrConflictingBody", func(t *testing.T) {
		_, err := client.R(t.Context()).
			SetURL(server.URL + "/large").
			SetBody(map[string]string{"a": "b"}).
			SetFormData(map[string]string{"c": "d"}).
			Post()
		if !errors.Is(err, ErrConflictingBody) {
			t.Errorf("error = %v, want %v", err, ErrConflictingBody)
		}
	})

	t.Run("Unserializable JSON body returns ErrMarshalBody", func(t *testing.T) {
		_, err := client.R(t.Context()).
			SetURL(server.URL + "/large").
			SetBody(make(chan int)).
			Post()
		if !errors.Is(err, ErrMarshalBody) {
			t.Errorf("error = %v, want %v", err, ErrMarshalBody)
		}
	})

	t.Run("Nil multipart file reader returns ErrMarshalBody", func(t *testing.T) {
		_, err := client.R(t.Context()).
			SetURL(server.URL+"/large").
			SetFile("doc", "a.txt", nil).
			Post()
		if !errors.Is(err, ErrMarshalBody) {
			t.Errorf("error = %v, want %v", err, ErrMarshalBody)
		}
	})

	t.Run("Response exceeding MaxBodySize returns ErrBodyTooLarge", func(t *testing.T) {
		_, err := client.R(t.Context()).SetURL(server.URL + "/large").Get()
		if !errors.Is(err, ErrBodyTooLarge) {
			t.Errorf("error = %v, want %v", err, ErrBodyTooLarge)
		}
	})

	t.Run("Invalid JSON response with SetResult returns ErrUnmarshalResponse", func(t *testing.T) {
		var out map[string]any
		bigClient := New(Options{MaxBodySize: 1024})
		_, err := bigClient.R(t.Context()).
			SetURL(server.URL + "/bad-json").
			SetResult(&out).
			Get()
		if !errors.Is(err, ErrUnmarshalResponse) {
			t.Errorf("error = %v, want %v", err, ErrUnmarshalResponse)
		}
	})

	t.Run("Transport error returns ErrRequestFailed and logs ErrorCtx", func(t *testing.T) {
		var logBuf bytes.Buffer
		log, _ := logger.New(logger.Options{
			MinLevel: logger.ERROR,
			Format:   logger.FormatText,
			Writer:   &logBuf,
		})

		failingClient := New(Options{
			Logger: log,
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("connection refused")
			}),
		})

		_, err := failingClient.R(t.Context()).SetURL("https://unreachable.local/test").Get()
		if !errors.Is(err, ErrRequestFailed) {
			t.Errorf("error = %v, want %v", err, ErrRequestFailed)
		}
		if !strings.Contains(logBuf.String(), "connection refused") {
			t.Errorf("expected error log, got %q", logBuf.String())
		}
	})
}
