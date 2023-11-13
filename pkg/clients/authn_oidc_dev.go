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
	"time"

	"golang.org/x/exp/slices"

	"github.com/cyberark/conjur-cli-go/pkg/prompts"
)

// OidcLogin attempts to login to Conjur using the OIDC flow. Username and password can be provided to
// bypass the browser and use the username and password to fetch an OIDC code. This option is meant for testing
// purposes only and will print a warning.
func OidcLogin(conjurClient ConjurClient, username string, password string) (ConjurClient, error) {
	username, password, err := prompts.MaybeAskForCredentials(username, password)
	if err != nil {
		return nil, err
	}

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
		case strings.Contains(providerURL, "idaptive"):
			fmt.Println("Using Identity as OIDC provider")
			return fetchOidcCodeFromIdentity(httpClient, username, password, providerURL)
		default:
			fmt.Println("Using Keycloak as OIDC provider")
			return fetchOidcCodeFromKeycloak(httpClient, username, password, providerURL)
		}
	}
}

func fetchOidcCodeFromKeycloak(httpClient *http.Client, username, password, providerURL string) error {
	// Note: This shouldn't be required to make keycloak work, but it doesn't matter much since it's only used for tests
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

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
		callbackRedirect(resp.Header.Get("Location"))
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
		callbackRedirect(resp.Header.Get("Location"))
		return nil
	}
	return errors.New("unable to fetch authorization code from Okta")
}

