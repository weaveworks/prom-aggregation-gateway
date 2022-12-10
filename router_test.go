package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
