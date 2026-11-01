/*
Package cors provee un middleware para manejar CORS (Cross-Origin Resource Sharing)
en aplicaciones web escritas en Go.

Uso básico:

	c, err := cors.NewCors(cors.Options{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	if err != nil {
		log.Fatal(err)
	}

	http.ListenAndServe(":8080", c.Middleware(mux))

Comportamiento:
  - Sin header Origin: pasa al siguiente handler sin modificar la respuesta.
  - Origen no permitido: retorna 403 Forbidden.
  - Preflight válido: retorna 204 No Content.
  - Método del preflight no permitido: retorna 405 Method Not Allowed.
  - Header solicitado no permitido: retorna 403 Forbidden.
*/
package cors

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type CorsMiddleware interface {
	Middleware(next http.Handler) http.Handler
}

var (
	ErrAllowedOriginsEmpty     = errors.New("cors: allowed origins cannot be empty")
	ErrWildcardWithCredentials = errors.New("cors: wildcard origin cannot be used with credentials")
	ErrEmptyAllowedOrigin      = errors.New("cors: allowed origins cannot contain empty values")
	ErrEmptyAllowedMethod      = errors.New("cors: allowed methods cannot contain empty values")
	ErrEmptyAllowedHeader      = errors.New("cors: allowed headers cannot contain empty values")
	ErrEmptyExposedHeader      = errors.New("cors: exposed headers cannot contain empty values")
	ErrInvalidMaxAge           = errors.New("cors: max age cannot be negative")
)

const (
	allowOriginHeader      = "Access-Control-Allow-Origin"
	allowMethodsHeader     = "Access-Control-Allow-Methods"
	allowHeadersHeader     = "Access-Control-Allow-Headers"
	allowCredentialsHeader = "Access-Control-Allow-Credentials"
	maxAgeHeader           = "Access-Control-Max-Age"
	exposeHeadersHeader    = "Access-Control-Expose-Headers"

	requestMethod  = "Access-Control-Request-Method"
	requestHeaders = "Access-Control-Request-Headers"
	originHeader   = "Origin"

	MsgStatusMethodNotAllowed = "The HTTP method used is not allowed for this resource."
	MsgStatusForbidden        = "You do not have permission to access this resource."

	Default = "*"
)

type Cors struct {
	options           Options
	formattedMethods  string
	formattedHeaders  string
	formattedExposed  string
	formattedMaxAge   string
	hasWildcardHeader bool
}

type Options struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func NewCors(options Options) (*Cors, error) {
	if err := validateOptions(&options); err != nil {
		return nil, err
	}

	c := &Cors{
		options:           options,
		hasWildcardHeader: containsWildcard(options.AllowedHeaders),
	}

	// Normalizamos y preformateamos cabeceras para optimizar las peticiones HTTP
	if len(options.AllowedMethods) > 0 {
		c.formattedMethods = strings.Join(normalizeSlice(options.AllowedMethods), ", ")
	}
	if len(options.AllowedHeaders) > 0 {
		c.formattedHeaders = strings.Join(normalizeSlice(options.AllowedHeaders), ", ")
	}
	if len(options.ExposedHeaders) > 0 {
		c.formattedExposed = strings.Join(normalizeSlice(options.ExposedHeaders), ", ")
	}
	if options.MaxAge > 0 {
		c.formattedMaxAge = strconv.Itoa(options.MaxAge)
	}

	return c, nil
}

func (c *Cors) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get(originHeader)

		// Las peticiones sin Origin no están sujetas a CORS.
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		allowedOrigin, ok := isOriginAllowed(origin, c.options.AllowedOrigins)
		if !ok {
			http.Error(w, MsgStatusForbidden, http.StatusForbidden)
			return
		}

		c.setAllowOriginHeader(w, allowedOrigin)
		c.setExposeHeadersHeader(w)
		c.setAllowCredentialsHeader(w)

		// Las peticiones que no son preflight continúan al siguiente handler.
		if r.Method != http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		c.setPreflightVaryHeaders(w)

		method := r.Header.Get(requestMethod)
		if method == "" {
			http.Error(w, MsgStatusMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}

		if !isMethodAllowed(method, c.options.AllowedMethods) {
			http.Error(w, MsgStatusMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}

		requestedHeaders := splitAndTrim(r.Header.Get(requestHeaders))

		if !areHeadersAllowed(requestedHeaders, c.options.AllowedHeaders) {
			http.Error(w, MsgStatusForbidden, http.StatusForbidden)
			return
		}

		c.setAllowMethodsHeader(w)
		c.setAllowHeadersHeader(w, requestedHeaders)
		c.setMaxAgeHeader(w)

		w.WriteHeader(http.StatusNoContent)
	})
}

func validateOptions(options *Options) error {
	var errs []error

	if len(options.AllowedOrigins) == 0 {
		errs = append(errs, ErrAllowedOriginsEmpty)
	}

	if options.AllowCredentials && containsWildcard(options.AllowedOrigins) {
		errs = append(errs, ErrWildcardWithCredentials)
	}

	if hasEmptyValues(options.AllowedOrigins) {
		errs = append(errs, ErrEmptyAllowedOrigin)
	}

	if hasEmptyValues(options.AllowedMethods) {
		errs = append(errs, ErrEmptyAllowedMethod)
	}

	if hasEmptyValues(options.AllowedHeaders) {
		errs = append(errs, ErrEmptyAllowedHeader)
	}

	if hasEmptyValues(options.ExposedHeaders) {
		errs = append(errs, ErrEmptyExposedHeader)
	}

	if options.MaxAge < 0 {
		errs = append(errs, ErrInvalidMaxAge)
	}

	return errors.Join(errs...)
}

func containsWildcard(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == Default {
			return true
		}
	}

	return false
}

func isOriginAllowed(target string, values []string) (string, bool) {
	for _, value := range values {
		value = strings.TrimSpace(value)

		if value == target {
			return value, true
		}
	}

	for _, value := range values {
		value = strings.TrimSpace(value)

		if value == Default {
			return value, true
		}
	}

	return "", false
}

func isMethodAllowed(target string, values []string) bool {
	if len(values) == 0 {
		return true
	}
	_, ok := isMatchFold(target, values)
	return ok
}

// isMatchFold verifica una coincidencia case-insensitive.
// Se utiliza para valores como métodos HTTP y nombres de headers.
func isMatchFold(target string, values []string) (string, bool) {
	target = strings.TrimSpace(target)

	for _, value := range values {
		value = strings.TrimSpace(value)

		if strings.EqualFold(value, target) {
			return value, true
		}
	}

	for _, value := range values {
		value = strings.TrimSpace(value)

		if value == Default {
			return value, true
		}
	}

	return "", false
}

// areHeadersAllowed verifica que todos los headers solicitados por el cliente
// estén permitidos por la configuración.
func areHeadersAllowed(requestedHeaders, allowedHeaders []string) bool {
	if len(requestedHeaders) == 0 {
		return true
	}

	if containsWildcard(allowedHeaders) {
		return true
	}

	for _, requested := range requestedHeaders {
		if _, ok := isMatchFold(requested, allowedHeaders); !ok {
			return false
		}
	}

	return true
}

func hasEmptyValues(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}

	return false
}

func normalizeSlice(values []string) []string {
	result := make([]string, 0, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
