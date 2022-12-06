package clients

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/browser"
)

type callbackEndpoint struct {
	server         *http.Server
	shutdownSignal chan string
	code           string
}

func OpenBrowser(url string) error {
	err := browser.OpenURL(url)
	if err != nil {
		fmt.Printf("Error opening browser %s.\nPlease open the following URL in your browser:\n%s\n", err, url)
	}
	// Don't return an error - we want to continue even if the browser didn't open because the user
	// can open it manually and continue the authn flow.
	return nil
}

func FetchOidcCodeFromKeycloak(providerUrl string) error {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(providerUrl)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)

	cookies := resp.Cookies()

	actionRegex := regexp.MustCompile(`<form[\s\S]+action="([^"]+)"`)
	action := html.UnescapeString(actionRegex.FindStringSubmatch(string(content))[1])

	// TODO: Pull username and password from environment variables?
	req, err := http.NewRequest("POST", action, strings.NewReader("username=alice&password=alice"))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// Disable redirect following
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 302 || resp.StatusCode == 301 {
		location := string(resp.Header.Get("Location"))
		go func() {
			resp, err = http.Get(location)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			content, _ = io.ReadAll(resp.Body)
		}()
		return nil
	}

	return errors.New("Something went wrong")
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
