package clients

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
)

type callbackEndpoint struct {
	server         *http.Server
	shutdownSignal chan string
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
	h.shutdownSignal <- "shutdown"
}

func handleOpenIDFlow(authEndpointURL string, generateStateFn func() string, openBrowserFn func(string) error) (string, error) {

	callbackEndpoint := &callbackEndpoint{}
	callbackEndpoint.state = generateStateFn()
	callbackEndpoint.shutdownSignal = make(chan string)
	server := &http.Server{
		Addr:         "127.0.0.1:8888",
		Handler:      nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	// No need to use HTTPS as per https://www.rfc-editor.org/rfc/rfc8252#section-8.3
	callbackEndpoint.server = server
	http.Handle("/callback", callbackEndpoint)

	authURL, authURLParseError := url.Parse(authEndpointURL)
	if authURLParseError != nil {
		return "", authURLParseError
	}

	queryVals := authURL.Query()

	// TODO: What if this port isn't available? Should this be configurable?
	if queryVals.Get("redirect_uri") != "http://127.0.0.1:8888/callback" {
		return "", fmt.Errorf("redirect_uri must be http://127.0.0.1:8888/callback")
	}

	queryVals.Set("state", callbackEndpoint.state)
	authURL.RawQuery = queryVals.Encode()

	// Start the local server
	go func() {
		server.ListenAndServe()
	}()

	// Wait for the server to start before opening the browser. This ensures that the browser won't
	// be opened if the server fails to start.
	time.Sleep(100 * time.Millisecond)

	// Open the browser to the OIDC provider
	err := openBrowserFn(authURL.String())
	if err != nil {
		return "", err
	}

	// Set a timeout and shut down the server if we don't get a response in time
	timeout := time.NewTimer(300 * time.Second)
	defer timeout.Stop()

	select {
	case <-timeout.C:
		callbackEndpoint.server.Shutdown(context.Background())
		return "", fmt.Errorf("timeout waiting for OIDC callback")
	case <-callbackEndpoint.shutdownSignal:
		callbackEndpoint.server.Shutdown(context.Background())
	}

	return callbackEndpoint.code, nil
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
