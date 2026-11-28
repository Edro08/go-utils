package cors

import (
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCors_setAllowCredentialsHeader(t *testing.T) {
	tests := []struct {
		name       string
		allowCreds bool
		wantHeader string
	}{
		{
			name:       "Allow credentials true sets header",
			allowCreds: true,
			wantHeader: "true",
		},
		{
			name:       "Allow credentials false does not set header",
			allowCreds: false,
			wantHeader: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cors{
				options: Options{AllowCredentials: tt.allowCreds},
			}
			w := httptest.NewRecorder()
			c.setAllowCredentialsHeader(w)
			got := w.Header().Get(allowCredentialsHeader)
			if got != tt.wantHeader {
				t.Errorf("setAllowCredentialsHeader() = %v, want %v", got, tt.wantHeader)
			}
		})
	}
}

func TestCors_setAllowHeadersHeader(t *testing.T) {
	tests := []struct {
		name             string
		hasWildcard      bool
		allowCreds       bool
		formattedHeaders string
		requestedHeaders []string
		wantHeader       string
	}{
		{
			name:             "Wildcard with requested headers mirrors requested",
			hasWildcard:      true,
			requestedHeaders: []string{"X-Trace-ID", "Content-Type"},
			wantHeader:       "X-Trace-ID, Content-Type",
		},
		{
			name:             "Wildcard without requested headers sets *",
			hasWildcard:      true,
			requestedHeaders: nil,
			wantHeader:       Default,
		},
		{
			name:             "Wildcard without requested headers and AllowCredentials true omits header",
			hasWildcard:      true,
			allowCreds:       true,
			requestedHeaders: nil,
			wantHeader:       "",
		},
		{
			name:             "Configured formatted headers",
			hasWildcard:      false,
			formattedHeaders: "Authorization, Content-Type",
			requestedHeaders: []string{"Authorization"},
			wantHeader:       "Authorization, Content-Type",
		},
		{
			name:             "Empty formatted headers does not set header",
			hasWildcard:      false,
			formattedHeaders: "",
			requestedHeaders: nil,
			wantHeader:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cors{
				options:           Options{AllowCredentials: tt.allowCreds},
				hasWildcardHeader: tt.hasWildcard,
				formattedHeaders:  tt.formattedHeaders,
			}
			w := httptest.NewRecorder()
			c.setAllowHeadersHeader(w, tt.requestedHeaders)
			got := w.Header().Get(allowHeadersHeader)
			if got != tt.wantHeader {
				t.Errorf("setAllowHeadersHeader() = %v, want %v", got, tt.wantHeader)
			}
		})
	}
}

func TestCors_setAllowMethodsHeader(t *testing.T) {
	tests := []struct {
		name              string
		formattedMethods  string
		hasWildcardMethod bool
		allowCredentials  bool
		requestedMethod   string
		wantHeader        string
	}{
		{
			name:             "Formatted methods present sets header",
			formattedMethods: "GET, POST, OPTIONS",
			requestedMethod:  "POST",
			wantHeader:       "GET, POST, OPTIONS",
		},
		{
			name:             "Empty formatted methods does not set header",
			formattedMethods: "",
			wantHeader:       "",
		},
		{
			name:              "Wildcard method with AllowCredentials mirrors requested method in uppercase",
			formattedMethods:  "*",
			hasWildcardMethod: true,
			allowCredentials:  true,
			requestedMethod:   "patch",
			wantHeader:        "PATCH",
		},
		{
			name:              "Wildcard method without AllowCredentials sets wildcard",
			formattedMethods:  "*",
			hasWildcardMethod: true,
			allowCredentials:  false,
			requestedMethod:   "PATCH",
			wantHeader:        "*",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cors{
				options:           Options{AllowCredentials: tt.allowCredentials},
				formattedMethods:  tt.formattedMethods,
				hasWildcardMethod: tt.hasWildcardMethod,
			}
			w := httptest.NewRecorder()
			c.setAllowMethodsHeader(w, tt.requestedMethod)
			got := w.Header().Get(allowMethodsHeader)
			if got != tt.wantHeader {
				t.Errorf("setAllowMethodsHeader() = %v, want %v", got, tt.wantHeader)
			}
		})
	}
}

