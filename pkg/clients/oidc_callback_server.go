package clients

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/browser"
)

var callbackServerTimeout = 5 * time.Minute

type callbackEndpoint struct {
	server         *http.Server
	shutdownSignal chan struct{}
	code           string
	state          string
}

// openBrowser attempts to open a web browser to the given URL
func openBrowser(url string) error {
	err := browser.OpenURL(url)
	if err != nil {
		return fmt.Errorf("Error opening browser")
	}
	return nil
}

func (h *callbackEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Verify request is coming from local machine
	if !strings.HasPrefix(r.RemoteAddr, "[::1]:") && !strings.HasPrefix(r.RemoteAddr, "127.0.0.1:") {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Failed to log in. You may close the browser and try again.")
		return
	}

	// Check state and code. Ignore request if state doesn't match or code is missing.
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state != h.state || code == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Failed to log in. You may close the browser and try again.")
		// Don't shut down the server so the user can try again. This prevents DoS attacks by
		// flooding the 127.0.0.1:8888/callback endpoint with requests.
		return
	}

	h.code = code
	fmt.Fprintln(w, "Logged in successfully. You may close the browser and return to the command line.")
	h.shutdownSignal <- struct{}{}
}

func handleOpenIDFlow(authEndpointURL string, generateStateFn func() string, openBrowserFn func(string) error) (string, error) {
	callbackEndpoint := &callbackEndpoint{}
	callbackEndpoint.state = generateStateFn()
	callbackEndpoint.shutdownSignal = make(chan struct{})

	authURL, authURLParseError := url.Parse(authEndpointURL)
	if authURLParseError != nil {
		return "", authURLParseError
	}

	queryVals := authURL.Query()
	port, err := parseAndValidateRedirectPort(queryVals.Get("redirect_uri"))
	if err != nil {
		return "", err
	}

	// No need to use HTTPS as per https://www.rfc-editor.org/rfc/rfc8252#section-8.3
	server := &http.Server{
		Handler:      nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	callbackEndpoint.server = server
	defer server.Shutdown(context.Background())
	mux := http.NewServeMux()
	mux.Handle("/callback", callbackEndpoint)
	server.Handler = mux

	server.Addr = "127.0.0.1:" + strconv.Itoa(port)

	queryVals.Set("state", callbackEndpoint.state)
	authURL.RawQuery = queryVals.Encode()

	// Start the local server. This will fail if the server fails to start.
	errorSignal := make(chan error)
	listener, err := net.Listen("tcp", server.Addr)

	// This will stop execution if the port is already in use
	if err != nil {
		return "", err
	}

	go func() {
		err = server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			errorSignal <- err
		}
	}()

	// Open the browser to the OIDC provider
	err = openBrowserFn(authURL.String())
	if err != nil {
		return "", err
	}

	// Set a timeout and shut down the server if we don't get a response in time
	timeout := time.NewTimer(callbackServerTimeout)
	defer timeout.Stop()

	select {
	case <-timeout.C:
		return "", fmt.Errorf("timeout waiting for OIDC callback")
	case <-callbackEndpoint.shutdownSignal:
		return callbackEndpoint.code, nil
	case err := <-errorSignal:
		return "", err
	}
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func parseAndValidateRedirectPort(redirectURI string) (int, error) {
	// Use regex to validate redirect_uri is in the format http://127.0.0.1[:port]/callback
	if !regexp.MustCompile(`^http://127\.0\.0\.1(:[0-9]+)?/callback$`).MatchString(redirectURI) {
		return 0, fmt.Errorf("redirect_uri must be http://127.0.0.1[:port]/callback")
	}

	// Get port from uri
	uri, err := url.Parse(redirectURI)
	if err != nil {
		return 0, fmt.Errorf("Unable to parse redirect_uri: %s", err)
	}
	if uri.Port() != "" {
		port, _ := strconv.Atoi(uri.Port()) // This is certain to be an integer since it passed the regex check
		if port < 1 || port > 65535 {
			return 0, fmt.Errorf("Port in redirect_uri must be between 1 and 65535")
		}
		return port, nil
	}
	// Default to port 80
	return 80, nil
}
