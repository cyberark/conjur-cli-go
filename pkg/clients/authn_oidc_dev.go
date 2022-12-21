//go:build dev
// +build dev

package clients

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/cyberark/conjur-cli-go/pkg/utils"
)

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

// fetchOidcCodeFromProvider attempts to bypass the browser by using the username and password to fetch
// an OIDC code. This works for Keycloak and possibly other OIDC providers but will not work if MFA is needed.
func fetchOidcCodeFromProvider(username, password string) func(providerURL string) error {
	return func(providerURL string) error {
		// Note: This shouldn't be required to make keycloak work, but it doesn't matter much since it's only used for tests
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		// Create a client with redirect following disabled so we can handle the redirect ourselves for tests
		httpClient := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		switch {
		case strings.Contains(providerURL, "okta"):
			fmt.Println("Using Okta as OIDC provider")
			return fetchOidcCodeFromOkta(httpClient, username, password, providerURL)
		default:
			fmt.Println("Using Keycloak as OIDC provider")
			return fetchOidcCodeFromKeycloak(httpClient, username, password, providerURL)
		}
	}
}

func fetchOidcCodeFromKeycloak(httpClient *http.Client, username, password, providerURL string) error {
	// Get the provider's login page
	resp, err := httpClient.Get(providerURL)

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

	resp, err = httpClient.Do(req)
	if err != nil {
		return err
	}

	// If the login was successful, the provider will redirect to the callback URL with a code
	if resp.StatusCode == 302 {
		callbackRedirect(resp)
		return nil
	}

	return errors.New("unable to fetch code from Keycloak")
}

func fetchOidcCodeFromOkta(httpClient *http.Client, username, password, providerURL string) error {
	sessionToken, err := fetchSessionTokenFromOkta(httpClient, providerURL, username, password)
	if err != nil {
		return err
	}

	providerURL += "&sessionToken=" + sessionToken

	resp, err := httpClient.Get(providerURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If the login was successful, redirect to the callback URL with a code
	if resp.StatusCode == 302 {
		callbackRedirect(resp)
		return nil
	}
	return errors.New("unable to fetch authorization code from Okta")
}

func fetchSessionTokenFromOkta(httpClient *http.Client, providerURL string, username string, password string) (string, error) {
	fmt.Println("Attempting to get session token from Okta using the provided username/password")

	type OktaRequestBody struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	requestJSON := OktaRequestBody{
		Username: username,
		Password: password,
	}

	payload, _ := json.Marshal(requestJSON)
	authEndpoint := extractHostname(providerURL) + "/api/v1/authn"
	req, err := http.NewRequest("POST", authEndpoint, bytes.NewBuffer(payload))

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	sessionToken, err := utils.ExtractValueFromJSON(string(respBody), "sessionToken")
	if err != nil {
		return "", err
	}
	return sessionToken, nil
}

func callbackRedirect(resp *http.Response) error {
	location := string(resp.Header.Get("Location"))
	go func() {
		// Send a request to the callback URL to pass the code back to the CLI
		resp, err := http.Get(location)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()
	}()
	return nil
}

func extractHostname(providerURL string) string {
	url, err := url.Parse(providerURL)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	hostname := url.Scheme + "://" + url.Host
	return hostname
}
