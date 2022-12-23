package clients

import (
	"context"
	"fmt"
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
		_, err := handleOpenIDFlow("https://example.com?redirect_uri=http%3A%2F%2F127.0.0.1%3A9999%2Fcallback", generateState, openBrowser)
		assert.ErrorContains(t, err, "redirect_uri must be http://127.0.0.1:8888/callback")
		assert.False(t, isServerRunning())
	})

	t.Run("Opens browser and listens for callback", func(t *testing.T) {
		openBrowserCalled := false

		mockOpenBrowser := func(url string) error {
			openBrowserCalled = true
			return nil
		}

		// Start the local server
		http.DefaultServeMux = http.NewServeMux()
		var code string
		var serverError error
		go func() {
			code, serverError = handleOpenIDFlow("https://example.com?redirect_uri=http%3A%2F%2F127.0.0.1%3A8888%2Fcallback", mockGenerateState, mockOpenBrowser)
		}()
		// Wait for the server to start up asynchronously
		time.Sleep(1 * time.Second)
		// Ensure the open browser function was called and the server is running
		assert.True(t, openBrowserCalled)
		assert.True(t, isServerRunning())
		// Make a request to the callback endpoint without a code...
		resp, err := httpClient.Get("http://127.0.0.1:8888/callback?state=test-state")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		// ...or with the wrong state...
		resp, err = httpClient.Get("http://127.0.0.1:8888/callback?code=1234&state=wrong-state")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		// ...and finally with a code and the correct state
		resp, err = httpClient.Get("http://127.0.0.1:8888/callback?code=1234&state=test-state")
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

	t.Run("Doesn't open browser if server can't start", func(t *testing.T) {
		openBrowserCalled := false

		mockOpenBrowser := func(url string) error {
			openBrowserCalled = true
			return nil
		}

		// First start a dummy server to occupy the port
		stopDummyServer := runDummyServer()
		defer stopDummyServer()

		// Now try to start the local server
		go func() {
			assert.Panics(t, func() {
				handleOpenIDFlow("https://example.com?redirect_uri=http%3A%2F%2F127.0.0.1%3A8888%2Fcallback", mockGenerateState, mockOpenBrowser)
			})
		}()

		// Wait for the server to (attempt to) start up asynchronously
		time.Sleep(1 * time.Second)

		assert.False(t, openBrowserCalled)
	})

	t.Run("Stops server after timeout", func(t *testing.T) {
		t.Cleanup(func() {
			// Restore the timeout
			callbackServerTimeout = 5 * time.Minute
		})

		openBrowserCalled := false

		mockOpenBrowser := func(url string) error {
			openBrowserCalled = true
			return nil
		}

		// Decrease the timeout
		callbackServerTimeout = 1 * time.Second
		// Start the local server
		http.DefaultServeMux = http.NewServeMux()
		go func() {
			handleOpenIDFlow("https://example.com?redirect_uri=http%3A%2F%2F127.0.0.1%3A8888%2Fcallback", mockGenerateState, mockOpenBrowser)
		}()

		// Wait for the server to start up asynchronously
		time.Sleep(1 * time.Second)
		// Ensure the open browser function was called
		assert.True(t, openBrowserCalled)
		// Ensure server is shut down
		time.Sleep(100 * time.Millisecond)
		assert.False(t, isServerRunning())
	})
}

func isServerRunning() bool {
	// Checks whether the server is running by attempting to connect to the port
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8888", 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func runDummyServer() func() {
	http.DefaultServeMux = http.NewServeMux()
	server := &http.Server{Addr: "127.0.0.1:8888"}
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintln(w, "This should never be called")
	})
	http.Handle("/callback", server.Handler)
	go func() {
		server.ListenAndServe()
	}()
	return func() {
		server.Shutdown(context.Background())
	}
}

func mockGenerateState() string {
	return "test-state"
}
