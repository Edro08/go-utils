package cors

import (
	"net/http"
	"strings"
)

// setAllowMethodsHeader establece Access-Control-Allow-Methods.
// Cuando AllowCredentials está activo y se configuró wildcard ("*"),
// refleja el método solicitado en lugar de enviar "*" para cumplir con la spec CORS.
func (c *Cors) setAllowMethodsHeader(w http.ResponseWriter, requestedMethod string) {
	if c.hasWildcardMethod && c.options.AllowCredentials {
		if requestedMethod != "" {
			w.Header().Set(allowMethodsHeader, strings.ToUpper(requestedMethod))
		}
		return
	}

	if c.formattedMethods == "" {
		return
	}

	w.Header().Set(allowMethodsHeader, c.formattedMethods)
}

// setExposeHeadersHeader establece Access-Control-Expose-Headers.
func (c *Cors) setExposeHeadersHeader(w http.ResponseWriter) {
	if c.formattedExposed == "" {
		return
	}

	w.Header().Set(exposeHeadersHeader, c.formattedExposed)
}

// setAllowHeadersHeader establece Access-Control-Allow-Headers.
func (c *Cors) setAllowHeadersHeader(w http.ResponseWriter, requestedHeaders []string) {
	if c.hasWildcardHeader {
		if len(requestedHeaders) > 0 {
			w.Header().Set(allowHeadersHeader, strings.Join(requestedHeaders, ", "))
		} else if !c.options.AllowCredentials {
			w.Header().Set(allowHeadersHeader, Default)
		}
		return
	}

	if c.formattedHeaders != "" {
		w.Header().Set(allowHeadersHeader, c.formattedHeaders)
	}
}

// setAllowCredentialsHeader establece Access-Control-Allow-Credentials.
func (c *Cors) setAllowCredentialsHeader(w http.ResponseWriter) {
	if c.options.AllowCredentials {
		w.Header().Set(allowCredentialsHeader, "true")
	}
}

// setMaxAgeHeader establece Access-Control-Max-Age.
func (c *Cors) setMaxAgeHeader(w http.ResponseWriter) {
	if c.formattedMaxAge == "" {
		return
	}

	w.Header().Set(maxAgeHeader, c.formattedMaxAge)
}

// setAllowOriginHeader establece Access-Control-Allow-Origin.
func (c *Cors) setAllowOriginHeader(
	w http.ResponseWriter,
	origin string,
) {
	w.Header().Set(allowOriginHeader, origin)

	if origin != Default || c.varyOrigin {
		addVary(w, originHeader)
	}
}

// setPreflightVaryHeaders indica que la respuesta de preflight depende
// de los valores utilizados para determinar la política CORS.
func (c *Cors) setPreflightVaryHeaders(w http.ResponseWriter) {
	addVary(w, originHeader)
	addVary(w, requestMethod)
	addVary(w, requestHeaders)
}

// addVary agrega un valor al header Vary evitando duplicados.
func addVary(w http.ResponseWriter, value string) {
	for _, existing := range w.Header().Values("Vary") {
		if strings.EqualFold(existing, value) {
			return
		}
	}

	w.Header().Add("Vary", value)
}

// extractRequestHeaders extrae y normaliza todos los valores de
// Access-Control-Request-Headers, soportando múltiples líneas de header.
func extractRequestHeaders(headerValues []string) []string {
	var result []string

	for _, raw := range headerValues {
		parsed := splitAndTrim(raw)
		result = append(result, parsed...)
	}

	return result
}

// splitAndTrim separa una lista de valores delimitada por comas,
// eliminando espacios y entradas vacías.
func splitAndTrim(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
