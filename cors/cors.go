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
