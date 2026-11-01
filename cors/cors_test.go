package cors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewCors(t *testing.T) {
	tests := []struct {
		name      string
		options   Options
		wantErr   bool
		targetErr error
	}{
		{
			name: "Valid options with origins, methods, headers and max age",
			options: Options{
				AllowedOrigins:   []string{"https://example.com", "http://localhost:3000"},
				AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
				AllowedHeaders:   []string{"Content-Type", "Authorization"},
				ExposedHeaders:   []string{"X-Trace-ID"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			wantErr: false,
		},
		{
			name: "Wildcard allowed headers sets hasWildcardHeader",
			options: Options{
				AllowedOrigins: []string{"*"},
				AllowedHeaders: []string{"*"},
			},
			wantErr: false,
		},
		{
			name: "Empty allowed origins returns ErrAllowedOriginsEmpty",
			options: Options{
				AllowedOrigins: []string{},
			},
			wantErr:   true,
			targetErr: ErrAllowedOriginsEmpty,
		},
		{
			name: "Wildcard origin with credentials returns ErrWildcardWithCredentials",
			options: Options{
				AllowedOrigins:   []string{"*"},
				AllowCredentials: true,
			},
			wantErr:   true,
			targetErr: ErrWildcardWithCredentials,
		},
		{
			name: "Negative MaxAge returns ErrInvalidMaxAge",
			options: Options{
				AllowedOrigins: []string{"https://example.com"},
				MaxAge:         -10,
			},
			wantErr:   true,
			targetErr: ErrInvalidMaxAge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCors(tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("NewCors() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && got == nil {
				t.Error("NewCors() returned nil Cors instance for valid options")
			}
		})
	}
}

func TestCors_Middleware(t *testing.T) {
	c, err := NewCors(Options{
		AllowedOrigins:   []string{"https://app.example.com", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Trace-ID"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           7200,
	})
	if err != nil {
		t.Fatalf("failed to create cors instance: %v", err)
	}

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler := c.Middleware(nextHandler)

	t.Run("Request without Origin header passes transparently to next", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if !nextCalled {
			t.Error("Middleware() did not call next handler for request without Origin")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if rec.Header().Get(allowOriginHeader) != "" {
			t.Errorf("allowOriginHeader should not be set without Origin header")
		}
	})

	t.Run("Request with disallowed origin returns 403 Forbidden", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		req.Header.Set(originHeader, "https://malicious.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if nextCalled {
			t.Error("Middleware() should not call next for disallowed origin")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("Valid GET request with allowed origin succeeds", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		req.Header.Set(originHeader, "https://app.example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if !nextCalled {
			t.Error("Middleware() should call next for allowed origin GET request")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if rec.Header().Get(allowOriginHeader) != "https://app.example.com" {
			t.Errorf("allowOriginHeader = %q, want %q", rec.Header().Get(allowOriginHeader), "https://app.example.com")
		}
		if rec.Header().Get(allowCredentialsHeader) != "true" {
			t.Errorf("allowCredentialsHeader = %q, want 'true'", rec.Header().Get(allowCredentialsHeader))
		}
		if rec.Header().Get(exposeHeadersHeader) != "X-Total-Count" {
			t.Errorf("exposeHeadersHeader = %q, want 'X-Total-Count'", rec.Header().Get(exposeHeadersHeader))
		}
	})

	t.Run("Preflight OPTIONS without Request-Method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
		req.Header.Set(originHeader, "https://app.example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", rec.Code)
		}
	})

	t.Run("Preflight OPTIONS with disallowed method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
		req.Header.Set(originHeader, "https://app.example.com")
		req.Header.Set(requestMethod, "DELETE")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 for disallowed method, got %d", rec.Code)
		}
	})

	t.Run("Preflight OPTIONS with disallowed request headers returns 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
		req.Header.Set(originHeader, "https://app.example.com")
		req.Header.Set(requestMethod, "POST")
		req.Header.Set(requestHeaders, "X-Not-Allowed-Header")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403 for disallowed request headers, got %d", rec.Code)
		}
	})

	t.Run("Valid Preflight OPTIONS returns 204 No Content with CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
		req.Header.Set(originHeader, "https://app.example.com")
		req.Header.Set(requestMethod, "POST")
		req.Header.Set(requestHeaders, "Content-Type, Authorization")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rec.Code)
		}
		if rec.Header().Get(allowMethodsHeader) == "" {
			t.Error("allowMethodsHeader should be set on preflight")
		}
		if rec.Header().Get(allowHeadersHeader) == "" {
			t.Error("allowHeadersHeader should be set on preflight")
		}
		if rec.Header().Get(maxAgeHeader) != "7200" {
			t.Errorf("maxAgeHeader = %q, want '7200'", rec.Header().Get(maxAgeHeader))
		}
	})
}

