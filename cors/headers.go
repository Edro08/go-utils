package cors

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// setAllowMethodsHeader: Set the allow methods header
func (c *Options) setAllowMethodsHeader(w http.ResponseWriter) {
	w.Header().Set(allowMethodsHeader, strings.Join(c.AllowedMethods, ", "))
}

// setAllowCredentialsHeader: Set the allow credentials header
func (c *Options) setConnectionHeader(w http.ResponseWriter) {
	if c.Connection != "" {
		w.Header().Set(connection, c.Connection)
	}
}

// setMaxAgeHeader: Set the max age header
func (c *Options) setExposeHeadersHeader(w http.ResponseWriter) {
	if len(c.ExposedHeaders) > 0 {
		w.Header().Set(exposeHeaders, strings.Join(c.ExposedHeaders, ", "))
	}
}

// setAllowHeadersHeader: Set the allow headers header
func (c *Options) setAllowHeadersHeader(w http.ResponseWriter, requestHeaders string) {
	allowedHeaders := c.AllowedHeaders
	if isDefaultAllowed(allowedHeaders) && requestHeaders != "" {
		allowedHeaders = splitAndTrim(requestHeaders)
	}
	if len(allowedHeaders) > 0 {
		w.Header().Set(allowHeadersHeader, strings.Join(allowedHeaders, ", "))
	}
}

// setAllowCredentialsHeader: Set the allow credentials header if it is true
func (c *Options) setAllowCredentialsHeader(w http.ResponseWriter) {
	if c.AllowCredentials {
		w.Header().Set(allowCredentialsHeader, "true")
	}
}

// setMaxAgeHeader: Set the max age header if it is greater than 0
func (c *Options) setMaxAgeHeader(w http.ResponseWriter) {
	if c.MaxAge > 0 {
		w.Header().Set(maxAgeHeader, strconv.Itoa(c.MaxAge))
	}
}

// setAllowOriginHeader: Validate the origin against allowed origins and set the header
func (c *Options) setAllowOriginHeader(w http.ResponseWriter, origin string) bool {
	allowedOrigin, ok := isMatch(origin, c.AllowedOrigins)
	if !ok {
		// If not allowed, send Forbidden and response JSON
		w.Header().Set(allowOriginHeader, strings.Join(c.AllowedOrigins, ", "))
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(Cors{Message: MsgStatusForbidden})
		return false
	}

	w.Header().Set(allowOriginHeader, allowedOrigin)
	return true
}

// handlePreflightMethod: Handle OPTIONS preflight method check
func (c *Options) handlePreflightMethod(w http.ResponseWriter, method string) {
	if _, ok := isMatch(method, c.AllowedMethods); ok {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Cors{Message: MsgStatusOK})
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(Cors{Message: MsgStatusMethodNotAllowed})
	}
}

// isMatch verifica si el target está contenido en la lista values.
// Primero busca una coincidencia exacta (ignorando mayúsculas/minúsculas).
// Solo si no encuentra ninguna coincidencia, busca si existe el valor Default ("*").
// Esto permite que el origen permitido sea explícito y evita usar el wildcard
// Default cuando hay coincidencias específicas, lo que es importante para el manejo correcto
// de las cabeceras CORS, especialmente para 'Allow-Credentials'.
func isMatch(target string, values []string) (string, bool) {
	// Buscar coincidencia exacta primero
	for _, v := range values {
		if strings.EqualFold(v, target) {
			return v, true
		}
	}

	// Si no hay coincidencia, buscar Default "*"
	for _, v := range values {
		if strings.EqualFold(v, Default) {
			return v, true
		}
	}

	return "", false
}

// isDefaultAllowed: Check if the AllowedHeaders contain only the Default wildcard "*"
func isDefaultAllowed(headers []string) bool {
	return len(headers) == 1 && headers[0] == Default
}

// splitAndTrim: Split string by comma and trim spaces
func splitAndTrim(input string) []string {
	parts := strings.Split(input, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
