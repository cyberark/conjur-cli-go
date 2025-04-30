package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cyberark/conjur-cli-go/pkg/clients/qr"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
)

const (
	defaultTimeout = 5 * time.Minute

	// Define constants for authentication mechanism types
	mechanismSecurityQuestion    = "SQ"
	mechanismUserPassword        = "UP"
	mechanismSMS                 = "SMS"
	mechanismEMAIL               = "EMAIL"
	mechanismOATHOneTimePasscode = "OATH"
	mechanismIdentityMobileApp   = "OTP"
	mechanismFIDO2SecurityKey    = "U2F"
	mechanismQRCode              = "QR"
	mechanismPhoneCall           = "PF"
)

type identityResp[T any] struct {
	Success         bool   `json:"success"`
	Result          T      `json:"Result"`
	Message         string `json:"Message"`
	MessageId       string `json:"MessageID"`
	Exception       string `json:"Exception"`
	ErrorId         string `json:"ErrorID"`
	ErrorCode       string `json:"ErrorCode"`
	IsSoftError     bool   `json:"IsSoftError"`
	InnerExceptions string `json:"InnerExceptions"`
}

type authStartResp struct {
	ClientHints struct {
		PersistDefault                bool `json:"PersistDefault"`
		AllowPersist                  bool `json:"AllowPersist"`
		AllowForgotPassword           bool `json:"AllowForgotPassword"`
		EndpointAuthenticationEnabled bool `json:"EndpointAuthenticationEnabled"`
	} `json:"ClientHints"`
	Version            string `json:"Version"`
	SessionId          string `json:"SessionId"`
	EventDescription   string `json:"EventDescription"`
	RetryWaitingTime   int    `json:"RetryWaitingTime"`
	SecurityImageName  string `json:"SecurityImageName"`
	AllowLoginMfaCache bool   `json:"AllowLoginMfaCache"`
	Challenges         []struct {
		Mechanisms []mechanismResp `json:"Mechanisms"`
	} `json:"Challenges"`
	TenantId            string `json:"TenantId"`
	PodFQDN             string `json:"PodFQDN"`
	IdpRedirectShortUrl string `json:"IdpRedirectShortUrl"`
	IdpLoginSessionId   string `json:"IdpLoginSessionId"`
}

