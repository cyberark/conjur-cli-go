package clients

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wiremock/go-wiremock"
)

const (
	wiremockURL = "http://wiremock:8080"
	username    = "user@example.com"
	password    = "password"
)

func mockResponse(wiremockClient *wiremock.Client, path, response string, status int64) {
	_ = wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo(path)).
		WillReturnResponse(
			wiremock.NewResponse().WithBody(response).
				WithHeader("Content-Type", "application/json").
				WithStatus(status),
		).
		AtPriority(1))
}

func mockResponseWithRequestBody(wiremockClient *wiremock.Client, path, requestBody, response string, status int64) {
	_ = wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo(path)).
		WithBodyPattern(wiremock.EqualToJson(requestBody)).
		WillReturnResponse(
			wiremock.NewResponse().WithBody(response).
				WithHeader("Content-Type", "application/json").
				WithStatus(status),
		).
		AtPriority(1))
}

func TestIdentityAuthenticator_GetToken(t *testing.T) {
	wiremockClient := wiremock.NewClient(wiremockURL)

	testCases := []struct {
		name          string
		expectedToken string
		expectedError error
		timeout       time.Duration
		beforeTest    func()
	}{
		{
			name:          "Successful authentication (UP challenge)",
			expectedToken: "valid-token",
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_pass_only.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				advanceAuthResponse, _ := os.ReadFile("test/identity_mock/advance_auth_pass_only.json")
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", string(advanceAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "Start authentication request error",
			expectedError: errors.New("received non-200 response: 500"),
			beforeTest: func() {
				mockResponse(wiremockClient, "/Security/StartAuthentication", "500 Internal Server Error", http.StatusInternalServerError)
			},
		},
		{
			name:          "Start authentication error - failed authentication",
			expectedError: errors.New("authentication failed: some error"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_failure.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "Start authentication error - JSON parsing failed",
			expectedError: errors.New("failed to parse response: invalid character 'o' in literal null (expecting 'u')"),
			beforeTest: func() {
				mockResponse(wiremockClient, "/Security/StartAuthentication", "not a JSON response", http.StatusOK)
			},
		},
		{
			name:          "Start authentication error - no challenges available",
			expectedError: errors.New("no challenges available for authentication"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_no_challenges.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "Start authentication error - no mechanism available",
			expectedError: errors.New("no mechanisms available for authentication"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_no_mechanism.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "Advanced authentication request error (UP challenge)",
			expectedError: errors.New("received non-200 response: 500"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_pass_only.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", "500 Internal Server Error", http.StatusInternalServerError)
			},
		},
		{
			name:          "Advanced authentication error (UP challenge) - JSON parsing failed",
			expectedError: errors.New("failed to parse response: invalid character 'o' in literal null (expecting 'u')"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_pass_only.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", "not a JSON response", http.StatusOK)
			},
		},
		{
			name:          "Successful authentication (SMS/Email MFA)",
			expectedToken: "valid-token",
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_mfa.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id"}`, password)
				advanceAuthResponse, _ := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				startOOBRequest := `{"Action":"StartOOB","MechanismId":"email_mechanism_id","SessionId":"session_id"}`
				startOOBResponse, _ := os.ReadFile("test/identity_mock/advance_start_oob.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", startOOBRequest, string(startOOBResponse), http.StatusOK)
				OOBSuccessRequest := `{"Action":"Answer","MechanismId":"email_mechanism_id","SessionId":"session_id"}`
				OOBSuccessResponse, _ := os.ReadFile("test/identity_mock/advance_oob_success.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", OOBSuccessRequest, string(OOBSuccessResponse), http.StatusOK)
				pollRequest := `{"Action":"Poll","MechanismId":"email_mechanism_id","SessionId":"session_id"}`
				pollResponse, _ := os.ReadFile("test/identity_mock/advance_oob_success.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", pollRequest, string(pollResponse), http.StatusOK)
			},
		},
		{
			name:          "Successful authentication (QR MFA)",
			expectedToken: "valid-token",
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_qr.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id"}`, password)
				advanceAuthResponse, _ := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				startPollResponse, _ := os.ReadFile("test/identity_mock/advance_oob_success.json")
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", string(startPollResponse), http.StatusOK)
			},
		},
		{
			name:          "Successful authentication (External Action MFA with PIN)",
			expectedToken: "valid-token",
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_external_pin.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				OOBSuccessRequest := `{"Action":"Answer","MechanismId":"OOBAUTHPIN","SessionId":"lileTEcF0UOBN0viru2gfGqxkdAILx2xTg2IS4suWM41"}`
				OOBSuccessResponse, _ := os.ReadFile("test/identity_mock/advance_oob_success.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", OOBSuccessRequest, string(OOBSuccessResponse), http.StatusOK)
			},
		},
		{
			name:          "MFA Authentication - Unsupported mechanism",
			expectedError: errors.New("unsupported authentication mechanism: UNSUPPORTED"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_unsupported_mfa.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id"}`, password)
				advanceAuthResponse, _ := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "MFA Authentication - Start OOB request error",
			expectedError: errors.New("received non-200 response: 500"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_mfa.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id"}`, password)
				advanceAuthResponse, _ := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				advanceRequest := `{"Action":"Answer","Answer":"password","MechanismId":"password_mechanism_id","SessionId":"session_id"}`
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", advanceRequest, "500 Internal Server Error", http.StatusInternalServerError)
			},
		},
		{
			name:          "MFA Authentication - advance authentication request error",
			expectedError: errors.New("failed to poll authentication status: received non-200 response: 500"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_mfa.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id"}`, password)
				advanceAuthResponse, _ := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				startOOBRequest := `{"Action":"StartOOB","MechanismId":"email_mechanism_id","SessionId":"session_id"}`
				startOOBResponse, _ := os.ReadFile("test/identity_mock/advance_start_oob.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", startOOBRequest, string(startOOBResponse), http.StatusOK)
				OOBFailureRequest := `{"Action":"Poll","MechanismId":"email_mechanism_id","SessionId":"session_id"}`
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", OOBFailureRequest, "500 Internal Server Error", http.StatusInternalServerError)
			},
		},
		{
			name:          "MFA Authentication - advance authentication error",
			timeout:       1 * time.Second,
			expectedError: errors.New("Timed out waiting for out-of-band authentication"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_mfa.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id"}`, password)
				advanceAuthResponse, _ := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				startOOBRequest := `{"Action":"StartOOB","MechanismId":"email_mechanism_id","SessionId":"session_id"}`
				startOOBResponse, _ := os.ReadFile("test/identity_mock/advance_start_oob.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", startOOBRequest, string(startOOBResponse), http.StatusOK)
				OOBFailureRequest := `{"Action":"Poll","MechanismId":"email_mechanism_id","SessionId":"session_id"}`
				OOBFailureResponse, _ := os.ReadFile("test/identity_mock/advance_failure.json")
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", OOBFailureRequest, string(OOBFailureResponse), http.StatusOK)
			},
		},
		{
			name:          "External action MFA - request error",
			expectedError: errors.New("received non-200 response: 500"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_external_pin.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", "500 internal server error", http.StatusInternalServerError)
			},
		},
		{
			name:          "External action MFA - authentication failure",
			expectedError: errors.New("Authentication with federated identity provider failed: Authentication (login or challenge) has failed. Please try again or contact your system administrator."),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_external_pin.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				advanceFailureResponse, _ := os.ReadFile("test/identity_mock/advance_failure.json")
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", string(advanceFailureResponse), http.StatusOK)
			},
		},
		{
			name:          "External action MFA - JSON parsing failed",
			expectedError: errors.New("failed to parse response: invalid character 'o' in literal null (expecting 'u')"),
			beforeTest: func() {
				startAuthResponse, _ := os.ReadFile("test/identity_mock/start_auth_external_pin.json")
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", "not a JSON response", http.StatusOK)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			timeout := tc.timeout
			if timeout == 0 {
				timeout = defaultTimeout
			}

			authenticator := &IdentityAuthenticator{
				identityURL: wiremockURL,
				timeout:     timeout,
			}

			// Set up the test case
			if tc.beforeTest != nil {
				tc.beforeTest()
			}

			// Call the method under test
			token, err := authenticator.GetToken(username, password)

			// Assert the results
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedToken, token)

			// Clear stubs on wiremock client
			_ = wiremockClient.Clear()
		})
	}
}
