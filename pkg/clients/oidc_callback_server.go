package clients

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"time"
)

type callbackEndpoint struct {
	server         *http.Server
	shutdownSignal chan string
	code           string
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

func handleOpenIDFlow(authEndpointURL string) (string, error) {

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
	callbackURL := "http://localhost:8888/callback"
	authURL, authURLParseError := url.Parse(authEndpointURL)
	if authURLParseError != nil {
		return "", authURLParseError
	}
	query := authURL.Query()
	query.Set("redirect_uri", callbackURL)
	authURL.RawQuery = query.Encode()

	// BEGIN TEST CODE
	// FIXME
	// Point to locally running keycloak instance for testing
	authURL.Scheme = "http"
	authURL.Host = "localhost:7777"
	// END TEST CODE

	// Start the local server
	go func() {
		server.ListenAndServe()
	}()

	// Open the browser to the OIDC provider
	cmd := exec.Command("open", authURL.String())
	cmdErorr := cmd.Start()
	if cmdErorr != nil {
		return "", cmdErorr
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
