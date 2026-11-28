package soap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FaultError representa un error <Fault> devuelto por el servidor SOAP (1.1 o 1.2).
type FaultError struct {
	Code    string
	Message string
	Actor   string
	Detail  string
}

// Error implementa la interfaz error.
func (f *FaultError) Error() string {
	if f.Code != "" && f.Message != "" {
		return fmt.Sprintf("%s: [%s] %s", ErrFault, f.Code, f.Message)
	}
	if f.Message != "" {
		return fmt.Sprintf("%s: %s", ErrFault, f.Message)
	}
	return ErrFault.Error()
}

// Unwrap permite detectar FaultError con errors.Is(err, soap.ErrFault).
func (f *FaultError) Unwrap() error {
	return ErrFault
}

// Response encapsula la respuesta HTTP y el sobre XML devuelto por el servicio SOAP.
type Response struct {
	StatusCode  int
	Status      string
	Headers     http.Header
	Body        []byte
	BodyContent []byte
	Fault       *FaultError
	Duration    time.Duration
}

// IsSuccess retorna true si el status HTTP es 2xx y no existe Fault SOAP.
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusMultipleChoices && r.Fault == nil
}

// String devuelve el XML completo de la respuesta como string.
func (r *Response) String() string {
	return string(r.Body)
}

// UnmarshalBody decodifica el contenido interno de <soap:Body> dentro de out.
func (r *Response) UnmarshalBody(out any) error {
	if err := xml.Unmarshal(r.BodyContent, out); err != nil {
		return fmt.Errorf("%w: %s", ErrUnmarshalResponse, err)
	}
	return nil
}

type incomingEnvelope struct {
	Body incomingBody `xml:"Body"`
}

type incomingBody struct {
	Fault   *incomingFault `xml:"Fault"`
	Content []byte         `xml:",innerxml"`
}

type incomingFault struct {
	// Campos SOAP 1.1
	FaultCode   string        `xml:"faultcode"`
	FaultString string        `xml:"faultstring"`
	FaultActor  string        `xml:"faultactor"`
	Detail11    *rawXMLDetail `xml:"detail"`

	// Campos SOAP 1.2
	Code12   *faultCode12   `xml:"Code"`
	Reason12 *faultReason12 `xml:"Reason"`
	Detail12 *rawXMLDetail  `xml:"Detail"`
}

type faultCode12 struct {
	Value string `xml:"Value"`
}

type faultReason12 struct {
	Text string `xml:"Text"`
}

type rawXMLDetail struct {
	Inner []byte `xml:",innerxml"`
}

func parseEnvelopeResponse(data []byte) ([]byte, *FaultError, error) {
	var env incomingEnvelope
	if err := xml.Unmarshal(data, &env); err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnmarshalResponse, err)
	}

	trimmedContent := bytes.TrimSpace(env.Body.Content)
	if env.Body.Fault == nil {
		return trimmedContent, nil, nil
	}

	fault := normalizeFault(env.Body.Fault)
	return trimmedContent, fault, nil
}

func normalizeFault(raw *incomingFault) *FaultError {
	code := strings.TrimSpace(raw.FaultCode)
	message := strings.TrimSpace(raw.FaultString)
	actor := strings.TrimSpace(raw.FaultActor)
	detail := ""

	if raw.Detail11 != nil {
		detail = strings.TrimSpace(string(raw.Detail11.Inner))
	}

	if code == "" && raw.Code12 != nil {
		code = strings.TrimSpace(raw.Code12.Value)
	}

	if message == "" && raw.Reason12 != nil {
		message = strings.TrimSpace(raw.Reason12.Text)
	}

	if detail == "" && raw.Detail12 != nil {
		detail = strings.TrimSpace(string(raw.Detail12.Inner))
	}

	return &FaultError{
		Code:    code,
		Message: message,
		Actor:   actor,
		Detail:  detail,
	}
}
