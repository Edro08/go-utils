/*
Package soap provee un cliente HTTP fluido para consumir servicios web SOAP (1.1 y 1.2) en Go.

Uso básico:

	c := soap.New(soap.Options{
		Version:     soap.Version11,
		Timeout:     15 * time.Second,
		Logger:      appLogger,
		Middlewares: []soap.Middleware{traceMiddleware},
	})

	var out GetBalanceResponse
	resp, err := c.R(ctx).
		SetURL("https://legacy.example.com/Service.asmx").
		SetAction("http://tempuri.org/GetBalance").
		SetSOAPHeader(authHeader).
		SetBody(GetBalanceRequest{AccountID: "123"}).
		SetResult(&out).
		Send()

Comportamiento:
  - Seguro para concurrencia: un mismo *Client reutiliza el pool de conexiones HTTP y cada c.R(ctx) crea un *Request aislado.
  - Soporta SOAP 1.1 (text/xml + header SOAPAction) y SOAP 1.2 (application/soap+xml + parámetro action).
  - Envuelve automáticamente el payload en <soap:Envelope> con <soap:Header> opcional y <soap:Body>.
  - Detecta y parsea respuestas <Fault> (SOAP 1.1 y 1.2) devolviendo *FaultError envuelto sobre ErrFault.
  - Soporta Middlewares sobre http.RoundTripper y logging estructurado opcional mediante logger.ILogger.
*/
package soap

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/edro08/go-utils/logger"
)

// Version representa la versión del protocolo SOAP a utilizar.
type Version string

const (
	Version11 Version = "1.1"
	Version12 Version = "1.2"

	namespace11 = "http://schemas.xmlsoap.org/soap/envelope/"
	namespace12 = "http://www.w3.org/2003/05/soap-envelope"

	defaultTimeout     = 30 * time.Second
	defaultMaxBodySize = 10 << 20 // 10 MB
)

var (
	ErrEmptyURL          = errors.New("soap: request URL cannot be empty")
	ErrInvalidURL        = errors.New("soap: invalid request URL")
	ErrEmptyBody         = errors.New("soap: request body or raw envelope is required")
	ErrMarshalEnvelope   = errors.New("soap: failed to encode SOAP envelope")
	ErrBuildRequest      = errors.New("soap: failed to build HTTP request")
	ErrRequestFailed     = errors.New("soap: HTTP request failed")
	ErrReadBody          = errors.New("soap: failed to read response body")
	ErrBodyTooLarge      = errors.New("soap: response body exceeds maximum allowed size")
	ErrUnmarshalResponse = errors.New("soap: failed to decode SOAP response")
	ErrFault             = errors.New("soap: server returned SOAP fault")
)

// IClient define el contrato para inyectar el cliente SOAP en adaptadores y servicios.
type IClient interface {
	R(ctx context.Context) *Request
}

// Middleware envuelve un http.RoundTripper para interceptar peticiones y respuestas salientes.
type Middleware func(next http.RoundTripper) http.RoundTripper

// Options configura el comportamiento del cliente SOAP.
type Options struct {
	Version     Version
	Timeout     time.Duration
	Transport   http.RoundTripper
	Middlewares []Middleware
	Logger      logger.ILogger
	MaxBodySize int64
}

// Client encapsula un *http.Client reutilizable configurado para llamadas SOAP.
type Client struct {
	httpClient     *http.Client
	defaultVersion Version
	maxBodySize    int64
}

// New construye una nueva instancia de Client lista para uso concurrente.
func New(opts Options) *Client {
	version := opts.Version
	if version != Version12 {
		version = Version11
	}

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
		defaultVersion: version,
		maxBodySize:    maxBodySize,
	}
}

// R crea una nueva petición SOAP fluida asociada al contexto dado.
func (c *Client) R(ctx context.Context) *Request {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Request{
		client:  c,
		ctx:     ctx,
		version: c.defaultVersion,
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
				"SOAP_CLIENT",
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
			"SOAP_CLIENT",
			"method", req.Method,
			"host", host,
			"path", path,
			"status", resp.StatusCode,
			"duration_ms", durationMs,
		)

		return resp, nil
	})
}
