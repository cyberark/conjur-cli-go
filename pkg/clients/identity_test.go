package clients

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
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
		insecureLogin bool
		beforeTest    func(t *testing.T)
		responses     map[string]string
	}{
		{
			name:          "Successful authentication (UP challenge)",
			expectedToken: "valid-token",
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_pass_only.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				advanceAuthResponse, err := os.ReadFile("test/identity_mock/advance_auth_pass_only.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", string(advanceAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "Start authentication request error",
			expectedError: errors.New("received non-200 response: 500"),
			beforeTest: func(t *testing.T) {
				mockResponse(wiremockClient, "/Security/StartAuthentication", "500 Internal Server Error", http.StatusInternalServerError)
			},
		},
		{
			name:          "Start authentication error - failed authentication",
			expectedError: errors.New("authentication failed: some error"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_failure.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "Start authentication error - JSON parsing failed",
			expectedError: errors.New("failed to parse response: invalid character 'o' in literal null (expecting 'u')"),
			beforeTest: func(t *testing.T) {
				mockResponse(wiremockClient, "/Security/StartAuthentication", "not a JSON response", http.StatusOK)
			},
		},
		{
			name:          "Start authentication error - no challenges available",
			expectedError: errors.New("no challenges available for authentication"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_no_challenges.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "Start authentication error - no mechanism available",
			expectedError: errors.New("no mechanisms available for authentication"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_no_mechanism.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "Advanced authentication request error (UP challenge)",
			expectedError: errors.New("received non-200 response: 500"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_pass_only.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", "500 Internal Server Error", http.StatusInternalServerError)
			},
		},
		{
			name:          "Advanced authentication error (UP challenge) - JSON parsing failed",
			expectedError: errors.New("failed to parse response: invalid character 'o' in literal null (expecting 'u')"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_pass_only.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", "not a JSON response", http.StatusOK)
			},
		},
		{
			name:          "Successful authentication (SMS/Email MFA)",
			expectedToken: "valid-token",
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_mfa.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`, password)
				advanceAuthResponse, err := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				startOOBRequest := `{"Action":"StartOOB","MechanismId":"email_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`
				startOOBResponse, err := os.ReadFile("test/identity_mock/advance_start_oob.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", startOOBRequest, string(startOOBResponse), http.StatusOK)
				OOBSuccessRequest := `{"Action":"Answer","MechanismId":"email_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`
				OOBSuccessResponse, err := os.ReadFile("test/identity_mock/advance_oob_success.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", OOBSuccessRequest, string(OOBSuccessResponse), http.StatusOK)
				pollRequest := `{"Action":"Poll","MechanismId":"email_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`
				pollResponse, err := os.ReadFile("test/identity_mock/advance_oob_success.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", pollRequest, string(pollResponse), http.StatusOK)
			},
		},
		{
			name:          "Successful authentication (QR MFA)",
			expectedToken: "valid-token",
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_qr.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`, password)
				advanceAuthResponse, err := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				startPollResponse, err := os.ReadFile("test/identity_mock/advance_oob_success.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", string(startPollResponse), http.StatusOK)
			},
		},
		{
			name:          "Successful authentication (External Action MFA with PIN)",
			expectedToken: "valid-token",
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				OOBSuccessRequest := `{"Action":"Answer","MechanismId":"OOBAUTHPIN","SessionId":"idp_session_id","TenantId":"ABQ5234","Answer":"123456"}`
				OOBSuccessResponse, err := os.ReadFile("test/identity_mock/advance_oob_success.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", OOBSuccessRequest, string(OOBSuccessResponse), http.StatusOK)
			},
			responses: map[string]string{"Enter the pin code that you received in the browser": "123456"},
		},
		{
			name:          "Successful authentication (multiple security questions)",
			expectedToken: "valid-token",
			responses: map[string]string{
				"Q1": "A1",
				"Q2": "A2",
				"Q3": "A3",
			},
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_sq.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				answerRequest := fmt.Sprintf(`{"Action":"Answer","Answer":{"id1":"A1","id2":"A2","id3":"A3"},"MechanismId":"sq_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`)
				advanceAuthResponse, err := os.ReadFile("test/identity_mock/advance_auth_pass_only.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", answerRequest, string(advanceAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "MFA Authentication - Unsupported mechanism",
			expectedError: errors.New("unsupported authentication mechanism: UNSUPPORTED"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_unsupported_mfa.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`, password)
				advanceAuthResponse, err := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "MFA Authentication - Start OOB request error",
			expectedError: errors.New("received non-200 response: 500"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_mfa.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`, password)
				advanceAuthResponse, err := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				advanceRequest := `{"Action":"Answer","Answer":"password","MechanismId":"password_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", advanceRequest, "500 Internal Server Error", http.StatusInternalServerError)
			},
		},
		{
			name:          "MFA Authentication - advance authentication request error",
			expectedError: errors.New("failed to poll authentication status: received non-200 response: 500"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_mfa.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`, password)
				advanceAuthResponse, err := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				startOOBRequest := `{"Action":"StartOOB","MechanismId":"email_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`
				startOOBResponse, err := os.ReadFile("test/identity_mock/advance_start_oob.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", startOOBRequest, string(startOOBResponse), http.StatusOK)
				OOBFailureRequest := `{"Action":"Poll","MechanismId":"email_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", OOBFailureRequest, "500 Internal Server Error", http.StatusInternalServerError)
			},
		},
		{
			name:          "MFA Authentication - advance authentication error",
			timeout:       1 * time.Second,
			expectedError: errors.New("Timed out waiting for out-of-band authentication"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_mfa.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				nextChallengeRequest := fmt.Sprintf(`{"Action":"Answer","Answer":"%s","MechanismId":"password_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`, password)
				advanceAuthResponse, err := os.ReadFile("test/identity_mock/advance_next_challenge.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", nextChallengeRequest, string(advanceAuthResponse), http.StatusOK)
				startOOBRequest := `{"Action":"StartOOB","MechanismId":"email_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`
				startOOBResponse, err := os.ReadFile("test/identity_mock/advance_start_oob.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", startOOBRequest, string(startOOBResponse), http.StatusOK)
				OOBFailureRequest := `{"Action":"Poll","MechanismId":"email_mechanism_id","SessionId":"session_id","TenantId":"ABD5189"}`
				OOBFailureResponse, err := os.ReadFile("test/identity_mock/advance_failure.json")
				assert.NoError(t, err)
				mockResponseWithRequestBody(wiremockClient, "/Security/AdvanceAuthentication", OOBFailureRequest, string(OOBFailureResponse), http.StatusOK)
			},
		},
		{
			name:          "External action MFA - request error",
			expectedError: errors.New("received non-200 response: 500"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", "500 internal server error", http.StatusInternalServerError)
			},
		},
		{
			name:          "External action MFA - authentication failure",
			expectedError: errors.New("Authentication with federated identity provider failed: Authentication (login or challenge) has failed. Please try again or contact your system administrator."),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				advanceFailureResponse, err := os.ReadFile("test/identity_mock/advance_failure.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", string(advanceFailureResponse), http.StatusOK)
			},
		},
		{
			name:          "External action MFA - JSON parsing failed",
			expectedError: errors.New("failed to parse response: invalid character 'o' in literal null (expecting 'u')"),
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/AdvanceAuthentication", "not a JSON response", http.StatusOK)
			},
		},
		// Insecure login (no-PIN) test cases — mirrors the original pre-PIN waitForExternalAction flow
		{
			name:          "External action MFA without PIN - rejected when insecureLogin flag not set",
			expectedError: errors.New("oob auth pin required for login with external identity provider"),
			insecureLogin: false,
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_no_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
			},
		},
		{
			name:          "Successful authentication (External Action MFA without PIN)",
			expectedToken: "valid-token",
			insecureLogin: true,
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_no_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				OOBSuccessResponse, err := os.ReadFile("test/identity_mock/oob_status_success.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/OobAuthStatus", string(OOBSuccessResponse), http.StatusOK)
			},
		},
		{
			name:          "External action MFA without PIN - OobAuthStatus request error",
			expectedError: errors.New("received non-200 response: 500"),
			insecureLogin: true,
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_no_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/OobAuthStatus", "500 internal server error", http.StatusInternalServerError)
			},
		},
		{
			name:          "External action MFA without PIN - authentication failure",
			expectedError: errors.New("authentication failed"),
			insecureLogin: true,
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_no_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				OOBFailureResponse, err := os.ReadFile("test/identity_mock/oob_status_failure.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/OobAuthStatus", string(OOBFailureResponse), http.StatusOK)
			},
		},
		{
			name:          "External action MFA without PIN - JSON parsing failed",
			expectedError: errors.New("failed to parse response: invalid character 'o' in literal null (expecting 'u')"),
			insecureLogin: true,
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_no_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				mockResponse(wiremockClient, "/Security/OobAuthStatus", "not a JSON response", http.StatusOK)
			},
		},
		{
			name:          "External action MFA without PIN - timeout",
			expectedError: errors.New("Timed out waiting for external authentication."),
			insecureLogin: true,
			timeout:       1 * time.Second,
			beforeTest: func(t *testing.T) {
				startAuthResponse, err := os.ReadFile("test/identity_mock/start_auth_external_no_pin.json")
				assert.NoError(t, err)
				mockResponse(wiremockClient, "/Security/StartAuthentication", string(startAuthResponse), http.StatusOK)
				OOBPendingResponse := `{"Result":{"State":"pending"}}`
				mockResponse(wiremockClient, "/Security/OobAuthStatus", OOBPendingResponse, http.StatusOK)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear stubs on wiremock client
			defer wiremockClient.Clear()

			timeout := tc.timeout
			if timeout == 0 {
				timeout = defaultTimeout
			}

			authenticator := &IdentityAuthenticator{
				identityURL:   wiremockURL,
				timeout:       timeout,
				insecureLogin: tc.insecureLogin,
			}

			// Set up the test case
			if tc.beforeTest != nil {
				tc.beforeTest(t)
			}

			if tc.responses != nil {
				cleanup, in, out, done := mockStdio(t)
				defer func() {
					// Cleanup after mocking stdio
					cleanup()
					<-done
				}()
				for question, answer := range tc.responses {
					go func() {
						for range time.Tick(100 * time.Millisecond) {
							if strings.Contains(out.String(), question) {
								_, _ = fmt.Fprintln(in, answer)
								return
							}
						}
					}()
				}
			}

			// Call the method under test
			token, err := authenticator.GetToken(username, password)

			// Verify there were no unexpected requests
			requests, _ := wiremockClient.FindUnmatchedRequests()
			if requests != nil {
				assert.Empty(t, requests.Requests)
			}

			// Assert the results
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedToken, token)
		})
	}
}

