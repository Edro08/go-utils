/*
Package cors provee un middleware para manejar CORS (Cross-Origin Resource Sharing)
en aplicaciones web escritas en Go.

Este paquete facilita la configuración de políticas CORS como:
- Orígenes permitidos
- Métodos HTTP permitidos
- Cabeceras permitidas y expuestas
- Credenciales
- Tiempo de caché para preflight requests

Uso básico:

	options := cors.Options{
	    AllowedOrigins:   []string{"https://example.com"},
	    AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
	    AllowedHeaders:   []string{"Content-Type", "Authorization"},
	    ExposedHeaders:   []string{"X-Custom-Header"},
	    AllowCredentials: true,
	    MaxAge:           3600,
	    Connection:       "keep-alive",
	}

	corsHandler := cors.NewCorsHandler(options)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	    corsHandler.CorsHandler(w, r)
	    // lógica adicional para manejar la petición
	})

Estructuras y funciones principales:

  - type Options: configuración principal para CORS, incluye listas de orígenes, métodos,
    cabeceras permitidas y otros parámetros.

- func NewCors(option Options) *Options: constructor para crear un handler CORS configurado.

  - func (c *Options) CorsHandler(w http.ResponseWriter, r *http.Request):
    maneja la respuesta CORS agregando los headers necesarios según la configuración
    y el request recibido.

  - Métodos auxiliares para establecer cada header CORS:
    setAllowMethodsHeader, setAllowHeadersHeader, setAllowOriginHeader, etc.

  - Función interna isMatch(target string, values []string) (string, bool):
    verifica si un valor está contenido en una lista (ignorando mayúsculas/minúsculas)
    y permite la presencia del valor wildcard "*".

- Función splitAndTrim(input string) []string: ayuda a procesar cabeceras separadas por coma.

Mensajes de estado para respuesta JSON:

  - MsgStatusForbidden: acceso denegado cuando el origen no está permitido.
  - MsgStatusMethodNotAllowed: método HTTP no permitido.
  - MsgStatusOK: petición preflight procesada correctamente.
*/
package cors

import (
	"net/http"
)

type ICors interface {
	CorsHandler(w http.ResponseWriter, r *http.Request)
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

func (c *Options) CorsHandler(w http.ResponseWriter, r *http.Request) {
	// Set all standard CORS headers
	c.setAllowMethodsHeader(w)
	c.setConnectionHeader(w)
	c.setExposeHeadersHeader(w)
	c.setAllowHeadersHeader(w, r.Header.Get(requestHeaders))
	c.setAllowCredentialsHeader(w)
	c.setMaxAgeHeader(w)

	// Validate and set Allow-Origin header, block if not allowed
	if !c.setAllowOriginHeader(w, r.Header.Get(originHeader)) {
		return
	}

	// If it is not a preflight OPTIONS request, just return (normal request)
	if r.Method != http.MethodOptions {
		return
	}

	// Handle preflight OPTIONS request method validation
	c.handlePreflightMethod(w, r.Header.Get(requestMethod))
}
