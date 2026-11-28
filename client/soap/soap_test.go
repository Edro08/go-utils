package soap

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/edro08/go-utils/logger"
)

type authHeader struct {
	XMLName xml.Name `xml:"AuthHeader"`
	Token   string   `xml:"Token"`
}

type traceHeader struct {
	XMLName xml.Name `xml:"TraceHeader"`
	TraceID string   `xml:"TraceID"`
}

type balanceRequest struct {
	XMLName   xml.Name `xml:"GetBalanceRequest"`
	AccountID string   `xml:"AccountID"`
}

type balanceResponse struct {
	XMLName xml.Name `xml:"GetBalanceResponse"`
	Amount  string   `xml:"Amount"`
}

func TestClient_SOAP11AndSOAP12(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyStr := string(bodyBytes)

		switch r.URL.Path {
		case "/v11":
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/xml") {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
			if r.Header.Get("SOAPAction") != `"http://tempuri.org/GetBalance"` {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if !strings.Contains(bodyStr, "<AuthHeader><Token>tok-1</Token></AuthHeader>") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if !strings.Contains(bodyStr, "<TraceHeader><TraceID>tr-1</TraceID></TraceHeader>") {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if !strings.Contains(bodyStr, "<AccountID>ACC-42</AccountID>") {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetBalanceResponse><Amount>150.50</Amount></GetBalanceResponse>
  </soap:Body>
</soap:Envelope>`))

		case "/v12":
			ct := r.Header.Get("Content-Type")
			if !strings.Contains(ct, "application/soap+xml") || !strings.Contains(ct, `action="http://tempuri.org/GetBalance"`) {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
			if !strings.Contains(bodyStr, namespace12) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<env:Envelope xmlns:env="http://www.w3.org/2003/05/soap-envelope">
  <env:Body>
    <GetBalanceResponse><Amount>999.00</Amount></GetBalanceResponse>
  </env:Body>
</env:Envelope>`))

		case "/fault11":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <soap:Fault>
      <faultcode>soap:Client</faultcode>
      <faultstring>Invalid account</faultstring>
      <faultactor>billing</faultactor>
      <detail><errCode>E100</errCode></detail>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`))

		case "/fault12":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<env:Envelope xmlns:env="http://www.w3.org/2003/05/soap-envelope">
  <env:Body>
    <env:Fault>
      <env:Code><env:Value>env:Sender</env:Value></env:Code>
      <env:Reason><env:Text>Missing header</env:Text></env:Reason>
      <env:Detail><info>auth required</info></env:Detail>
    </env:Fault>
  </env:Body>
</env:Envelope>`))
		}
	}))
	defer server.Close()

	var client IClient = New(Options{
		Version: Version11,
		Timeout: 5 * time.Second,
	})

	t.Run("SOAP 1.1 with multiple SOAPHeaders, SOAPAction and SetResult", func(t *testing.T) {
		var out balanceResponse
		resp, err := client.R(t.Context()).
			SetURL(server.URL + "/v11").
			SetAction("http://tempuri.org/GetBalance").
			SetSOAPHeaders(
				authHeader{Token: "tok-1"},
				traceHeader{TraceID: "tr-1"},
			).
			SetBody(balanceRequest{AccountID: "ACC-42"}).
			SetResult(&out).
			Send()

		if err != nil {
			t.Fatalf("Send() error = %v", err)
		}
		if !resp.IsSuccess() {
			t.Errorf("expected IsSuccess=true, status=%d", resp.StatusCode)
		}
		if out.Amount != "150.50" {
			t.Errorf("Amount = %q, want %q", out.Amount, "150.50")
		}
	})

	t.Run("SOAP 1.2 via SetVersion override and raw string body", func(t *testing.T) {
		var out balanceResponse
		resp, err := client.R(t.Context()).
			SetURL(server.URL + "/v12").
			SetVersion(Version12).
			SetAction("http://tempuri.org/GetBalance").
			SetBody("<GetBalanceRequest><AccountID>ACC-1</AccountID></GetBalanceRequest>").
			SetResult(&out).
			Send()

		if err != nil {
			t.Fatalf("Send(v12) error = %v", err)
		}
		if !resp.IsSuccess() || out.Amount != "999.00" {
			t.Errorf("unexpected v12 response: success=%v amount=%q", resp.IsSuccess(), out.Amount)
		}
	})

	t.Run("SOAP 1.1 Fault returns FaultError matching ErrFault", func(t *testing.T) {
		resp, err := client.R(t.Context()).
			SetURL(server.URL + "/fault11").
			SetBody(balanceRequest{AccountID: "BAD"}).
			Send()

		if !errors.Is(err, ErrFault) {
			t.Fatalf("error = %v, want %v", err, ErrFault)
		}
		var faultErr *FaultError
		if !errors.As(err, &faultErr) {
			t.Fatalf("expected *FaultError, got %T", err)
		}
		if faultErr.Code != "soap:Client" || faultErr.Message != "Invalid account" || faultErr.Actor != "billing" {
			t.Errorf("unexpected fault fields: %+v", faultErr)
		}
		if !strings.Contains(faultErr.Detail, "E100") {
			t.Errorf("fault detail = %q, want E100", faultErr.Detail)
		}
		if resp.IsSuccess() {
			t.Error("expected resp.IsSuccess() to be false on SOAP fault")
		}
	})

	t.Run("SOAP 1.2 Fault even with HTTP 200 returns FaultError", func(t *testing.T) {
		_, err := client.R(t.Context()).
			SetURL(server.URL + "/fault12").
			SetVersion(Version12).
			SetBody([]byte("<Ping/>")).
			Send()

		if !errors.Is(err, ErrFault) {
			t.Fatalf("error = %v, want %v", err, ErrFault)
		}
		var faultErr *FaultError
		if !errors.As(err, &faultErr) {
			t.Fatalf("expected *FaultError, got %T", err)
		}
		if faultErr.Code != "env:Sender" || faultErr.Message != "Missing header" {
			t.Errorf("unexpected fault 1.2 fields: %+v", faultErr)
		}
	})
}