func TestIdentityAuthenticator_OobAuthStatus_ForwardsCookies(t *testing.T) {
	const sessionCookie = "shard-session"
	const sessionCookieValue = "abc123"
	oobStatusCalled := false
	cookieForwarded := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Security/StartAuthentication":
			http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: sessionCookieValue})
			w.Header().Set("Content-Type", "application/json")
			startResp, _ := os.ReadFile("test/identity_mock/start_auth_external_no_pin.json")
			_, _ = w.Write(startResp)
		case "/Security/OobAuthStatus":
			oobStatusCalled = true
			c, err := r.Cookie(sessionCookie)
			cookieForwarded = err == nil && c.Value == sessionCookieValue
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Result":{"State":"success","Token":"valid-token"}}`))
		}
	}))
	defer srv.Close()

	jar, _ := cookiejar.New(nil)
	authenticator := &IdentityAuthenticator{
		identityURL:     srv.URL,
		timeout:         defaultTimeout,
		insecureLogin:   true,
		cookieJarClient: &http.Client{Jar: jar},
	}

	token, err := authenticator.GetToken(username, password)

	assert.NoError(t, err)
	assert.Equal(t, "valid-token", token)
	assert.True(t, oobStatusCalled, "OobAuthStatus should have been called")
	assert.True(t, cookieForwarded, "cookie set by StartAuthentication should be forwarded to OobAuthStatus")
}

