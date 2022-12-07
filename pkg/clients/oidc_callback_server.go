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

// Attempts to bypass the browser by using the username and password to fetch an OIDC code.
// This works for Keycloak and possibly other OIDC providers but will not work if MFA is needed.
func FetchOidcCodeFromProvider(username, password string) func(providerUrl string) error {
	return func(providerUrl string) error {
		//FIXME: This shouldn't be required to make keycloak work
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		// Get the provider's login page
		resp, err := http.Get(providerUrl)

		if err != nil {
			return err
		}
		defer resp.Body.Close()
		content, _ := io.ReadAll(resp.Body)

		cookies := resp.Cookies()

		// Assuming there's a form on the page, get the action URL
		actionRegex := regexp.MustCompile(`<form[\s\S]+action="([^"]+)"`)
		matches := actionRegex.FindStringSubmatch(string(content))
		if len(matches) < 2 {
			return errors.New("Unable to find login form")
		}
		action := html.UnescapeString(matches[1])

		// Pass the username and password to the login form
		credentials := fmt.Sprintf("username=%s&password=%s", username, password)
		req, err := http.NewRequest("POST", action, strings.NewReader(credentials))
		if err != nil {
			return err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Include the cookies from the login page. These are used to identify the user session.
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}

		// Disable redirect following so we can handle the redirect ourselves
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err = client.Do(req)
		if err != nil {
			return err
		}

		// If the login was successful, the provider will redirect to the callback URL with a code
		if resp.StatusCode == 302 || resp.StatusCode == 301 {
			location := string(resp.Header.Get("Location"))
			go func() {
				// Send a request to the callback URL to pass the code back to the CLI
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

		return errors.New("Unable to fetch code from OIDC provider")
	}
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

	authURL, authURLParseError := url.Parse(authEndpointURL)
	if authURLParseError != nil {
		return "", authURLParseError
	}

	//TODO: Fail if the callback URL is not localhost:8888/callback
	// callbackURL := "http://localhost:8888/callback"
	if authURL.Query().Get("redirect_uri") != "http://localhost:8888/callback" {
		return "", fmt.Errorf("redirect_uri must be http://localhost:8888/callback")
	}

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