type mechanismResp struct {
	AnswerType       string `json:"AnswerType"`
	Name             string `json:"Name"`
	PromptMechChosen string `json:"PromptMechChosen"`
	PromptSelectMech string `json:"PromptSelectMech"`
	MechanismId      string `json:"MechanismId"`
	Enrolled         bool   `json:"Enrolled"`
	Image            string `json:"Image"`
}
type authAdvanceResp struct {
	Summary            string `json:"Summary"`
	GeneratedAuthValue string `json:"GeneratedAuthValue"`
	Token              string `json:"Token"`
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// IdentityAuthenticator struct
type IdentityAuthenticator struct {
	identityURL     string
	customTenantURL string
	client          ConjurClient
	sessionID       string
	httpClient      HTTPClient
}

// NewIdentityAuthenticator creates a new instance of IdentityAuthenticator
func NewIdentityAuthenticator(client ConjurClient, identityURL string, httpClient *http.Client) *IdentityAuthenticator {
	return &IdentityAuthenticator{
		identityURL: identityURL,
		client:      client,
		httpClient:  httpClient,
	}
}

func (ia *IdentityAuthenticator) GetToken(username, password string) (string, error) {
	startResp, err := ia.startAuthentication(username)
	if err != nil {
		return "", err
	}
	if !startResp.Success {
		return "", fmt.Errorf("authentication failed: %s", startResp.Message)
	}
	if len(startResp.Result.IdpRedirectShortUrl) > 0 && len(startResp.Result.IdpLoginSessionId) > 0 {
		return ia.waitForExternalAction(startResp.Result.IdpRedirectShortUrl, startResp.Result.IdpLoginSessionId)
	}
	if len(startResp.Result.Challenges) == 0 {
		return "", errors.New("no challenges available for authentication")
	}
	ia.sessionID = startResp.Result.SessionId
	for _, challenge := range startResp.Result.Challenges {
		if len(challenge.Mechanisms) == 0 {
			return "", errors.New("no mechanisms available for authentication")
		}
		mechanism, err := chooseMechanism(challenge.Mechanisms)
		if err != nil {
			return "", err
		}
		var advanceResp *identityResp[authAdvanceResp]
		switch strings.ToUpper(mechanism.Name) {
		case mechanismUserPassword:
			password, err = prompts.MaybeAskForPassword(password)
			if err != nil {
				return "", err
			}
			advanceResp, err = ia.advanceAuthentication(mechanism.MechanismId, "Answer", password)
		case mechanismSecurityQuestion, mechanismSMS, mechanismEMAIL,
			mechanismOATHOneTimePasscode, mechanismIdentityMobileApp,
			mechanismFIDO2SecurityKey, mechanismPhoneCall:
			advanceResp, err = ia.advanceAuthentication(mechanism.MechanismId, "StartOOB", "")
			if err != nil {
				return "", err
			}
			if answerable(mechanism.Name) {
				advanceResp, err = ia.handleAnswerableMechanism(mechanism)
			} else {
				advanceResp, err = ia.startPoll(mechanism.MechanismId)
			}
		case mechanismQRCode:
			err = qr.DisplayQRCode(mechanism.Image)
			if err != nil {
				return "", fmt.Errorf("failed to display QR code: %w", err)
			}
			advanceResp, err = ia.startPoll(mechanism.MechanismId)
		default:
			return "", fmt.Errorf("unsupported authentication mechanism: %s", mechanism.Name)
		}
		if err != nil {
			return "", err
		}
		if advanceResp != nil && !advanceResp.Success {
			return "", fmt.Errorf("authentication failed: %s", advanceResp.Message)
		}
		if advanceResp != nil && len(advanceResp.Result.Token) > 0 {
			return advanceResp.Result.Token, nil
		}
	}
	return "", nil
}

func chooseMechanism(mechanisms []mechanismResp) (*mechanismResp, error) {
	if len(mechanisms) == 1 {
		return &mechanisms[0], nil
	}
	options := make([]string, 0, len(mechanisms))
	for _, mech := range mechanisms {
		if mech.Enrolled {
			options = append(options, mech.PromptSelectMech)
		}
	}
	chosen, err := prompts.AskForMFAMechanism(options)
	if err != nil {
		return nil, err
	}
	for i, mech := range mechanisms {
		if mech.PromptSelectMech == chosen {
			return &mechanisms[i], nil
		}
	}
	return nil, fmt.Errorf("selected mechanism not found: %s", chosen)
}

func (ia *IdentityAuthenticator) answer(ctx context.Context, mechanism *mechanismResp) (*identityResp[authAdvanceResp], error) {
	answer, err := prompts.AskForPrompt(ctx, mechanism.PromptMechChosen, defaultTimeout)
	// if answer is empty, it means the input was canceled
	if err != nil || len(answer) == 0 {
		return nil, err
	}
	return ia.advanceAuthentication(mechanism.MechanismId, "Answer", answer)
}

func answerable(mechanismName string) bool {
	switch strings.ToUpper(mechanismName) {
	case mechanismSecurityQuestion, mechanismUserPassword, mechanismSMS,
		mechanismEMAIL, mechanismOATHOneTimePasscode, mechanismIdentityMobileApp,
		mechanismFIDO2SecurityKey:
		return true
	default:
		return false
	}
}

func (ia *IdentityAuthenticator) startPoll(mechanismID string) (advanceResp *identityResp[authAdvanceResp], err error) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(defaultTimeout)
	for {
		select {
		case <-timeout:
			return nil, errors.New("Timed out waiting for out-of-band authentication")
		case <-ticker.C:
			advanceResp, _ = ia.advanceAuthentication(mechanismID, "Poll", "")
			if advanceResp.Result.Summary == "LoginSuccess" || advanceResp.Result.Summary == "StartNextChallenge" {
				return
			}
		}
	}
}