func TestIdentityAuthenticator_AdvanceAuthentication_DoesNotForwardCookies(t *testing.T) {
	const sessionCookie = "shard-session"
	const sessionCookieValue = "abc123"
	cookieSentOnAdvance := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/Security/StartAuthentication":
			http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: sessionCookieValue})
			startResp, _ := os.ReadFile("test/identity_mock/start_auth_pass_only.json")
			_, _ = w.Write(startResp)
		case "/Security/AdvanceAuthentication":
			c, err := r.Cookie(sessionCookie)
			cookieSentOnAdvance = err == nil && c.Value == sessionCookieValue
			advanceResp, _ := os.ReadFile("test/identity_mock/advance_oob_success.json")
			_, _ = w.Write(advanceResp)
		}
	}))
	defer srv.Close()

	jar, _ := cookiejar.New(nil)
	authenticator := &IdentityAuthenticator{
		identityURL:     srv.URL,
		timeout:         defaultTimeout,
		cookieJarClient: &http.Client{Jar: jar},
	}

	_, _ = authenticator.GetToken(username, password)

	assert.False(t, cookieSentOnAdvance, "cookie from StartAuthentication should not be forwarded to AdvanceAuthentication")
}

func Test_promptMechChosen(t *testing.T) {
	tests := []struct {
		name      string
		mechanism *mechanismResp
		want      string
	}{{
		name:      "nil mechanism",
		mechanism: nil,
		want:      "",
	}, {
		name:      "unsupported mechanism",
		mechanism: &mechanismResp{Name: "UNSUPPORTED"},
		want:      "Please provide the required input for the selected authentication mechanism",
	}, {
		name:      "with prompt message",
		mechanism: &mechanismResp{PromptMechChosen: "Enter the PIN from your authenticator app"},
		want:      "Enter the PIN from your authenticator app",
	}, {
		name:      "without prompt message - SQ",
		mechanism: &mechanismResp{Name: "SQ"},
		want:      "Please answer your security question",
	}, {
		name:      "without prompt message - UP",
		mechanism: &mechanismResp{Name: "UP"},
		want:      "Please enter your password",
	}, {
		name:      "without prompt message - SMS",
		mechanism: &mechanismResp{Name: "SMS"},
		want:      "Please enter the code sent to your phone via SMS",
	}, {
		name:      "without prompt message - Email",
		mechanism: &mechanismResp{Name: "Email"},
		want:      "Please enter the code sent to your email",
	}, {
		name:      "without prompt message - OATH",
		mechanism: &mechanismResp{Name: "OATH"},
		want:      "Please enter your OATH one-time passcode",
	}, {
		name:      "without prompt message - OTP",
		mechanism: &mechanismResp{Name: "OTP"},
		want:      "Please enter the code from your identity mobile app",
	}, {
		name:      "without prompt message - U2F",
		mechanism: &mechanismResp{Name: "U2F"},
		want:      "Please complete the FIDO2 security key challenge",
	}, {}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, promptMechChosen(tt.mechanism), "promptMechChosen(%v)", tt.mechanism)
		})
	}
}

