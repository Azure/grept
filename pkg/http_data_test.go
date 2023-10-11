package pkg

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpDatasource_Load(t *testing.T) {
	// Create a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	// Create a HttpDatasource instance
	h := &HttpDatasource{
		BaseData: &BaseData{
			name: "test",
		},
		Url:    ts.URL,
		Method: "GET",
	}

	err := h.Load(context.TODO())

	// Assert no error from Load function
	require.NoError(t, err)
	// Assert the response body
	assert.Equal(t, "Hello, client\n", h.ResponseBody)
	assert.Equal(t, 200, h.StatusCode)
	assert.Equal(t, ts.URL, h.Url)
	assert.True(t, len(h.ResponseHeaders) > 1)
	ct, ok := h.ResponseHeaders["Content-Type"]
	assert.True(t, ok)
	assert.Equal(t, "text/plain", ct)
}
