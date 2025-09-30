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
	"runtime"
	"strings"
	"time"

	"github.com/cyberark/conjur-cli-go/pkg/clients/qr"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/cyberark/conjur-cli-go/pkg/version"
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

	pinChallengeMechanism = "OOBAUTHPIN"
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

type authStartReq struct {
	User     string `json:"User"`
	Version  string `json:"Version"`
	TenantID string `json:"TenantId,omitempty"`
}

type authStartResp struct {
	ClientHints struct {
		PersistDefault                bool `json:"PersistDefault"`
		AllowPersist                  bool `json:"AllowPersist"`
		AllowForgotPassword           bool `json:"AllowForgotPassword"`
		EndpointAuthenticationEnabled bool `json:"EndpointAuthenticationEnabled"`
	} `json:"ClientHints"`
	Version            string `json:"Version"`
	SessionID          string `json:"SessionId"`
	EventDescription   string `json:"EventDescription"`
	RetryWaitingTime   int    `json:"RetryWaitingTime"`
	SecurityImageName  string `json:"SecurityImageName"`
	AllowLoginMfaCache bool   `json:"AllowLoginMfaCache"`
	Challenges         []struct {
		Mechanisms []mechanismResp `json:"Mechanisms"`
	} `json:"Challenges"`
	TenantID              string `json:"TenantId"`
	PodFQDN               string `json:"PodFQDN"`
	IdpRedirectShortURL   string `json:"IdpRedirectShortUrl"`
	IdpRedirectURL        string `json:"IdpRedirectUrl"`
	IdpLoginSessionID     string `json:"IdpLoginSessionId"`
	IdpOobAuthPinRequired bool   `json:"IdpOobAuthPinRequired"`
}

type mechanismResp struct {
	AnswerType         string              `json:"AnswerType"`
	Name               string              `json:"Name"`
	PromptMechChosen   string              `json:"PromptMechChosen"`
	PromptSelectMech   string              `json:"PromptSelectMech"`
	MechanismId        string              `json:"MechanismId"`
	Enrolled           bool                `json:"Enrolled"`
	MultipartMechanism *multiPartMechanism `json:"MultipartMechanism,omitempty"`
	Image              string              `json:"Image"`
}

type multiPartMechanism struct {
	PromptSelectMech string `json:"PromptSelectMech"`
	MechanismParts   []struct {
		Uuid             string `json:"Uuid"`
		QuestionText     string `json:"QuestionText"`
		PromptMechChosen string `json:"PromptMechChosen"`
	} `json:"MechanismParts"`
}

type authAdvanceReq struct {
	SessionID   string `json:"SessionId"`
	MechanismId string `json:"MechanismId"`
	Action      string `json:"Action"`
	Answer      any    `json:"Answer,omitempty"`
	TenantID    string `json:"TenantId,omitempty"`
}

type authAdvanceResp struct {
	Summary            string `json:"Summary"`
	GeneratedAuthValue string `json:"GeneratedAuthValue"`
	Token              string `json:"Token"`
}

type authStatusOOBReq struct {
	SessionID string `json:"SessionId"`
	TenantID  string `json:"TenantId,omitempty"`
}

type authStatusOOBResp struct {
	Result struct {
		State string `json:"State"`
		Token string `json:"Token,omitempty"`
	} `json:"Result"`
}

// IdentityAuthenticator struct
type IdentityAuthenticator struct {
	identityURL string
	tenantID    string
	client      ConjurClient
	sessionID   string
	timeout     time.Duration
}

