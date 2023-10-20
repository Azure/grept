package pkg

import (
	"crypto/tls"
	"fmt"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/net/http/httpproxy"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var _ Data = &HttpDatasource{}

type HttpDatasource struct {
	*BaseData
	Url             string            `hcl:"url"`
	Method          string            `hcl:"method,optional"`
	RequestBody     string            `hcl:"request_body,optional"`
	RequestHeaders  map[string]string `hcl:"request_headers,optional"`
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
	retryClient.Logger = log.New(os.Stderr, fmt.Sprintf("data.http.%s:", h.name), log.LstdFlags)
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
	diag := gohcl.DecodeBody(b.Body, h.EvalContext(), h)
	if diag.HasErrors() {
		return diag
	}
	if h.Method == "" {
		h.Method = "GET"
	}
	return nil
}

var validHttpMethods = hashset.New("GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH")

func (h *HttpDatasource) Validate() error {
	var err error
	if !validHttpMethods.Contains(h.Method) {
		err = multierror.Append(err, fmt.Errorf(`"method"" must be one of "GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"`))
	}
	return err
}
