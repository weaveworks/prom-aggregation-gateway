package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(cfg apiRouterConfig) *gin.Engine {
	agg := newAggregate()
	promConfig := metrics.Config{
		Registry: prometheus.NewRegistry(),
	}
	return setupAPIRouter(cfg, agg, promConfig)
}

func TestHealthCheck(t *testing.T) {
	router := gin.New()
	router.GET("/", handleHealthCheck)

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseHeaders := w.Header()
	assert.Equal(t, "application/json", responseHeaders.Get("content-type"))

	responseData, _ := io.ReadAll(w.Body)
	var response HealthResponse
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)
	assert.Equal(t, true, response.IsAlive)
}

func TestMultiLabelPosting(t *testing.T) {
	tests := []struct {
		name         string
		path, metric string
		expected     string
	}{
		{
			"multiple labels",
			"/metrics/label1/value1/label2/value2",
			`# TYPE some_counter counter
some_counter 1
`,
			`# TYPE some_counter counter
some_counter{label1="value1",label2="value2"} 1
`},
		{
			"job label",
			"/metrics/job/someJob",
			`# TYPE some_counter counter
some_counter 1
`,
			`# TYPE some_counter counter
some_counter{job="someJob"} 1
`,
		},
		{
			"no labels, no trailing slash",
			"/metrics",
			"# TYPE some_counter counter\nsome_counter 1\n",
			"# TYPE some_counter counter\nsome_counter 1\n",
		},
		{
			"no labels, trailing slash",
			"/metrics/",
			"# TYPE some_counter counter\nsome_counter 1\n",
			"# TYPE some_counter counter\nsome_counter 1\n",
		},
		{
			"duplicate labels",
			"/metrics/testing/one/testing/two/testing/three",
			"# TYPE some_counter counter\n some_counter 1\n",
			"# TYPE some_counter counter\nsome_counter{testing=\"one\",testing=\"two\",testing=\"three\"} 1\n",
		},
	}

	for idx, test := range tests {
		t.Run(fmt.Sprintf("test #%d: %s", idx+1, test.name), func(t *testing.T) {
			// setup router
			router := setupTestRouter(apiRouterConfig{corsDomain: "https://cors-domain"})

			// ---- insert metric ----
			// setup request
			buf := bytes.NewBufferString(test.metric)
			req, err := http.NewRequest("PUT", test.path, buf)
			require.NoError(t, err)

			// make request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 202, w.Code)

			// ---- retrieve metric ----
			req, err = http.NewRequest("GET", "/metrics", nil)
			require.NoError(t, err)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			body := w.Body.String()
			assert.Equal(t, test.expected, body)
		})
	}
}

func TestAuthRouter(t *testing.T) {
	tests := []struct {
		name                   string
		path, metric           string
		accounts               gin.Accounts
		authName, authPassword string
		statusCode             int
		expected               string
	}{
		{
			"no labels, no trailing slash",
			"/metrics",
			"# TYPE some_counter counter\nsome_counter 1\n",
			gin.Accounts{"user": "password"},
			"user", "password",
			202,
			"# TYPE some_counter counter\nsome_counter 1\n",
		},
		{
			"no labels, no trailing slash",
			"/metrics",
			"# TYPE some_counter counter\nsome_counter 1\n",
			gin.Accounts{"user": "password"},
			"user1", "password1",
			401,
			"",
		},
	}

	for idx, test := range tests {
		t.Run(fmt.Sprintf("test #%d: %s", idx+1, test.name), func(t *testing.T) {
			// setup router
			router := setupTestRouter(apiRouterConfig{corsDomain: "https://cors-domain", accounts: test.accounts})

			buf := bytes.NewBufferString(test.metric)
			req, err := http.NewRequest("PUT", test.path, buf)
			require.NoError(t, err)

			req.SetBasicAuth(test.authName, test.authPassword)

			// make request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.statusCode, w.Code)

			// ---- retrieve metric ----
			req, err = http.NewRequest("GET", "/metrics", nil)
			require.NoError(t, err)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			body := w.Body.String()
			assert.Equal(t, test.expected, body)

		})
	}
}