func Test_promptSelectMech(t *testing.T) {
	tests := []struct {
		name string
		mech mechanismResp
		want string
	}{
		{
			name: "with prompt message",
			mech: mechanismResp{PromptSelectMech: "Custom Prompt"},
			want: "Custom Prompt",
		},
		{
			name: "Security Question",
			mech: mechanismResp{Name: "SQ"},
			want: "Security Question",
		},
		{
			name: "Password",
			mech: mechanismResp{Name: "UP"},
			want: "Password",
		},
		{
			name: "SMS",
			mech: mechanismResp{Name: "SMS"},
			want: "SMS",
		},
		{
			name: "Email",
			mech: mechanismResp{Name: "EMAIL"},
			want: "Email",
		},
		{
			name: "OATH One-Time Passcode",
			mech: mechanismResp{Name: "OATH"},
			want: "OATH One-Time Passcode",
		},
		{
			name: "Identity Mobile App",
			mech: mechanismResp{Name: "OTP"},
			want: "Identity Mobile App",
		},
		{
			name: "FIDO2 Security Key",
			mech: mechanismResp{Name: "U2F"},
			want: "FIDO2 Security Key",
		},
		{
			name: "QR Code",
			mech: mechanismResp{Name: "QR"},
			want: "QR Code",
		},
		{
			name: "Phone Call",
			mech: mechanismResp{Name: "PF"},
			want: "Phone Call",
		},
		{
			name: "unsupported mechanism",
			mech: mechanismResp{Name: "UNSUPPORTED"},
			want: "UNSUPPORTED",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, promptSelectMech(tt.mech), "promptSelectMech(%v)", tt.mech)
		})
	}
}