func Test_areHeadersAllowed(t *testing.T) {
	tests := []struct {
		name             string
		requestedHeaders []string
		allowedHeaders   []string
		want             bool
	}{
		{
			name:             "Empty requested headers always allowed",
			requestedHeaders: []string{},
			allowedHeaders:   []string{"Content-Type"},
			want:             true,
		},
		{
			name:             "Wildcard in allowed headers permits everything",
			requestedHeaders: []string{"X-Custom", "X-Special"},
			allowedHeaders:   []string{"*"},
			want:             true,
		},
		{
			name:             "All requested headers match case-insensitively",
			requestedHeaders: []string{"content-type", "AUTHORIZATION"},
			allowedHeaders:   []string{"Content-Type", "Authorization"},
			want:             true,
		},
		{
			name:             "One requested header not allowed",
			requestedHeaders: []string{"Content-Type", "X-Disallowed"},
			allowedHeaders:   []string{"Content-Type"},
			want:             false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := areHeadersAllowed(tt.requestedHeaders, tt.allowedHeaders); got != tt.want {
				t.Errorf("areHeadersAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_containsWildcard(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   bool
	}{
		{"Contains literal wildcard", []string{"GET", "*"}, true},
		{"Contains wildcard with spaces", []string{"  *  "}, true},
		{"Does not contain wildcard", []string{"GET", "POST"}, false},
		{"Empty slice", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsWildcard(tt.values); got != tt.want {
				t.Errorf("containsWildcard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasEmptyValues(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   bool
	}{
		{"Contains empty string", []string{"a", "", "b"}, true},
		{"Contains whitespace only", []string{"a", "   ", "b"}, true},
		{"All valid non-empty values", []string{"a", "b", "c"}, false},
		{"Empty slice", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasEmptyValues(tt.values); got != tt.want {
				t.Errorf("hasEmptyValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isMatchFold(t *testing.T) {
	tests := []struct {
		name  string
		target string
		values []string
		want  string
		want1 bool
	}{
		{"Exact match", "GET", []string{"GET", "POST"}, "GET", true},
		{"Case-insensitive match", "content-type", []string{"Content-Type"}, "Content-Type", true},
		{"Wildcard match fallback", "X-Arbitrary", []string{"Origin", "*"}, "*", true},
		{"No match", "PUT", []string{"GET", "POST"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := isMatchFold(tt.target, tt.values)
			if got != tt.want {
				t.Errorf("isMatchFold() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("isMatchFold() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_isMethodAllowed(t *testing.T) {
	tests := []struct {
		name   string
		target string
		values []string
		want   bool
	}{
		{"Empty allowed methods permits any method", "PATCH", []string{}, true},
		{"Allowed method matches", "get", []string{"GET", "POST"}, true},
		{"Disallowed method returns false", "DELETE", []string{"GET", "POST"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMethodAllowed(tt.target, tt.values); got != tt.want {
				t.Errorf("isMethodAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isOriginAllowed(t *testing.T) {
	tests := []struct {
		name   string
		target string
		values []string
		want   string
		want1  bool
	}{
		{"Exact origin match", "https://app.com", []string{"https://app.com"}, "https://app.com", true},
		{"Wildcard origin match", "https://unknown.com", []string{"*"}, "*", true},
		{"Origin not allowed", "https://notallowed.com", []string{"https://app.com"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := isOriginAllowed(tt.target, tt.values)
			if got != tt.want {
				t.Errorf("isOriginAllowed() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("isOriginAllowed() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_normalizeSlice(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []string
	}{
		{
			name:   "Trims spaces and removes empty entries",
			values: []string{"  GET ", "", " POST  ", "  "},
			want:   []string{"GET", "POST"},
		},
		{
			name:   "Empty slice returns empty slice",
			values: []string{},
			want:   []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSlice(tt.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateOptions(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
		wantErr bool
	}{
		{
			name: "Valid options",
			options: &Options{
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
				ExposedHeaders: []string{"X-Header"},
				MaxAge:         300,
			},
			wantErr: false,
		},
		{
			name: "Empty origins produces error",
			options: &Options{
				AllowedOrigins: []string{},
			},
			wantErr: true,
		},
		{
			name: "Empty string in allowed methods produces error",
			options: &Options{
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET", "  "},
			},
			wantErr: true,
		},
		{
			name: "Empty string in allowed headers produces error",
			options: &Options{
				AllowedOrigins: []string{"https://example.com"},
				AllowedHeaders: []string{""},
			},
			wantErr: true,
		},
		{
			name: "Empty string in exposed headers produces error",
			options: &Options{
				AllowedOrigins: []string{"https://example.com"},
				ExposedHeaders: []string{" "},
			},
			wantErr: true,
		},
		{
			name: "Negative max age produces error",
			options: &Options{
				AllowedOrigins: []string{"https://example.com"},
				MaxAge:         -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateOptions(tt.options); (err != nil) != tt.wantErr {
				t.Errorf("validateOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
