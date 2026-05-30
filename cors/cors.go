/*
Package cors provee un middleware para manejar CORS (Cross-Origin Resource Sharing)
en aplicaciones web escritas en Go.

Uso básico:

	options := cors.Options{
	    AllowedOrigins:   []string{"https://example.com"},
	    AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
	    AllowedHeaders:   []string{"Content-Type", "Authorization"},
	    AllowCredentials: true,
	    MaxAge:           3600,
	}

	c := cors.NewCors(options)

	http.ListenAndServe(":8080", c.Middleware(mux))

Estructuras y funciones principales:

  - type Options: configuración de políticas CORS (orígenes, métodos, cabeceras).

  - func NewCors(option Options) *Options: constructor.

  - func (c *Options) Middleware(next http.Handler) http.Handler:
    envuelve un handler HTTP y maneja CORS automáticamente. Si el origen no está
    permitido, retorna 403. Si es una petición preflight (OPTIONS), valida el
    método y retorna 204; en caso contrario, delega al siguiente handler.

Comportamiento:
  - Sin header Origin: pasa al siguiente handler sin modificar la respuesta.
  - Origen no permitido: retorna 403 Forbidden.
  - OPTIONS permitido: retorna 204 No Content.
  - OPTIONS denegado: retorna 405 Method Not Allowed.
*/
package cors

import (
	"net/http"
)

type ICors interface {
	Middleware(next http.Handler) http.Handler
}

const (
	allowOriginHeader      = "Access-Control-Allow-Origin"
	allowMethodsHeader     = "Access-Control-Allow-Methods"
	allowHeadersHeader     = "Access-Control-Allow-Headers"
	allowCredentialsHeader = "Access-Control-Allow-Credentials"
	maxAgeHeader           = "Access-Control-Max-Age"
	exposeHeaders          = "Access-Control-Expose-Headers"
	connection             = "Connection"

	requestMethod  = "Access-Control-Request-Method"
	requestHeaders = "Access-Control-Request-Headers"
	originHeader   = "Origin"

	MsgStatusMethodNotAllowed = "The HTTP method used is not allowed for this resource."
	MsgStatusForbidden        = "You do not have permission to access this resource."
	MsgStatusOK               = "Request processed successfully."

	Default = "*"
)

type Cors struct {
	Message string `json:"message"`
}

type Options struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
	Connection       string
}

func NewCors(option Options) *Options {
	return &option
}

func (c *Options) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get(originHeader)
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		isPreflight := r.Method == http.MethodOptions

		c.setConnectionHeader(w)
		c.setExposeHeadersHeader(w)
		c.setAllowCredentialsHeader(w)

		if isPreflight {
			c.setAllowMethodsHeader(w)
			c.setAllowHeadersHeader(w, r.Header.Get(requestHeaders))
			c.setMaxAgeHeader(w)
		}

		if !c.setAllowOriginHeader(w, origin) {
			return
		}

		if isPreflight {
			c.handlePreflightMethod(w, r.Header.Get(requestMethod))
			return
		}

		next.ServeHTTP(w, r)
	})
}
