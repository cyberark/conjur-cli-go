package clients

import (
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var httpClient = &http.Client{
	Timeout: 1 * time.Second,
}

var testCases = []struct {
	name          string
	redirectURI   string
	expectedError string
}{
	{
		name:        "Succeeds with port 8888",
		redirectURI: "http://127.0.0.1:8888/callback",
	},
	{
		name:        "Succeeds with random port",
		redirectURI: fmt.Sprintf("http://127.0.0.1:%d/callback", randomPort()),
	},
	{
		name:        "Succeeds with port 80",
		redirectURI: "http://127.0.0.1:80/callback",
	},
	{
		name:        "Default port is used if none is specified",
		redirectURI: "http://127.0.0.1/callback",
	},
	{
		name:        "Succeeds with ipv6 address and port",
		redirectURI: "http://[::1]:8888/callback",
	},
	{
		name:        "Succeeds with ipv6 address and default port",
		redirectURI: "http://[::1]/callback",
	},
	{
		name:        "Succeeds with localhost and port 8888",
		redirectURI: "http://localhost:8888/callback",
	},
	{
		name:        "Succeeds with localhost and default port",
		redirectURI: "http://localhost/callback",
	},
	{
		name:          "Fails if redirect_uri path is wrong",
		redirectURI:   "http://127.0.0.1:9999/incorrect_callback",
		expectedError: "redirect_uri must be http://localhost[:port]/callback or http://127.0.0.1[:port]/callback or http://[::1][:port]/callback",
	},
	{
		name:          "Fails if redirect_uri port is not an integer",
		redirectURI:   "http://127.0.0.1:invalid/callback",
		expectedError: "redirect_uri must be http://localhost[:port]/callback or http://127.0.0.1[:port]/callback or http://[::1][:port]/callback",
	},
	{
		name:          "Fails if redirect_uri port is invalid",
		redirectURI:   "http://127.0.0.1:65536/callback",
		expectedError: "Port in redirect_uri must be between 1 and 65535",
	},
}

func TestHandleOidcFlow(t *testing.T) {
	callbackServerTimeout = 5 * time.Second
	// Note: It may be good to use the httptest package here instead of starting a real server which may allow
	// running the tests in parallel.

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var code string
			var serverError error
			mockOpenBrowser := func(url string) error { return nil }
			redirectURI := "https://example.com?redirect_uri=" + url.QueryEscape(tc.redirectURI)
			go func() {
				code, serverError = handleOpenIDFlow(redirectURI, mockGenerateState, mockOpenBrowser)
			}()
			// Wait for the server to start up asynchronously
			time.Sleep(200 * time.Millisecond)

			// Check that the server is or is not running
			if tc.expectedError != "" {
				assert.False(t, isServerRunning(tc.redirectURI))
				assert.ErrorContains(t, serverError, tc.expectedError)
			} else {
				assert.True(t, isServerRunning(tc.redirectURI))
				_, err := httpClient.Get(tc.redirectURI + "?code=1234&state=test-state")
				assert.NoError(t, err)
				// Wait for the server to stop
				time.Sleep(100 * time.Millisecond)
				assert.False(t, isServerRunning(tc.redirectURI))
				assert.Equal(t, "1234", code)
				assert.NoError(t, serverError)
			}
		})
	}

	t.Run("Opens browser and listens for callback", func(t *testing.T) {
		openBrowserCalled := false

		mockOpenBrowser := func(url string) error {
			openBrowserCalled = true
			return nil
		}

		// Start the local server
		var code string
		var serverError error
		port := fmt.Sprint(randomPort())
		redirectURI := "https://example.com?redirect_uri=http%3A%2F%2F127.0.0.1%3A" + port + "%2Fcallback"
		go func() {
			code, serverError = handleOpenIDFlow(redirectURI, mockGenerateState, mockOpenBrowser)
		}()
		// Wait for the server to start up asynchronously
		time.Sleep(1 * time.Second)
		// Ensure the open browser function was called and the server is running
		assert.True(t, openBrowserCalled)
		assert.True(t, isServerRunning("http://127.0.0.1:"+port))
		// Make a request to the callback endpoint without a code...
		resp, err := httpClient.Get("http://127.0.0.1:" + port + "/callback?state=test-state")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		// ...or with the wrong state...
		resp, err = httpClient.Get("http://127.0.0.1:" + port + "/callback?code=1234&state=wrong-state")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		// ...and finally with a code and the correct state
		resp, err = httpClient.Get("http://127.0.0.1:" + port + "/callback?code=1234&state=test-state")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		content, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(content), "Logged in successfully")
		// Ensure server is shut down
		time.Sleep(1 * time.Second)
		assert.False(t, isServerRunning("http://127.0.0.1:"+port))
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
		port := fmt.Sprint(randomPort())
		listener, err := net.Listen("tcp", "127.0.0.1:"+port)
		assert.NoError(t, err)
		defer listener.Close()

		// Now try to start the local server
		_, err = handleOpenIDFlow("https://example.com?redirect_uri=http%3A%2F%2F127.0.0.1%3A"+port+"%2Fcallback", mockGenerateState, mockOpenBrowser)

		assert.ErrorContains(t, err, "address already in use")
		assert.False(t, openBrowserCalled)
	})

	t.Run("Stops server after timeout", func(t *testing.T) {
		t.Cleanup(func() {
			// Restore the timeout
			callbackServerTimeout = 5 * time.Second
		})

		openBrowserCalled := false

		mockOpenBrowser := func(url string) error {
			openBrowserCalled = true
			return nil
		}

		// Decrease the timeout
		callbackServerTimeout = 1 * time.Second
		// Start the local server
		port := fmt.Sprint(randomPort())
		redirectURI := "https://example.com?redirect_uri=http%3A%2F%2F127.0.0.1%3A" + port + "%2Fcallback"
		go func() {
			handleOpenIDFlow(redirectURI, mockGenerateState, mockOpenBrowser)
		}()

		// Wait for the server to start up asynchronously
		time.Sleep(1 * time.Second)
		// Ensure the open browser function was called
		assert.True(t, openBrowserCalled)
		// Ensure server is shut down
		time.Sleep(250 * time.Millisecond)
		assert.False(t, isServerRunning("http://127.0.0.1:"+port))
	})
}

func TestGenerateState(t *testing.T) {
	state := generateState()
	// Check that the state is a valid base64 string
	decodedState, err := base64.StdEncoding.DecodeString(state)
	assert.NoError(t, err)
	assert.Len(t, decodedState, 16)
}

func isServerRunning(uri string) bool {
	// Checks whether the server is running by attempting to connect to the host and port
	u, err := url.Parse(uri)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// If host is missing the port number append the default (:80)
	if _, _, err := net.SplitHostPort(u.Host); err != nil {
		u.Host = fmt.Sprintf("%s:80", u.Host)
	}

	conn, err := net.DialTimeout("tcp", u.Host, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func mockGenerateState() string {
	return "test-state"
}

func randomPort() int {
	return rand.Intn(65535-1024) + 1024
}
