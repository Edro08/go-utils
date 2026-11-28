package soap

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type outgoingEnvelope struct {
	XMLName xml.Name
	Header  *outgoingHeader `xml:"soap:Header,omitempty"`
	Body    outgoingBody    `xml:"soap:Body"`
	XMLNS   string          `xml:"xmlns:soap,attr"`
}

type outgoingHeader struct {
	Content []byte `xml:",innerxml"`
}

type outgoingBody struct {
	Content []byte `xml:",innerxml"`
}

// Request representa una petición SOAP individual construida de forma fluida.
type Request struct {
	client      *Client
	ctx         context.Context
	rawURL      string
	action      string
	version     Version
	headers     http.Header
	soapHeaders []any
	body        any
	rawEnvelope []byte
	result      any
}

// SetURL define la URL del endpoint SOAP.
func (r *Request) SetURL(rawURL string) *Request {
	r.rawURL = strings.TrimSpace(rawURL)
	return r
}

// SetAction define el valor de SOAPAction (SOAP 1.1) o el parámetro action del Content-Type (SOAP 1.2).
func (r *Request) SetAction(action string) *Request {
	r.action = strings.TrimSpace(action)
	return r
}

// SetVersion sobrescribe la versión SOAP (1.1 o 1.2) únicamente para esta petición.
func (r *Request) SetVersion(version Version) *Request {
	if version == Version11 || version == Version12 {
		r.version = version
	}
	return r
}

// SetHeader establece un header HTTP sobrescribiendo cualquier valor previo con la misma clave.
func (r *Request) SetHeader(key, value string) *Request {
	r.headers.Set(key, value)
	return r
}

// SetHeaders establece múltiples headers HTTP.
func (r *Request) SetHeaders(headers map[string]string) *Request {
	for key, value := range headers {
		r.headers.Set(key, value)
	}
	return r
}

// SetBearerToken establece el header Authorization con esquema Bearer.
func (r *Request) SetBearerToken(token string) *Request {
	r.headers.Set("Authorization", "Bearer "+token)
	return r
}

// SetBasicAuth establece el header Authorization con esquema Basic.
func (r *Request) SetBasicAuth(username, password string) *Request {
	credentials := username + ":" + password
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	r.headers.Set("Authorization", "Basic "+encoded)
	return r
}

// SetSOAPHeader define un único bloque opcional dentro de <soap:Header> (struct XML, []byte o string).
func (r *Request) SetSOAPHeader(header any) *Request {
	if header == nil {
		r.soapHeaders = nil
		return r
	}
	r.soapHeaders = []any{header}
	return r
}

// SetSOAPHeaders define múltiples bloques dentro de <soap:Header> (structs XML, []byte o string).
func (r *Request) SetSOAPHeaders(headers ...any) *Request {
	r.soapHeaders = nil
	for _, header := range headers {
		if header != nil {
			r.soapHeaders = append(r.soapHeaders, header)
		}
	}
	return r
}

// SetBody define el payload interno de <soap:Body> (struct XML, []byte o string).
func (r *Request) SetBody(body any) *Request {
	r.body = body
	return r
}

// SetRawEnvelope permite enviar un documento <Envelope> XML completo sin envolverlo automáticamente.
func (r *Request) SetRawEnvelope(xmlData []byte) *Request {
	r.rawEnvelope = xmlData
	return r
}

// SetResult indica el puntero donde decodificar el contenido de <soap:Body> si la respuesta no contiene Fault.
func (r *Request) SetResult(out any) *Request {
	r.result = out
	return r
}

// Send construye el sobre SOAP, envía la petición POST y procesa la respuesta o Fault.
func (r *Request) Send() (*Response, error) {
	targetURL, err := r.validateURL()
	if err != nil {
		return nil, err
	}

	payload, err := r.buildEnvelopeBytes()
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(r.ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBuildRequest, err)
	}

	r.applyHeaders(httpReq)

	start := time.Now()
	httpResp, err := r.client.httpClient.Do(httpReq)
	duration := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRequestFailed, err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	bodyBytes, err := readLimitedBody(httpResp.Body, r.client.maxBodySize)
	if err != nil {
		return nil, err
	}

	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    httpResp.Header.Clone(),
		Body:       bodyBytes,
		Duration:   duration,
	}

	if len(bodyBytes) == 0 {
		return resp, nil
	}

	bodyContent, fault, err := parseEnvelopeResponse(bodyBytes)
	if err != nil {
		return resp, err
	}

	resp.BodyContent = bodyContent
	resp.Fault = fault

	if fault != nil {
		return resp, fault
	}

	if resp.IsSuccess() && r.result != nil {
		if err := resp.UnmarshalBody(r.result); err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func (r *Request) validateURL() (string, error) {
	if r.rawURL == "" {
		return "", ErrEmptyURL
	}

	parsed, err := url.Parse(r.rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidURL
	}

	return parsed.String(), nil
}

func (r *Request) buildEnvelopeBytes() ([]byte, error) {
	if len(r.rawEnvelope) > 0 {
		return r.rawEnvelope, nil
	}

	if r.body == nil {
		return nil, ErrEmptyBody
	}

	bodyXML, err := encodeXMLFragment(r.body)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMarshalEnvelope, err)
	}

	ns := namespace11
	if r.version == Version12 {
		ns = namespace12
	}

	env := outgoingEnvelope{
		XMLName: xml.Name{Local: "soap:Envelope"},
		XMLNS:   ns,
		Body:    outgoingBody{Content: bodyXML},
	}

	if len(r.soapHeaders) > 0 {
		var headerBuf bytes.Buffer
		for _, header := range r.soapHeaders {
			headerXML, err := encodeXMLFragment(header)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", ErrMarshalEnvelope, err)
			}
			headerBuf.Write(headerXML)
		}
		env.Header = &outgoingHeader{Content: headerBuf.Bytes()}
	}

	encoded, err := xml.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMarshalEnvelope, err)
	}

	return append([]byte(xml.Header), encoded...), nil
}

func (r *Request) applyHeaders(req *http.Request) {
	for key, values := range r.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if req.Header.Get("Content-Type") == "" {
		if r.version == Version12 {
			contentType := "application/soap+xml; charset=utf-8"
			if r.action != "" {
				contentType += fmt.Sprintf(`; action="%s"`, r.action)
			}
			req.Header.Set("Content-Type", contentType)
		} else {
			req.Header.Set("Content-Type", "text/xml; charset=utf-8")
		}
	}

	if r.version == Version11 && r.action != "" && req.Header.Get("SOAPAction") == "" {
		req.Header.Set("SOAPAction", fmt.Sprintf(`"%s"`, r.action))
	}
}

func encodeXMLFragment(v any) ([]byte, error) {
	switch typed := v.(type) {
	case []byte:
		return typed, nil
	case string:
		return []byte(typed), nil
	default:
		return xml.Marshal(typed)
	}
}

func readLimitedBody(body io.Reader, maxSize int64) ([]byte, error) {
	limited := io.LimitReader(body, maxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrReadBody, err)
	}

	if int64(len(data)) > maxSize {
		return nil, ErrBodyTooLarge
	}

	return data, nil
}
