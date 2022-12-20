//go:build dev
// +build dev

package clients

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// fetchOidcCodeFromProvider attempts to bypass the browser by using the username and password to fetch
// an OIDC code. This works for Keycloak and possibly other OIDC providers but will not work if MFA is needed.
func fetchOidcCodeFromProvider(username, password string) func(providerUrl string) error {
	return func(providerUrl string) error {
		// Note: This shouldn't be required to make keycloak work, but it doesn't matter much since it's only used for tests
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		// Get the provider's login page
		resp, err := httpClient.Get(providerUrl)

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
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		resp, err = httpClient.Do(req)
		if err != nil {
			return err
		}

		// If the login was successful, the provider will redirect to the callback URL with a code
		if resp.StatusCode == 302 {
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

// OidcLogin attempts to login to Conjur using the OIDC flow. Username and password can be provided to
// bypass the browser and use the username and password to fetch an OIDC code. This option is meant for testing
// purposes only and will print a warning.
func OidcLogin(conjurClient ConjurClient, username string, password string) (ConjurClient, error) {
	// If a username and password is provided, attempt to use them to fetch an OIDC code instead of opening a browser
	oidcPromptHandler := openBrowser
	if username != "" && password != "" {
		fmt.Println("Attempting to use username and password to fetch OIDC code. " +
			"This is meant for testing purposes only and should not be used in production.")
		oidcPromptHandler = fetchOidcCodeFromProvider(username, password)
	}

	return oidcLogin(conjurClient, oidcPromptHandler)
}
