package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response encapsula la respuesta HTTP recibida por el cliente REST.
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
}

// IsSuccess retorna true si el código de estado está en el rango 2xx.
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusMultipleChoices
}

// IsError retorna true si el código de estado es >= 400.
func (r *Response) IsError() bool {
	return r.StatusCode >= http.StatusBadRequest
}

// String devuelve el cuerpo de la respuesta como string.
func (r *Response) String() string {
	return string(r.Body)
}

// JSON decodifica el cuerpo de la respuesta dentro de out.
func (r *Response) JSON(out any) error {
	if err := json.Unmarshal(r.Body, out); err != nil {
		return fmt.Errorf("%w: %s", ErrUnmarshalResponse, err)
	}
	return nil
}