func TestCors_setAllowOriginHeader(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		wantOrigin string
		wantVary   bool
	}{
		{
			name:       "Specific origin sets header and Vary",
			origin:     "https://example.com",
			wantOrigin: "https://example.com",
			wantVary:   true,
		},
		{
			name:       "Wildcard origin sets header without Vary",
			origin:     Default,
			wantOrigin: Default,
			wantVary:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cors{}
			w := httptest.NewRecorder()
			c.setAllowOriginHeader(w, tt.origin)

			if got := w.Header().Get(allowOriginHeader); got != tt.wantOrigin {
				t.Errorf("setAllowOriginHeader() = %v, want %v", got, tt.wantOrigin)
			}

			vary := w.Header().Get("Vary")
			hasVary := vary == originHeader
			if hasVary != tt.wantVary {
				t.Errorf("setAllowOriginHeader() vary present = %v, want %v (Vary: %q)", hasVary, tt.wantVary, vary)
			}
		})
	}
}

func TestCors_setExposeHeadersHeader(t *testing.T) {
	tests := []struct {
		name       string
		formatted  string
		wantHeader string
	}{
		{
			name:       "Formatted exposed headers present sets header",
			formatted:  "X-Custom-Header, X-Total-Count",
			wantHeader: "X-Custom-Header, X-Total-Count",
		},
		{
			name:       "Empty exposed headers does not set header",
			formatted:  "",
			wantHeader: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cors{formattedExposed: tt.formatted}
			w := httptest.NewRecorder()
			c.setExposeHeadersHeader(w)
			got := w.Header().Get(exposeHeadersHeader)
			if got != tt.wantHeader {
				t.Errorf("setExposeHeadersHeader() = %v, want %v", got, tt.wantHeader)
			}
		})
	}
}

func TestCors_setMaxAgeHeader(t *testing.T) {
	tests := []struct {
		name       string
		formatted  string
		wantHeader string
	}{
		{
			name:       "MaxAge formatted present sets header",
			formatted:  "3600",
			wantHeader: "3600",
		},
		{
			name:       "Empty formatted max age does not set header",
			formatted:  "",
			wantHeader: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cors{formattedMaxAge: tt.formatted}
			w := httptest.NewRecorder()
			c.setMaxAgeHeader(w)
			got := w.Header().Get(maxAgeHeader)
			if got != tt.wantHeader {
				t.Errorf("setMaxAgeHeader() = %v, want %v", got, tt.wantHeader)
			}
		})
	}
}

func TestCors_setPreflightVaryHeaders(t *testing.T) {
	c := &Cors{}
	w := httptest.NewRecorder()
	c.setPreflightVaryHeaders(w)

	varyHeaders := w.Header().Values("Vary")
	expected := []string{originHeader, requestMethod, requestHeaders}

	if !reflect.DeepEqual(varyHeaders, expected) {
		t.Errorf("setPreflightVaryHeaders() = %v, want %v", varyHeaders, expected)
	}
}

func Test_addVary(t *testing.T) {
	w := httptest.NewRecorder()
	addVary(w, "Accept-Encoding")
	addVary(w, "Origin")
	addVary(w, "origin")

	varyValues := w.Header().Values("Vary")
	expected := []string{"Accept-Encoding", "Origin"}
	if !reflect.DeepEqual(varyValues, expected) {
		t.Errorf("addVary() = %v, want %v", varyValues, expected)
	}
}

func Test_extractRequestHeaders(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []string
	}{
		{
			name:   "Empty values returns nil",
			values: nil,
			want:   nil,
		},
		{
			name:   "Single header with comma-separated values",
			values: []string{"Content-Type, Authorization"},
			want:   []string{"Content-Type", "Authorization"},
		},
		{
			name:   "Multiple header lines combined and trimmed",
			values: []string{"Content-Type, X-Trace-ID", " Authorization , X-Custom "},
			want:   []string{"Content-Type", "X-Trace-ID", "Authorization", "X-Custom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRequestHeaders(tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractRequestHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_splitAndTrim(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "Empty string returns nil",
			input: "",
			want:  nil,
		},
		{
			name:  "Whitespace string returns nil",
			input: "   ",
			want:  nil,
		},
		{
			name:  "Comma-separated values with spaces trimmed",
			input: "Content-Type,  Authorization , X-Trace-ID",
			want:  []string{"Content-Type", "Authorization", "X-Trace-ID"},
		},
		{
			name:  "Extra consecutive commas ignored",
			input: "A,, B , ,C",
			want:  []string{"A", "B", "C"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitAndTrim(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitAndTrim() = %v, want %v", got, tt.want)
			}
		})
	}
}