// NewIdentityAuthenticator creates a new instance of IdentityAuthenticator
func NewIdentityAuthenticator(client ConjurClient, identityURL, tenantID string) *IdentityAuthenticator {
	timeout := time.Duration(client.GetConfig().ConjurCloudTimeout)
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &IdentityAuthenticator{
		identityURL: identityURL,
		tenantID:    tenantID,
		client:      client,
		timeout:     timeout,
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
	if len(startResp.Result.IdpRedirectShortURL) > 0 && len(startResp.Result.IdpLoginSessionID) > 0 {
		if startResp.Result.IdpOobAuthPinRequired == false {
			return "", fmt.Errorf("oob auth pin required for login with external identity provider")
		}
		ia.sessionID = startResp.Result.IdpLoginSessionID
		return ia.loginWithPIN(startResp.Result.IdpRedirectShortURL)
	}
	if len(startResp.Result.Challenges) == 0 {
		return "", errors.New("no challenges available for authentication")
	}
	ia.sessionID = startResp.Result.SessionID
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
		case mechanismSMS, mechanismEMAIL,
			mechanismOATHOneTimePasscode, mechanismIdentityMobileApp,
			mechanismFIDO2SecurityKey, mechanismPhoneCall:
			advanceResp, err = ia.advanceAuthentication(mechanism.MechanismId, "StartOOB", nil)
			if err != nil {
				return "", err
			}
			if answerable(mechanism.Name) {
				advanceResp, err = ia.handleAnswerableMechanism(mechanism)
			} else {
				advanceResp, err = ia.startPoll(mechanism.MechanismId)
			}
		case mechanismSecurityQuestion:
			advanceResp, err = ia.handleSecurityQuestions(mechanism)
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
	options := make([]prompts.Option, 0, len(mechanisms))
	for _, mech := range mechanisms {
		if mech.Enrolled && len(mech.Name) > 0 {
			options = append(options, prompts.Option{promptSelectMech(mech), mech.Name})
		}
	}
	chosen, err := prompts.AskForMFAMechanism(options)
	if err != nil {
		return nil, err
	}
	for i, mech := range mechanisms {
		if mech.Name == chosen {
			return &mechanisms[i], nil
		}
	}
	return nil, fmt.Errorf("selected mechanism not found: %s", chosen)
}

func (ia *IdentityAuthenticator) answer(ctx context.Context, mechanism *mechanismResp) (*identityResp[authAdvanceResp], error) {
	answer, err := prompts.AskForPrompt(ctx, promptMechChosen(mechanism), ia.timeout)
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
	timeout := time.After(ia.timeout)
	for {
		select {
		case <-timeout:
			return nil, errors.New("Timed out waiting for out-of-band authentication")
		case <-ticker.C:
			advanceResp, err = ia.advanceAuthentication(mechanismID, "Poll", nil)
			if err != nil {
				return nil, fmt.Errorf("failed to poll authentication status: %w", err)
			}
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
	if len(ia.tenantID) == 0 {
		ia.tenantID = startAuthResponse.Result.TenantID
	}
	return &startAuthResponse, nil
}

func (ia *IdentityAuthenticator) loginWithPIN(idpURL string) (string, error) {
	log.Printf("Opening browser for external authentication: %s", idpURL)
	_ = openBrowser(idpURL) // this may fail on headless systems, but we don't want to block the authentication process
	ctx := context.Background()
	answer, err := prompts.AskForPrompt(ctx, "Enter the pin code that you received in the browser", ia.timeout)
	if err != nil {
		return "", err
	}
	resp, err := ia.advanceAuthentication(pinChallengeMechanism, "Answer", answer)
	if err != nil {
		return "", err
	}
	if resp.Success != true {
		return "", fmt.Errorf("Authentication with federated identity provider failed: %s", resp.Message)
	}
	if len(resp.Result.Token) == 0 {
		return "", errors.New("Authentication with federated identity provider failed: no token received")
	}
	return resp.Result.Token, nil
}

func (ia *IdentityAuthenticator) waitForExternalAction(idpURL, idpSessionID string) (string, error) {
	log.Printf("Opening browser for external authentication: %s", idpURL)
	_ = openBrowser(idpURL) // this may fail on headless systems, but we don't want to block the authentication process
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(ia.timeout)
	for {
		select {
		case <-timeout:
			return "", errors.New("Timed out waiting for external authentication.")
		case <-ticker.C:
			token, err := ia.getAuthStatusToken(idpSessionID)
			if err != nil {
				return "", err
			}
			if len(token) > 0 {
				return token, nil
			}
		}
	}
}

func (ia *IdentityAuthenticator) advanceAuthentication(mechanismID, action string, answer any) (*identityResp[authAdvanceResp], error) {
	payload := authAdvanceReq{
		SessionID:   ia.sessionID,
		MechanismId: mechanismID,
		Action:      action,
		TenantID:    ia.tenantID,
		Answer:      answer,
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
	payload := authStartReq{
		User:     userName,
		Version:  "1.0",
		TenantID: ia.tenantID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return ia.invokeEndpoint(http.MethodPost, fmt.Sprintf("%s/Security/StartAuthentication", ia.identityURL), payloadBytes)
}

func (ia *IdentityAuthenticator) getAuthStatusToken(sessionID string) (string, error) {
	payload := authStatusOOBReq{
		SessionID: sessionID,
		TenantID:  ia.tenantID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	response, err := ia.invokeEndpoint(http.MethodPost, fmt.Sprintf("%s/Security/OobAuthStatus", ia.identityURL), payloadBytes)
	if err != nil {
		return "", err
	}

	var authStatus authStatusOOBResp
	if err := json.Unmarshal(response, &authStatus); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	switch strings.ToLower(authStatus.Result.State) {
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
	req.Header.Set("OobIdPAuth", "true")
	req.Header.Set("User-Agent", userAgent())

	client := http.DefaultClient
	if ia.client != nil && ia.client.GetHttpClient() != nil {
		client = ia.client.GetHttpClient()
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func userAgent() string {
	var platform string
	switch runtime.GOOS {
	case "darwin":
		platform = "Macintosh"
	case "linux":
		platform = "Linux"
	case "windows":
		platform = "Windows"
	default:
		platform = runtime.GOOS
	}
	return fmt.Sprintf("conjur-cli-go/%s (%s; %s %s)", version.Version, platform, runtime.GOOS, runtime.GOARCH)
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

func (ia *IdentityAuthenticator) handleSecurityQuestions(mechanism *mechanismResp) (*identityResp[authAdvanceResp], error) {
	if mechanism.MultipartMechanism == nil || len(mechanism.MultipartMechanism.MechanismParts) < 2 {
		return ia.answer(context.Background(), mechanism)
	}
	questions := make([]string, len(mechanism.MultipartMechanism.MechanismParts))
	for i, part := range mechanism.MultipartMechanism.MechanismParts {
		questions[i] = part.QuestionText
	}
	answers, err := prompts.AskForPrompts(
		context.Background(),
		mechanism.MultipartMechanism.PromptSelectMech,
		questions,
	)
	if err != nil {
		return nil, err
	}
	resp := make(map[string]string)
	for i, part := range mechanism.MultipartMechanism.MechanismParts {
		resp[part.Uuid] = answers[i]
	}
	return ia.advanceAuthentication(mechanism.MechanismId, "Answer", resp)
}

func promptMechChosen(mech *mechanismResp) string {
	if mech == nil {
		return ""
	}
	if len(mech.PromptMechChosen) > 0 {
		return mech.PromptMechChosen
	}
	switch strings.ToUpper(mech.Name) {
	case mechanismSecurityQuestion:
		return "Please answer your security question"
	case mechanismUserPassword:
		return "Please enter your password"
	case mechanismSMS:
		return "Please enter the code sent to your phone via SMS"
	case mechanismEMAIL:
		return "Please enter the code sent to your email"
	case mechanismOATHOneTimePasscode:
		return "Please enter your OATH one-time passcode"
	case mechanismIdentityMobileApp:
		return "Please enter the code from your identity mobile app"
	case mechanismFIDO2SecurityKey:
		return "Please complete the FIDO2 security key challenge"
	default:
		return "Please provide the required input for the selected authentication mechanism"
	}
}

func promptSelectMech(mech mechanismResp) string {
	if len(mech.PromptSelectMech) > 0 {
		return mech.PromptSelectMech
	}
	switch strings.ToUpper(mech.Name) {
	case mechanismSecurityQuestion:
		return "Security Question"
	case mechanismUserPassword:
		return "Password"
	case mechanismSMS:
		return "SMS"
	case mechanismEMAIL:
		return "Email"
	case mechanismOATHOneTimePasscode:
		return "OATH One-Time Passcode"
	case mechanismIdentityMobileApp:
		return "Identity Mobile App"
	case mechanismFIDO2SecurityKey:
		return "FIDO2 Security Key"
	case mechanismQRCode:
		return "QR Code"
	case mechanismPhoneCall:
		return "Phone Call"
	default:
		return mech.Name
	}
}
