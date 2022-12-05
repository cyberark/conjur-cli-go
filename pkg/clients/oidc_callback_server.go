package clients

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/browser"
)

type callbackEndpoint struct {
	server         *http.Server
	shutdownSignal chan string
	code           string
}

func OpenBrowser(url string) error {
	return browser.OpenURL(url)
}

func (h *callbackEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code != "" {
		h.code = code
		fmt.Fprintln(w, "Logged in successfully. You may close the browser and return to the command line.")
	} else {
		fmt.Fprintln(w, "Failed to log in. You may close the browser and try again.")
	}
	h.shutdownSignal <- "shutdown"
}

func handleOpenIDFlow(authEndpointURL string, openBrowserFn func(string) error) (string, error) {

	callbackEndpoint := &callbackEndpoint{}
	callbackEndpoint.shutdownSignal = make(chan string)
	server := &http.Server{
		Addr:           ":8888",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	callbackEndpoint.server = server
	http.Handle("/callback", callbackEndpoint)

	// Replace the callback URL returned by Conjur with the URL to the local server we're starting
	//TODO: Instead, fail if the callback URL is not localhost:8888
	// callbackURL := "http://localhost:8888/callback"
	authURL, authURLParseError := url.Parse(authEndpointURL)
	if authURLParseError != nil {
		return "", authURLParseError
	}
	// query := authURL.Query()
	// query.Set("redirect_uri", callbackURL)
	// authURL.RawQuery = query.Encode()

	// BEGIN TEST CODE
	// FIXME
	// Point to locally running keycloak instance for testing
	// authURL.Scheme = "http"
	// authURL.Host = "localhost:7777"
	// END TEST CODE

	// Start the local server
	go func() {
		server.ListenAndServe()
	}()

	// Open the browser to the OIDC provider
	browserErr := openBrowserFn(authURL.String())
	if browserErr != nil {
		fmt.Printf("Error opening browser. Please open the following URL in your browser:\n%s\n", authURL.String())
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

	log.Println("Authorization code is ", callbackEndpoint.code)

	return callbackEndpoint.code, nil
}
