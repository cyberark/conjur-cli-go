package clients

import (
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var httpClient = &http.Client{
	Timeout: 1 * time.Second,
}

func TestHandleOidcFlow(t *testing.T) {
	// Note: It may be good to use the httptest package here instead of starting a real server which may allow
	// running the tests in parallel.

	t.Run("Fails if redirect_uri is wrong", func(t *testing.T) {
		_, err := handleOpenIDFlow("https://example.com?redirect_uri=http%3A%2F%2Flocalhost%3A9999%2Fcallback", generateState, openBrowser)
		assert.ErrorContains(t, err, "redirect_uri must be http://localhost:8888/callback")
		assert.False(t, isServerRunning())
	})

	t.Run("Opens browser and listens for callback", func(t *testing.T) {
		openBrowserCalled := false

		mockOpenBrowser := func(url string) error {
			openBrowserCalled = true
			return nil
		}

		mockGenerateState := func() string {
			return "test-state"
		}

		// Start the local server
		http.DefaultServeMux = http.NewServeMux()
		var code string
		var serverError error
		go func() {
			code, serverError = handleOpenIDFlow("https://example.com?redirect_uri=http%3A%2F%2Flocalhost%3A8888%2Fcallback", mockGenerateState, mockOpenBrowser)
		}()
		// Wait for the server to start up asynchronously
		time.Sleep(1 * time.Second)
		// Ensure the open browser function was called and the server is running
		assert.True(t, openBrowserCalled)
		assert.True(t, isServerRunning())
		// Make a request to the callback endpoint with a code
		resp, err := httpClient.Get("http://localhost:8888/callback?code=1234&state=test-state")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		content, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(content), "Logged in successfully")
		// Ensure server is shut down
		time.Sleep(1 * time.Second)
		assert.False(t, isServerRunning())
		// Ensure the code was returned
		assert.Equal(t, "1234", code)
		assert.NoError(t, serverError)
	})
}

func isServerRunning() bool {
	// Checks whether the server is running by attempting to connect to the port
	conn, err := net.DialTimeout("tcp", "localhost:8888", 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
