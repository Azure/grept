package pkg

import (
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/net/http/httpproxy"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var _ Data = &HttpDatasource{}

type HttpDatasource struct {
	*BaseData
	Url             string
	Method          string
	RequestBody     string
	RequestHeaders  map[string]string
	ResponseBody    string
	ResponseHeaders map[string]string
	StatusCode      int
}

func (h *HttpDatasource) Load() error {
	tr, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return fmt.Errorf("error http: can't configure http transport")
	}
	clonedTr := tr.Clone()

	// Prevent issues with tests caching the proxy configuration.
	clonedTr.Proxy = func(req *http.Request) (*url.URL, error) {
		return httpproxy.FromEnvironment().ProxyFunc()(req.URL)
	}

	if clonedTr.TLSClientConfig == nil {
		clonedTr.TLSClientConfig = &tls.Config{}
	}

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = clonedTr
	request, err := retryablehttp.NewRequestWithContext(h.Context(), h.Method, h.Url, strings.NewReader(h.RequestBody))
	if err != nil {
		return fmt.Errorf("error creating request data.http.%s, %s", h.name, err.Error())
	}
	for k, v := range h.RequestHeaders {
		request.Header.Set(k, v)
	}
	response, err := retryClient.Do(request)
	if err != nil {
		return fmt.Errorf("error making request data.http.%s, detail: %s", h.name, err.Error())
	}
	defer func() { _ = response.Body.Close() }()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body data.http.%s, detail: %s", h.name, err.Error())
	}
	h.ResponseBody = string(bytes)
	h.ResponseHeaders = make(map[string]string)
	for k, v := range response.Header {
		// Concatenate according to RFC9110 https://www.rfc-editor.org/rfc/rfc9110.html#section-5.2
		h.ResponseHeaders[k] = strings.Join(v, ", ")
	}
	h.StatusCode = response.StatusCode
	return nil
}

func (h *HttpDatasource) Type() string {
	return "http"
}

func (h *HttpDatasource) Value() cty.Value {
	attrs := h.BaseValue()
	attrs["url"] = cty.StringVal(h.Url)
	attrs["method"] = cty.StringVal(h.Method)
	attrs["request_body"] = cty.StringVal(h.RequestBody)
	attrs["response_body"] = cty.StringVal(h.ResponseBody)
	attrs["status_code"] = cty.NumberIntVal(int64(h.StatusCode))
	attrs["request_headers"] = h.HeaderValue(h.RequestHeaders)
	attrs["response_headers"] = h.HeaderValue(h.ResponseHeaders)
	return cty.ObjectVal(attrs)
}

func (h *HttpDatasource) HeaderValue(headers map[string]string) cty.Value {
	if len(headers) == 0 {
		return cty.MapValEmpty(cty.String)
	}
	inner := make(map[string]cty.Value, 0)
	for k, v := range headers {
		inner[k] = cty.StringVal(v)
	}
	return cty.MapVal(inner)
}

func (h *HttpDatasource) Parse(b *hclsyntax.Block) error {
	var err error
	if err = h.BaseData.Parse(b); err != nil {
		return err
	}
	if h.Url, err = readRequiredStringAttribute(b, "url", h.EvalContext()); err != nil {
		return err
	}
	if h.Method, err = readOptionalStringAttribute(b, "method", h.EvalContext()); err != nil {
		return err
	}
	if h.Method == "" {
		h.Method = "GET"
	}
	if h.RequestBody, err = readOptionalStringAttribute(b, "request_body", h.EvalContext()); err != nil {
		return err
	}
	if h.RequestHeaders, err = readOptionalMapAttribute(b, "request_headers", h.EvalContext()); err != nil {
		return err
	}
	return nil
}
