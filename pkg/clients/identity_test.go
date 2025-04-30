package clients

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strings"
	"testing"
)

type MockedHTTPClient struct {
	StartAuthFunc     func() (*http.Response, error)
	AdvancedAuthFunc  func() (*http.Response, error)
	OOBAuthStatusFunc func() (*http.Response, error)
}

func URLEndsWith(req *http.Request, suffix string) bool {
	return strings.HasSuffix(req.URL.String(), suffix)
}

func (m *MockedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	switch {
	case URLEndsWith(req, "/Security/StartAuthentication"):
		return m.StartAuthFunc()
	case URLEndsWith(req, "/Security/AdvanceAuthentication"):
		return m.AdvancedAuthFunc()
	case URLEndsWith(req, "/Security/OobAuthenticationStatus"):
		return m.OOBAuthStatusFunc()
	default:
		return nil, fmt.Errorf("unexpected URL: %s", req.URL.String())
	}
}

func TestIdentityAuthenticator_GetToken(t *testing.T) {
	testCases := []struct {
		name              string
		username          string
		password          string
		expectedToken     string
		expectedError     error
		startAuthFunc     func() (*http.Response, error)
		advancedAuthFunc  func() (*http.Response, error)
		oobAuthStatusFunc func() (*http.Response, error)
	}{
		{
			name:          "Successful authentication (UP challenge)",
			username:      "user@example.com",
			password:      "password",
			expectedToken: "valid-token",
			startAuthFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{	
	"success":true,
    "Result":{
      "ClientHints":{
        "PersistDefault":false,
        "AllowPersist":false,
        "AllowForgotPassword":false,
        "EndpointAuthenticationEnabled":false
      },
      "Version":"1.0",
      "SessionId":"8vwknPnxZ0ywsj95Y6-uK1kIr6XbtufC310N6qiE8Lc1",
      "EventDescription":null,
      "RetryWaitingTime":0,
      "SecurityImageName":null,
      "AllowLoginMfaCache":false,
      "Challenges":[
        {
          "Mechanisms":[
            {
              "AnswerType":"Text",
              "Name":"UP",
              "PromptMechChosen":"Enter Password",
              "PromptSelectMech":"Password",
              "MechanismId":"Vu71O06uC0K9uyAZjOlT5B50H63ftBR6vZPPHd5R80U1",
              "Enrolled":true
            }
          ]
        }
      ],
      "Summary":"NewPackage",
      "TenantId":"ABC1234"
    },
    "Message":null,
    "MessageID":null,
    "Exception":null,
    "ErrorID":null,
    "ErrorCode":null,
    "IsSoftError":false,
    "InnerExceptions":null}`)),
				}, nil
			},
			advancedAuthFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
	"success":true,
    "Result":{
      "AuthLevel":"Normal",
      "DisplayName":"User",
      "Token":"valid-token",
      "Auth":"valid-auth-token",
      "UserId":"some-user-id",
      "EmailAddress":"user@example.com",
      "UserDirectory":"CDS",
      "PodFqdn":"abc1234.id.integration-cyberark.cloud",
      "User":"user@example.cloud.371805",
      "CustomerID":"ABC1234",
      "SystemID":"ABC1234",
      "SourceDsType":"CDS",
      "Summary":"LoginSuccess"
    },
    "Message":null,
    "MessageID":null,
    "Exception":null,
    "ErrorID":null,
    "ErrorCode":null,
    "IsSoftError":false,
    "InnerExceptions":null}`)),
				}, nil
			},
		},
		{
			name:          "Start authentication request error",
			username:      "user@example.com",
			expectedError: errors.New("failed to send HTTP request: 500 Internal Server Error"),
			startAuthFunc: func() (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError}, errors.New("500 Internal Server Error")
			},
		},
		{
			name:          "Start authentication error - failed authentication",
			username:      "user@example.com",
			expectedError: errors.New("authentication failed: some error"),
			startAuthFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{	
	"success":false,
    "Message":"some error",
    "MessageID":null,
    "Exception":null,
    "ErrorID":null,
    "ErrorCode":null,
    "IsSoftError":false,
    "InnerExceptions":null}`)),
				}, nil
			},
		},
		{
			name:          "Start authentication error - JSON parsing failed",
			username:      "user@example.com",
			expectedError: errors.New("failed to parse response: invalid character 'o' in literal null (expecting 'u')"),
			startAuthFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`not a JSON response`)),
				}, nil
			},
		},
		{
			name:          "Start authentication error - no challenges available",
			username:      "user@example.com",
			expectedError: errors.New("no challenges available for authentication"),
			startAuthFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{	
	"success":true,
    "Result":{
      "ClientHints":{
        "PersistDefault":false,
        "AllowPersist":false,
        "AllowForgotPassword":false,
        "EndpointAuthenticationEnabled":false
      },
      "Version":"1.0",
      "SessionId":"8vwknPnxZ0ywsj95Y6-uK1kIr6XbtufC310N6qiE8Lc1",
      "EventDescription":null,
      "RetryWaitingTime":0,
      "SecurityImageName":null,
      "AllowLoginMfaCache":false,
      "Challenges":[],
      "Summary":"NewPackage",
      "TenantId":"ABC1234"
    },
    "Message":null,
    "MessageID":null,
    "Exception":null,
    "ErrorID":null,
    "ErrorCode":null,
    "IsSoftError":false,
    "InnerExceptions":null}`)),
				}, nil
			},
		},
		{
			name:          "Start authentication error - no mechanism available",
			username:      "user@example.com",
			expectedError: errors.New("no mechanisms available for authentication"),
			startAuthFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{	
	"success":true,
    "Result":{
      "ClientHints":{
        "PersistDefault":false,
        "AllowPersist":false,
        "AllowForgotPassword":false,
        "EndpointAuthenticationEnabled":false
      },
      "Version":"1.0",
      "SessionId":"8vwknPnxZ0ywsj95Y6-uK1kIr6XbtufC310N6qiE8Lc1",
      "EventDescription":null,
      "RetryWaitingTime":0,
      "SecurityImageName":null,
      "AllowLoginMfaCache":false,
      "Challenges":[
        {
          "Mechanisms":[]
        }
      ],
      "Summary":"NewPackage",
      "TenantId":"ABC1234"
    },
    "Message":null,
    "MessageID":null,
    "Exception":null,
    "ErrorID":null,
    "ErrorCode":null,
    "IsSoftError":false,
    "InnerExceptions":null}`)),
				}, nil
			},
		},
		{
			name:          "Advanced authentication request error (UP challenge)",
			username:      "user@example.com",
			password:      "password",
			expectedError: errors.New("failed to send HTTP request: 500 Internal Server Error"),
			startAuthFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{	
	"success":true,
    "Result":{
      "ClientHints":{
        "PersistDefault":false,
        "AllowPersist":false,
        "AllowForgotPassword":false,
        "EndpointAuthenticationEnabled":false
      },
      "Version":"1.0",
      "SessionId":"8vwknPnxZ0ywsj95Y6-uK1kIr6XbtufC310N6qiE8Lc1",
      "EventDescription":null,
      "RetryWaitingTime":0,
      "SecurityImageName":null,
      "AllowLoginMfaCache":false,
      "Challenges":[
        {
          "Mechanisms":[
            {
              "AnswerType":"Text",
              "Name":"UP",
              "PromptMechChosen":"Enter Password",
              "PromptSelectMech":"Password",
              "MechanismId":"Vu71O06uC0K9uyAZjOlT5B50H63ftBR6vZPPHd5R80U1",
              "Enrolled":true
            }
          ]
        }
      ],
      "Summary":"NewPackage",
      "TenantId":"ABC1234"
    },
    "Message":null,
    "MessageID":null,
    "Exception":null,
    "ErrorID":null,
    "ErrorCode":null,
    "IsSoftError":false,
    "InnerExceptions":null}`)),
				}, nil
			},
			advancedAuthFunc: func() (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError}, errors.New("500 Internal Server Error")
			},
		},
		{
			name:          "Advanced authentication error (UP challenge) - JSON parsing failed",
			username:      "user@example.com",
			password:      "password",
			expectedError: errors.New("failed to parse response: invalid character 'o' in literal null (expecting 'u')"),
			startAuthFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{	
	"success":true,
    "Result":{
      "ClientHints":{
        "PersistDefault":false,
        "AllowPersist":false,
        "AllowForgotPassword":false,
        "EndpointAuthenticationEnabled":false
      },
      "Version":"1.0",
      "SessionId":"8vwknPnxZ0ywsj95Y6-uK1kIr6XbtufC310N6qiE8Lc1",
      "EventDescription":null,
      "RetryWaitingTime":0,
      "SecurityImageName":null,
      "AllowLoginMfaCache":false,
      "Challenges":[
        {
          "Mechanisms":[
            {
              "AnswerType":"Text",
              "Name":"UP",
              "PromptMechChosen":"Enter Password",
              "PromptSelectMech":"Password",
              "MechanismId":"Vu71O06uC0K9uyAZjOlT5B50H63ftBR6vZPPHd5R80U1",
              "Enrolled":true
            }
          ]
        }
      ],
      "Summary":"NewPackage",
      "TenantId":"ABC1234"
    },
    "Message":null,
    "MessageID":null,
    "Exception":null,
    "ErrorID":null,
    "ErrorCode":null,
    "IsSoftError":false,
    "InnerExceptions":null}`)),
				}, nil
			},
			advancedAuthFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`not a JSON response`)),
				}, nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockedClient := &MockedHTTPClient{
				StartAuthFunc:     tc.startAuthFunc,
				AdvancedAuthFunc:  tc.advancedAuthFunc,
				OOBAuthStatusFunc: tc.oobAuthStatusFunc,
			}

			authenticator := &IdentityAuthenticator{
				httpClient: mockedClient,
			}

			token, err := authenticator.GetToken(tc.username, tc.password)
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
			assert.Equal(t, tc.expectedToken, token)
		})
	}
}