func TestIdentityAuthenticator_startAuthentication_PodFQDNNormalization(t *testing.T) {
	passOnlyJSON, err := os.ReadFile("test/identity_mock/start_auth_pass_only.json")
	if err != nil {
		t.Fatal(err)
	}

	// srv is used for both the initial request and as the redirect target.
	// The "no protocol" case uses only the host:port of srv, relying on the fix to prepend https://.
	// We use a separate http server for the http:// case to avoid TLS issues.
	callCount := 0
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Return a PodFQDN that is just the host:port (no protocol) — should get https:// prepended.
			// The second request will fail to reach the https endpoint, but we only care about
			// the URL normalization, so we check identityURL before the second request errors.
			podFQDN := strings.TrimPrefix(srv.URL, "http://")
			redirectResp := fmt.Sprintf(`{"success":true,"Result":{"PodFQDN":%q},"IsSoftError":false}`, podFQDN)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(redirectResp))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(passOnlyJSON)
	}))
	defer srv.Close()

	t.Run("PodFQDN without protocol gets https:// prepended", func(t *testing.T) {
		callCount = 0
		ia := &IdentityAuthenticator{
			identityURL: srv.URL,
			timeout:     defaultTimeout,
		}
		// The second StartAuthentication call will fail (http server, not https),
		// but the URL normalization happens before the call, so we check identityURL after.
		_, _ = ia.startAuthentication(username)
		hostPort := strings.TrimPrefix(srv.URL, "http://")
		assert.Equal(t, "https://"+hostPort, ia.identityURL)
	})

	t.Run("PodFQDN with http:// gets upgraded to https://", func(t *testing.T) {
		callCount = 0
		srvInitial := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// Return a PodFQDN with http:// scheme — should be upgraded to https://.
				redirectResp := fmt.Sprintf(`{"success":true,"Result":{"PodFQDN":%q},"IsSoftError":false}`, srv.URL)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(redirectResp))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(passOnlyJSON)
		}))
		defer srvInitial.Close()

		ia := &IdentityAuthenticator{
			identityURL: srvInitial.URL,
			timeout:     defaultTimeout,
		}
		// The recursive call will fail (http server, not https), but normalization happens before it.
		_, _ = ia.startAuthentication(username)
		hostPort := strings.TrimPrefix(srv.URL, "http://")
		assert.Equal(t, "https://"+hostPort, ia.identityURL)
	})

	t.Run("PodFQDN with https:// is not modified", func(t *testing.T) {
		callCount = 0
		srvInitial := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// Return a PodFQDN that already has https:// — should pass through unchanged.
				redirectResp := fmt.Sprintf(`{"success":true,"Result":{"PodFQDN":%q},"IsSoftError":false}`, "https://pod.example.com")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(redirectResp))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(passOnlyJSON)
		}))
		defer srvInitial.Close()

		ia := &IdentityAuthenticator{
			identityURL: srvInitial.URL,
			timeout:     defaultTimeout,
		}
		// The recursive call will fail (unreachable host), but normalization happens before it.
		_, _ = ia.startAuthentication(username)
		assert.Equal(t, "https://pod.example.com", ia.identityURL)
	})
}

func mockStdio(t *testing.T) (func(), io.Writer, *bytes.Buffer, chan bool) {
	t.Helper()

	// Create pipes for stdin and stdout
	stdinR, stdinW, err := os.Pipe()
	assert.NoError(t, err)
	stdoutR, stdoutW, err := os.Pipe()
	assert.NoError(t, err)

	// Save original stdin, stdout, and stderr
	origStdin, origStdout, origStderr := os.Stdin, os.Stdout, os.Stderr

	// Redirect stdin, stdout, and stderr
	os.Stdin, os.Stdout, os.Stderr = stdinR, stdoutW, stdoutW

	cleanup := func() {
		os.Stdin, os.Stdout, os.Stderr = origStdin, origStdout, origStderr
		_, _, _ = stdinR.Close(), stdinW.Close(), stdoutW.Close()
	}

	done := make(chan bool)
	output := &bytes.Buffer{}
	go func() {
		defer close(done)
		_, err = io.Copy(output, stdoutR)
		stdoutR.Close()
	}()

	return cleanup, stdinW, output, done
}
