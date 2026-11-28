/*
Package rest provee un cliente HTTP fluido para consumir APIs REST en Go.

Uso básico:

	c := rest.New(rest.Options{
		Timeout:     10 * time.Second,
		Logger:      appLogger,
		Middlewares: []rest.Middleware{traceMiddleware},
	})

	var out UserResponse
	resp, err := c.R(ctx).
		SetURL("https://api.example.com/v1/users").
		SetBearerToken(token).
		SetQueryParam("active", "true").
		SetBody(createUserReq).
		SetResult(&out).
		Post()

Comportamiento:
  - Seguro para concurrencia: un mismo *Client reutiliza el pool de conexiones HTTP y cada c.R(ctx) crea un *Request aislado.
  - Soporta cuerpos JSON, raw ([]byte, string, io.Reader), application/x-www-form-urlencoded y multipart/form-data.
  - Soporta Middlewares sobre http.RoundTripper y logging estructurado opcional mediante logger.ILogger.
  - Limita el tamaño máximo de lectura de respuesta (por defecto 10 MB) para proteger memoria.
*/
package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/edro08/go-utils/logger"
)

const (
	defaultTimeout     = 30 * time.Second
	defaultMaxBodySize = 10 << 20 // 10 MB
)

var (
	ErrEmptyURL          = errors.New("rest: request URL cannot be empty")
	ErrInvalidURL        = errors.New("rest: invalid request URL")
	ErrConflictingBody   = errors.New("rest: cannot combine JSON/raw body with form or multipart data")
	ErrMarshalBody       = errors.New("rest: failed to encode request body")
	ErrBuildRequest      = errors.New("rest: failed to build HTTP request")
	ErrRequestFailed     = errors.New("rest: HTTP request failed")
	ErrReadBody          = errors.New("rest: failed to read response body")
	ErrBodyTooLarge      = errors.New("rest: response body exceeds maximum allowed size")
	ErrUnmarshalResponse = errors.New("rest: failed to decode response body")
)

// IClient define el contrato para inyectar el cliente REST en adaptadores y servicios.
type IClient interface {
	R(ctx context.Context) *Request
}

// Middleware envuelve un http.RoundTripper para interceptar peticiones y respuestas salientes.
type Middleware func(next http.RoundTripper) http.RoundTripper

// Options configura el comportamiento del cliente REST.
type Options struct {
	Timeout     time.Duration
	Transport   http.RoundTripper
	Middlewares []Middleware
	Logger      logger.ILogger
	MaxBodySize int64
}

// Client encapsula un *http.Client reutilizable con middlewares y límites defensivos.
type Client struct {
	httpClient  *http.Client
	maxBodySize int64
}

// New construye una nueva instancia de Client lista para uso concurrente.
func New(opts Options) *Client {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	maxBodySize := opts.MaxBodySize
	if maxBodySize <= 0 {
		maxBodySize = defaultMaxBodySize
	}

	transport := opts.Transport
	if transport == nil {
		if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
			transport = defaultTransport.Clone()
		} else {
			transport = http.DefaultTransport
		}
	}

	// Aplicamos los middlewares de derecha a izquierda para que el primero
	// de la lista se ejecute primero al entrar y último al salir.
	for i := len(opts.Middlewares) - 1; i >= 0; i-- {
		if opts.Middlewares[i] != nil {
			transport = opts.Middlewares[i](transport)
		}
	}

	if opts.Logger != nil {
		transport = newLoggingRoundTripper(opts.Logger, transport)
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		maxBodySize: maxBodySize,
	}
}

// R crea una nueva petición fluida asociada al contexto dado.
func (c *Client) R(ctx context.Context) *Request {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Request{
		client:  c,
		ctx:     ctx,
		headers: make(http.Header),
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newLoggingRoundTripper(log logger.ILogger, next http.RoundTripper) http.RoundTripper {
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		start := time.Now()
		resp, err := next.RoundTrip(req)
		durationMs := time.Since(start).Milliseconds()

		host := ""
		path := ""
		if req.URL != nil {
			host = req.URL.Host
			path = req.URL.Path
		}

		if err != nil {
			log.ErrorCtx(
				req.Context(),
				"HTTP_CLIENT",
				"method", req.Method,
				"host", host,
				"path", path,
				"duration_ms", durationMs,
				"error", err.Error(),
			)
			return nil, err
		}

		log.InfoCtx(
			req.Context(),
			"HTTP_CLIENT",
			"method", req.Method,
			"host", host,
			"path", path,
			"status", resp.StatusCode,
			"duration_ms", durationMs,
		)

		return resp, nil
	})
}
