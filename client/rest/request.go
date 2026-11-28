package rest

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type multipartFile struct {
	fieldName string
	fileName  string
	reader    io.Reader
}

// Request representa una petición HTTP individual construida de forma fluida.
type Request struct {
	client          *Client
	ctx             context.Context
	rawURL          string
	headers         http.Header
	queryParams     url.Values
	body            any
	hasBody         bool
	formData        url.Values
	multipartFields map[string]string
	multipartFiles  []multipartFile
	result          any
	errResult       any
}

// SetURL define la URL destino completa de la petición.
func (r *Request) SetURL(rawURL string) *Request {
	r.rawURL = strings.TrimSpace(rawURL)
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

// SetQueryParam agrega o reemplaza un parámetro de query string.
func (r *Request) SetQueryParam(key, value string) *Request {
	if r.queryParams == nil {
		r.queryParams = make(url.Values)
	}
	r.queryParams.Set(key, value)
	return r
}

// SetQueryParams agrega o reemplaza múltiples parámetros de query string.
func (r *Request) SetQueryParams(params map[string]string) *Request {
	if r.queryParams == nil {
		r.queryParams = make(url.Values, len(params))
	}
	for key, value := range params {
		r.queryParams.Set(key, value)
	}
	return r
}

// SetBody define el cuerpo de la petición.
// Soporta []byte, string, io.Reader o cualquier struct/map (serializado automáticamente a JSON).
func (r *Request) SetBody(body any) *Request {
	r.body = body
	r.hasBody = true
	return r
}

// SetFormData configura un cuerpo application/x-www-form-urlencoded.
func (r *Request) SetFormData(data map[string]string) *Request {
	if r.formData == nil {
		r.formData = make(url.Values, len(data))
	}
	for key, value := range data {
		r.formData.Set(key, value)
	}
	return r
}

// SetMultipartFormData agrega campos de texto para una petición multipart/form-data.
func (r *Request) SetMultipartFormData(fields map[string]string) *Request {
	if r.multipartFields == nil {
		r.multipartFields = make(map[string]string, len(fields))
	}
	for key, value := range fields {
		r.multipartFields[key] = value
	}
	return r
}

// SetFile adjunta un archivo para una petición multipart/form-data.
func (r *Request) SetFile(fieldName, fileName string, reader io.Reader) *Request {
	r.multipartFiles = append(r.multipartFiles, multipartFile{
		fieldName: fieldName,
		fileName:  fileName,
		reader:    reader,
	})
	return r
}

// SetResult indica el puntero donde decodificar la respuesta JSON si el status es 2xx.
func (r *Request) SetResult(out any) *Request {
	r.result = out
	return r
}

// SetError indica el puntero donde decodificar la respuesta JSON si el status es >= 400.
func (r *Request) SetError(out any) *Request {
	r.errResult = out
	return r
}

// Get ejecuta una petición HTTP GET.
func (r *Request) Get() (*Response, error) {
	return r.Do(http.MethodGet)
}

// Post ejecuta una petición HTTP POST.
func (r *Request) Post() (*Response, error) {
	return r.Do(http.MethodPost)
}

// Put ejecuta una petición HTTP PUT.
func (r *Request) Put() (*Response, error) {
	return r.Do(http.MethodPut)
}

// Patch ejecuta una petición HTTP PATCH.
func (r *Request) Patch() (*Response, error) {
	return r.Do(http.MethodPatch)
}

// Delete ejecuta una petición HTTP DELETE.
func (r *Request) Delete() (*Response, error) {
	return r.Do(http.MethodDelete)
}

// Do construye y ejecuta la petición con el método HTTP indicado.
func (r *Request) Do(method string) (*Response, error) {
	targetURL, err := r.buildURL()
	if err != nil {
		return nil, err
	}

	bodyReader, defaultContentType, err := r.buildBody()
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(r.ctx, method, targetURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBuildRequest, err)
	}

	r.applyHeaders(httpReq, defaultContentType)

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

	if err := r.decodeResponse(resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func (r *Request) buildURL() (string, error) {
	if r.rawURL == "" {
		return "", ErrEmptyURL
	}

	parsed, err := url.Parse(r.rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidURL
	}

	if len(r.queryParams) > 0 {
		q := parsed.Query()
		for key, values := range r.queryParams {
			for _, value := range values {
				q.Set(key, value)
			}
		}
		parsed.RawQuery = q.Encode()
	}

	return parsed.String(), nil
}

func (r *Request) buildBody() (io.Reader, string, error) {
	hasForm := len(r.formData) > 0
	hasMultipart := len(r.multipartFields) > 0 || len(r.multipartFiles) > 0

	modes := 0
	if r.hasBody {
		modes++
	}
	if hasForm {
		modes++
	}
	if hasMultipart {
		modes++
	}
	if modes > 1 {
		return nil, "", ErrConflictingBody
	}

	if hasForm {
		encoded := r.formData.Encode()
		return strings.NewReader(encoded), "application/x-www-form-urlencoded", nil
	}

	if hasMultipart {
		return r.buildMultipartBody()
	}

	if !r.hasBody || r.body == nil {
		return nil, "", nil
	}

	switch typed := r.body.(type) {
	case []byte:
		return bytes.NewReader(typed), "", nil
	case string:
		return strings.NewReader(typed), "", nil
	case io.Reader:
		return typed, "", nil
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return nil, "", fmt.Errorf("%w: %s", ErrMarshalBody, err)
		}
		return bytes.NewReader(encoded), "application/json", nil
	}
}

func (r *Request) buildMultipartBody() (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, value := range r.multipartFields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", fmt.Errorf("%w: %s", ErrMarshalBody, err)
		}
	}

	for _, file := range r.multipartFiles {
		if file.reader == nil {
			return nil, "", fmt.Errorf("%w: file reader for %q is nil", ErrMarshalBody, file.fieldName)
		}

		part, err := writer.CreateFormFile(file.fieldName, file.fileName)
		if err != nil {
			return nil, "", fmt.Errorf("%w: %s", ErrMarshalBody, err)
		}

		if _, err := io.Copy(part, file.reader); err != nil {
			return nil, "", fmt.Errorf("%w: %s", ErrMarshalBody, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("%w: %s", ErrMarshalBody, err)
	}

	return &buf, writer.FormDataContentType(), nil
}

func (r *Request) applyHeaders(req *http.Request, defaultContentType string) {
	for key, values := range r.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if defaultContentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", defaultContentType)
	}
}

func (r *Request) decodeResponse(resp *Response) error {
	if len(resp.Body) == 0 {
		return nil
	}

	if resp.IsSuccess() && r.result != nil {
		return resp.JSON(r.result)
	}

	if resp.IsError() && r.errResult != nil {
		return resp.JSON(r.errResult)
	}

	return nil
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
