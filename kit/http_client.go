package kit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net"
	"net/http"
	neturl "net/url"
	"time"

	mwinfra "github.com/FourWD/middleware/infra"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type HTTPStatusCode int

type HTTPResponse struct {
	StatusCode HTTPStatusCode
	Headers    http.Header
	Body       io.ReadCloser
}

type HTTPBody interface {
	ToReader() (io.Reader, error)
}

type HTTPRequestParams struct {
	Query   Maybe[map[string]string]
	Headers Maybe[http.Header]
	Body    Maybe[HTTPBody]
}

type HTTPClientOption func(*HTTPClientOptions)

type HTTPClientOptions struct{}

type HTTPRequestContext interface {
	Header() http.Header
	Query() neturl.Values
}

type HTTPRequestInterceptor func(ctx context.Context, request HTTPRequestContext) error

type HTTPClient interface {
	Get(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error)
	Post(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error)
	Put(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error)
	Patch(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error)
	Delete(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error)
}

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type StdHTTPClient struct {
	tracer       trace.Tracer
	client       doer
	interceptors []HTTPRequestInterceptor
}

type httpRequestContext struct {
	header http.Header
	query  neturl.Values
}

type jsonHTTPBody struct {
	value any
}

type readerHTTPBody struct {
	reader io.Reader
}

var ErrBodyIsNotSupported = errors.New("body is not supported")

func NewStdHTTPClient(interceptors ...HTTPRequestInterceptor) *StdHTTPClient {
	return NewCustomStdHTTPClient(http.DefaultClient, interceptors...)
}

func NewCustomStdHTTPClient(client doer, interceptors ...HTTPRequestInterceptor) *StdHTTPClient {
	return &StdHTTPClient{
		tracer:       otel.Tracer("middleware.kit.outbound_http"),
		client:       client,
		interceptors: interceptors,
	}
}

func JSONBody(value any) HTTPBody {
	return jsonHTTPBody{value: value}
}

func ReaderBody(reader io.Reader) HTTPBody {
	return readerHTTPBody{reader: reader}
}

func (b jsonHTTPBody) ToReader() (io.Reader, error) {
	data, err := json.Marshal(b.value)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (b readerHTTPBody) ToReader() (io.Reader, error) {
	return b.reader, nil
}

func (r *httpRequestContext) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

func (r *httpRequestContext) Query() neturl.Values {
	if r.query == nil {
		r.query = make(neturl.Values)
	}
	return r.query
}

func (c *StdHTTPClient) Get(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error) {
	return mwinfra.TraceResult(c.tracer, ctx, "outbound_http.get", func(ctx context.Context, _ trace.Span) (HTTPResponse, error) {
		return c.request(ctx, http.MethodGet, url, params, options...)
	})
}

func (c *StdHTTPClient) Post(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error) {
	return mwinfra.TraceResult(c.tracer, ctx, "outbound_http.post", func(ctx context.Context, _ trace.Span) (HTTPResponse, error) {
		return c.request(ctx, http.MethodPost, url, params, options...)
	})
}

func (c *StdHTTPClient) Put(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error) {
	return mwinfra.TraceResult(c.tracer, ctx, "outbound_http.put", func(ctx context.Context, _ trace.Span) (HTTPResponse, error) {
		return c.request(ctx, http.MethodPut, url, params, options...)
	})
}

func (c *StdHTTPClient) Patch(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error) {
	return mwinfra.TraceResult(c.tracer, ctx, "outbound_http.patch", func(ctx context.Context, _ trace.Span) (HTTPResponse, error) {
		return c.request(ctx, http.MethodPatch, url, params, options...)
	})
}

func (c *StdHTTPClient) Delete(ctx context.Context, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error) {
	return mwinfra.TraceResult(c.tracer, ctx, "outbound_http.delete", func(ctx context.Context, _ trace.Span) (HTTPResponse, error) {
		return c.request(ctx, http.MethodDelete, url, params, options...)
	})
}

func (c *StdHTTPClient) request(ctx context.Context, method, url string, params HTTPRequestParams, options ...HTTPClientOption) (HTTPResponse, error) {
	var body io.Reader
	if payload, ok := params.Body.TryGetValue(); ok {
		reader, err := payload.ToReader()
		if err != nil {
			return HTTPResponse{}, err
		}
		body = reader
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return HTTPResponse{}, err
	}

	queries := make(neturl.Values)
	if query, ok := params.Query.TryGetValue(); ok {
		for key, value := range query {
			queries.Set(key, value)
		}
	}

	if headers, ok := params.Headers.TryGetValue(); ok {
		req.Header = headers.Clone()
	} else {
		req.Header = make(http.Header)
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	if len(c.interceptors) > 0 {
		requestContext := &httpRequestContext{}
		for _, interceptor := range c.interceptors {
			if err := interceptor(ctx, requestContext); err != nil {
				return HTTPResponse{}, err
			}
		}
		if requestContext.header != nil {
			maps.Copy(req.Header, requestContext.header)
		}
		if requestContext.query != nil {
			maps.Copy(queries, requestContext.query)
		}
	}

	if len(queries) > 0 {
		req.URL.RawQuery = queries.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return HTTPResponse{}, err
	}

	return HTTPResponse{
		StatusCode: HTTPStatusCode(resp.StatusCode),
		Headers:    resp.Header,
		Body:       resp.Body,
	}, nil
}

// --- Legacy HTTP helpers migrated from common ---

func CallUrl(url string) string {
	response, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return ""
	}

	return string(body)
}

// DefaultHTTPTimeout is the default timeout for the legacy shared HTTP client.
const DefaultHTTPTimeout = 30 * time.Second

// legacyHTTPClient is a shared HTTP client with timeout configured.
var legacyHTTPClient = &http.Client{
	Timeout: DefaultHTTPTimeout,
}

func HttpRequest(url string, method string, token string, jsonString string) (string, error) {
	jsonByte, err := json.Marshal(jsonString)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonByte))
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	response, err := legacyHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func NewHttpClient(timeoutSec int) *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	return &http.Client{
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: time.Duration(timeoutSec) * time.Second,
	}
}