func fetchSessionTokenFromOkta(httpClient *http.Client, providerURL string, username string, password string) (string, error) {
	fmt.Println("Attempting to get session token from Okta using the provided username/password")

	type OktaRequestData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	requestJSON := OktaRequestData{
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

	type OktaResponseData struct {
		ExpiresAt    string `json:"expiresAt"`
		Status       string `json:"status"`
		SessionToken string `json:"sessionToken"`
	}

	var respJSON OktaResponseData

	err = json.Unmarshal(respBody, &respJSON)
	if err != nil {
		return "", err
	}

	return respJSON.SessionToken, nil
}

func callbackRedirect(location string) error {
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

func fetchOidcCodeFromIdentity(httpClient *http.Client, username, password, providerURL string) error {
	authToken, err := fetchAuthTokenFromIdentity(httpClient, providerURL, username, password)
	if err != nil {
		return err
	}

	target := providerURL
	code := 0
	for !strings.Contains(target, "127.0.0.1:8888/callback") {
		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Accept", "*/*")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}

		target = resp.Header.Get("Location")
		code = resp.StatusCode
	}

	if code == 302 {
		callbackRedirect(target)
		return nil
	}
	return errors.New("unable to fetch authorization code from Identity")
}

type startAuthData struct {
	Version  string `json:"Version"`
	Username string `json:"User"`
}

type advanceAuthData struct {
	Action      string `json:"Action"`
	Answer      string `json:"Answer,omitempty"`
	MechanismId string `json:"MechanismId"`
	SessionId   string `json:"SessionId"`
}

type mechanism struct {
	PromptSelectMech string `json:"PromptSelectMech"`
	MechanismId      string `json:"MechanismId"`
}

type startAuthResponse struct {
	Success bool `json:"success"`
	Result  struct {
		SessionId  string `json:"SessionId,omitempty"`
		Challenges []struct {
			Mechanisms []mechanism `json:"Mechanisms"`
		} `json:"Challenges,omitempty"`
		Summary string `json:"Summary,omitempty"`
		PodFqdn string `json:"PodFqdn",omitempty`
	} `json:"Result"`
	Message         string `json:"Message"`
	MessageId       string `json:"MessageID"`
	Exception       string `json:"Exception"`
	ErrorId         string `json:"ErrorID"`
	ErrorCode       string `json:"ErrorCode"`
	IsSoftError     bool   `json:"IsSoftError"`
	InnerExceptions string `json:"InnerExceptions"`
}

type advanceAuthResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Summary            string `json:"Summary"`
		GeneratedAuthValue string `json:"GeneratedAuthValue"`
	} `json:"Result"`
	Message         string `json:"Message"`
	MessageId       string `json:"MessageID"`
	Exception       string `json:"Exception"`
	ErrorId         string `json:"ErrorID"`
	ErrorCode       string `json:"ErrorCode"`
	IsSoftError     bool   `json:"IsSoftError"`
	InnerExceptions string `json:"InnerExceptions"`
}

func authRequest[data startAuthData | advanceAuthData, responseContent startAuthResponse | advanceAuthResponse](
	endpoint string,
	payload data,
	httpClient *http.Client,
) (*http.Response, responseContent, error) {
	var content responseContent

	byteData, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(byteData))
	if err != nil {
		fmt.Println(err)
		return nil, content, err
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, content, err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, content, err
	}

	err = json.Unmarshal(responseBody, &content)
	if err != nil {
		return nil, content, err
	}

	return resp, content, nil
}

func startAuthRequest(hostname string, data startAuthData, httpClient *http.Client) (*http.Response, startAuthResponse, error) {
	endpoint := hostname + "/Security/StartAuthentication"
	return authRequest[startAuthData, startAuthResponse](endpoint, data, httpClient)
}

func advanceAuthRequest(hostname string, data advanceAuthData, httpClient *http.Client) (*http.Response, advanceAuthResponse, error) {
	endpoint := hostname + "/Security/AdvanceAuthentication"
	return authRequest[advanceAuthData, advanceAuthResponse](endpoint, data, httpClient)
}

func fetchAuthTokenFromIdentity(httpClient *http.Client, providerURL string, username string, password string) (string, error) {
	fmt.Println("Attempting to get auth token from Identity using the provided username/password")

	// A request to /Security/StartAuthentication begins the login process,
	// and returns a list of authentication mechanisms to engage with.

	host := extractHostname(providerURL)
	startData := startAuthData{
		Version:  "1.0",
		Username: username,
	}
	_, startResp, err := startAuthRequest(host, startData, httpClient)
	if err != nil {
		return "", err
	}
	if startResp.Result.PodFqdn != "" {
		host = fmt.Sprintf("https://%s", startResp.Result.PodFqdn)
		_, startResp, err = startAuthRequest(host, startData, httpClient)
		if err != nil {
			return "", err
		}
	}

	primaryMechanisms := startResp.Result.Challenges[0].Mechanisms

	// Usually, we would iterate through MFA challenges sequentially.
	// For our purposes, though, we want to make sure to use the Password
	// and Mobile Authenticator mechanisms.

	passwordMechanism := primaryMechanisms[slices.IndexFunc(
		primaryMechanisms,
		func(m mechanism) bool {
			return m.PromptSelectMech == "Password"
		},
	)]

	// Advance Password-based authentication handshake.

	resp, advanceResp, err := advanceAuthRequest(host, advanceAuthData{
		Action:      "Answer",
		Answer:      password,
		MechanismId: passwordMechanism.MechanismId,
		SessionId:   startResp.Result.SessionId,
	}, httpClient)
	if err != nil {
		return "", nil
	}
	if !(advanceResp.Success) {
		return "", errors.New(advanceResp.Message)
	}

	// If only one challenge exists (Password), return the token immediately
	if len(startResp.Result.Challenges) == 1 {
		return resp.Cookies()[slices.IndexFunc(
			resp.Cookies(), func(c *http.Cookie) bool {
				return c.Name == ".ASPXAUTH"
			},
		)].Value, nil
	} else {

		// Otherwise advance the Mobile Authenticator-based authentication handshake.
		secondaryMechanisms := startResp.Result.Challenges[1].Mechanisms

		mobileAuthMechanism := secondaryMechanisms[slices.IndexFunc(
			secondaryMechanisms,
			func(m mechanism) bool {
				return m.PromptSelectMech == "Mobile Authenticator"
			},
		)]

		resp, advanceResp, err = advanceAuthRequest(host, advanceAuthData{
			Action:      "StartOOB",
			MechanismId: mobileAuthMechanism.MechanismId,
			SessionId:   startResp.Result.SessionId,
		}, httpClient)
		if err != nil {
			return "", err
		}
		if !(advanceResp.Success) {
			return "", errors.New(advanceResp.Message)
		}
		fmt.Println(fmt.Sprintf("\nDev env users: select %s in your Identity notification", advanceResp.Result.GeneratedAuthValue))

		// For 30 seconds, Poll for out-of-band authentication success.

		poll := func() error {
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()
			timeout := time.After(30 * time.Second)
			for {
				select {
				case <-timeout:
					return errors.New("Timed out waiting for out-of-band authentication")
				case <-ticker.C:
					resp, advanceResp, err = advanceAuthRequest(host, advanceAuthData{
						Action:      "Poll",
						MechanismId: mobileAuthMechanism.MechanismId,
						SessionId:   startResp.Result.SessionId,
					}, httpClient)
					if err != nil {
						return err
					}

					if advanceResp.Result.Summary == "LoginSuccess" {
						return nil
					}
				}
			}
		}

		err = poll()
		if err != nil {
			return "", err
		}

		// When the OOB Poll response indicates LoginSuccess, the bearer token is
		// included as an .ASPXAUTH cookie.

		return resp.Cookies()[slices.IndexFunc(
			resp.Cookies(), func(c *http.Cookie) bool {
				return c.Name == ".ASPXAUTH"
			},
		)].Value, nil
	}
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