func (ia *IdentityAuthenticator) startAuthentication(userName string) (*identityResp[authStartResp], error) {
	response, err := ia.callIdentityStartAuth(userName)
	if err != nil {
		return nil, err
	}

	var startAuthResponse identityResp[authStartResp]
	if err := json.Unmarshal(response, &startAuthResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if len(startAuthResponse.Result.PodFQDN) > 0 {
		ia.identityURL = startAuthResponse.Result.PodFQDN
		return ia.startAuthentication(userName)
	}
	return &startAuthResponse, nil
}

func (ia *IdentityAuthenticator) waitForExternalAction(idpURL, idpSessionID string) (string, error) {
	log.Printf("Opening browser for external authentication: %s", idpURL)
	_ = openBrowser(idpURL) // this may fail on headless systems, but we don't want to block the authentication process
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(defaultTimeout)
	for {
		select {
		case <-timeout:
			return "", errors.New("Timed out waiting for external authentication.")
		case <-ticker.C:
			token, _ := ia.getAuthStatusToken(idpSessionID)
			if len(token) > 0 {
				return token, nil
			}
		}
	}
}

func (ia *IdentityAuthenticator) advanceAuthentication(mechanismID, action, answer string) (*identityResp[authAdvanceResp], error) {
	payload := map[string]interface{}{
		"SessionId":   ia.sessionID,
		"MechanismId": mechanismID,
		"Action":      action,
	}
	if action == "Answer" && len(answer) > 0 {
		payload["Answer"] = answer
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	response, err := ia.invokeEndpoint(http.MethodPost, fmt.Sprintf("%s/Security/AdvanceAuthentication", ia.identityURL), payloadBytes)
	if err != nil {
		return nil, err
	}

	var advanceResponse identityResp[authAdvanceResp]
	if err := json.Unmarshal(response, &advanceResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &advanceResponse, nil
}

func (ia *IdentityAuthenticator) callIdentityStartAuth(userName string) ([]byte, error) {
	payload := map[string]string{
		"User":    userName,
		"Version": "1.0",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return ia.invokeEndpoint(http.MethodPost, fmt.Sprintf("%s/Security/StartAuthentication", ia.identityURL), payloadBytes)
}

func (ia *IdentityAuthenticator) getAuthStatusToken(sessionID string) (string, error) {
	payload := map[string]string{
		"SessionId": sessionID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	response, err := ia.invokeEndpoint(http.MethodPost, fmt.Sprintf("%s/Security/OobAuthStatus", ia.identityURL), payloadBytes)
	if err != nil {
		return "", err
	}

	var authStatus struct {
		Result struct {
			State string `json:"State"`
			Token string `json:"Token,omitempty"`
		} `json:"Result"`
	}
	if err := json.Unmarshal(response, &authStatus); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	switch authStatus.Result.State {
	case "pending":
		return "", nil
	case "success":
		return authStatus.Result.Token, nil
	default:
		return "", errors.New("authentication failed")
	}
}

func (ia *IdentityAuthenticator) invokeEndpoint(method, url string, payload []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-IDAP-NATIVE-CLIENT", "true")

	resp, err := ia.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (ia *IdentityAuthenticator) handleAnswerableMechanism(mechanism *mechanismResp) (*identityResp[authAdvanceResp], error) {
	respCh := make(chan *identityResp[authAdvanceResp])
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	// the done chanel is needed to wait until the user input is completed, either by answering the question or canceling the input
	// this is required to allow user input library to cleanup the console before we can move on to the following challenge
	done := make(chan bool)
	go func() {
		defer func() { done <- true }()
		resp, err := ia.answer(ctx, mechanism)
		if err != nil {
			errCh <- err
		}
		if resp != nil {
			respCh <- resp
		}
	}()
	go func() {
		resp, err := ia.startPoll(mechanism.MechanismId)
		if err != nil {
			errCh <- err
		}
		if resp != nil {
			respCh <- resp
		}
	}()
	select {
	case advanceResp := <-respCh:
		cancel()
		<-done
		return advanceResp, nil
	case err := <-errCh:
		cancel()
		<-done
		return nil, err
	}
}