func TestClient_RawEnvelopeMiddlewaresLoggerAndErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/raw":
			if r.Header.Get("X-MW") != "applied" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Ok/></soap:Body></soap:Envelope>`))
		case "/large":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("0123456789ABCDEF"))
		case "/bad-xml":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<not-closed"))
		}
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	type ctxKey string
	const traceKey ctxKey = "traceID"

	log, err := logger.New(logger.Options{
		MinLevel: logger.INFO,
		Format:   logger.FormatText,
		Writer:   &logBuf,
		ContextExtractor: func(ctx context.Context) []slog.Attr {
			if v, ok := ctx.Value(traceKey).(string); ok {
				return []slog.Attr{slog.String("traceID", v)}
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	addHeaderMW := func(next http.RoundTripper) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			cloned := req.Clone(req.Context())
			cloned.Header.Set("X-MW", "applied")
			return next.RoundTrip(cloned)
		})
	}

	client := New(Options{
		Timeout:     5 * time.Second,
		Logger:      log,
		Middlewares: []Middleware{addHeaderMW},
	})

	t.Run("SetRawEnvelope with Middleware and Logger", func(t *testing.T) {
		ctx := context.WithValue(t.Context(), traceKey, "tr-soap-1")
		raw := []byte(`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Ping/></soap:Body></soap:Envelope>`)

		resp, err := client.R(ctx).
			SetURL(server.URL+"/raw").
			SetBearerToken("abc").
			SetHeader("X-Custom", "1").
			SetHeaders(map[string]string{"X-Two": "2"}).
			SetRawEnvelope(raw).
			Send()

		if err != nil {
			t.Fatalf("Send(raw) error = %v", err)
		}
		if !resp.IsSuccess() || !strings.Contains(resp.String(), "<Ok/>") {
			t.Errorf("unexpected raw response: %q", resp.String())
		}

		logged := logBuf.String()
		for _, want := range []string{"SOAP_CLIENT", "method=POST", "path=/raw", "status=200", "traceID=tr-soap-1"} {
			if !strings.Contains(logged, want) {
				t.Errorf("log output %q missing %q", logged, want)
			}
		}
	})

	t.Run("Validation errors", func(t *testing.T) {
		if _, err := client.R(t.Context()).Send(); !errors.Is(err, ErrEmptyURL) {
			t.Errorf("empty URL error = %v, want %v", err, ErrEmptyURL)
		}

		if _, err := client.R(t.Context()).SetURL("/only-path").Send(); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("invalid URL error = %v, want %v", err, ErrInvalidURL)
		}

		if _, err := client.R(t.Context()).SetURL(server.URL + "/raw").Send(); !errors.Is(err, ErrEmptyBody) {
			t.Errorf("empty body error = %v, want %v", err, ErrEmptyBody)
		}

		smallClient := New(Options{MaxBodySize: 8})
		if _, err := smallClient.R(t.Context()).SetURL(server.URL + "/large").SetBody("<P/>").Send(); !errors.Is(err, ErrBodyTooLarge) {
			t.Errorf("large body error = %v, want %v", err, ErrBodyTooLarge)
		}

		if _, err := client.R(t.Context()).SetURL(server.URL + "/bad-xml").SetBody("<P/>").Send(); !errors.Is(err, ErrUnmarshalResponse) {
			t.Errorf("bad xml error = %v, want %v", err, ErrUnmarshalResponse)
		}
	})
}
