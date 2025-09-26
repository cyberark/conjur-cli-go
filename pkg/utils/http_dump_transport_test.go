package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDumpTransport(t *testing.T) {
	reqTestCases := []struct {
		description string
		path        string
		headers     map[string]string
		body        string
		assert      func(t *testing.T, req *http.Request, dump string)
	}{
		{
			description: "Only Authz header is redacted",
			headers: map[string]string{
				"Authorization": "some-token",
				"Other-Header":  "other-value",
			},
			assert: func(t *testing.T, req *http.Request, dump string) {
				assert.NotContains(t, dump, "some-token")
				assert.Contains(t, dump, "other-value")

				assert.Equal(t, req.Header.Get("Authorization"), "some-token")
				assert.Equal(t, req.Header.Get("Other-Header"), "other-value")
			},
		},
		{
			description: "Request body is redacted on authentication requests",
			// Specifically, ensure that if the request also contains "issuer/", it is
			// still redacted.
			path: "/authn-xyz/account/login/host/issuers/authenticate",
			body: "some-body",
			assert: func(t *testing.T, req *http.Request, dump string) {
				assert.Contains(t, dump, redactedString)
				assert.NotContains(t, dump, "some-body")

				reqBody, err := io.ReadAll(req.Body)
				assert.Nil(t, err)
				assert.Equal(t, string(reqBody), "some-body")
			},
		},
		{
			description: "Request body is redacted on issuer request with the data key",
			path:        "/issuers/create",
			body:        `{ "data": { "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" } }`,
			assert: func(t *testing.T, req *http.Request, dump string) {
				assert.Contains(t, dump, redactedString)
				assert.NotContains(t, dump, `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`)

				reqBody, err := io.ReadAll(req.Body)
				assert.Nil(t, err)
				assert.Equal(t, string(reqBody), `{ "data": { "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" } }`)
			},
		},
		{
			description: "Request body is maintained  on issuer requests without the data key",
			path:        "/issuers/update",
			body:        `{ "id": test-id", "max_ttl": 3001}`,
			assert: func(t *testing.T, req *http.Request, dump string) {
				assert.Contains(t, dump, `{ "id": test-id", "max_ttl": 3001}`)
			},
		},
		{
			description: "Request body is maintained on other requests",
			body:        "some-body",
			assert: func(t *testing.T, req *http.Request, dump string) {
				assert.Contains(t, dump, "some-body")
			},
		},
	}

	for _, tc := range reqTestCases {
		t.Run(tc.description, func(t *testing.T) {
			req, err := http.NewRequest(
				"POST",
				fmt.Sprintf("http://somehost.com%s", tc.path),
				bytes.NewBuffer([]byte(tc.body)),
			)
			assert.Nil(t, err)
			for k, v := range tc.headers {
				req.Header.Add(k, v)
			}

			dump := NewDumpTransport(nil, nil).dumpRequest(req)
			tc.assert(t, req, string(dump))
		})
	}

	respTestCases := []struct {
		description string
		body        string
		path        string
		assert      func(t *testing.T, res *http.Response, dump string)
	}{
		{
			description: "Body is redacted if it contains a Conjur token",
			body:        "{\"protected\":\"abcde\",\"payload\":\"fghijk\",\"signature\":\"lmnop\"}",
			assert: func(t *testing.T, res *http.Response, dump string) {
				assert.Contains(t, dump, redactedString)
				assert.NotContains(t, dump, "{\"protected\":\"abcde\",\"payload\":\"fghijk\",\"signature\":\"lmnop\"}")

				reqBody, err := io.ReadAll(res.Body)
				assert.Nil(t, err)
				assert.Contains(t, string(reqBody), "{\"protected\":\"abcde\",\"payload\":\"fghijk\",\"signature\":\"lmnop\"}")
			},
		},
		{
			description: "Body is redacted if part of identity flow",
			path:        "/Security/StartAuthentication",
			body:        "{\"some\":\"body\"}",
			assert: func(t *testing.T, res *http.Response, dump string) {
				assert.Contains(t, dump, redactedString)
			},
		},
		{
			description: "Body is maintained otherwise",
			body:        "some-body",
			assert: func(t *testing.T, res *http.Response, dump string) {
				assert.Contains(t, dump, "some-body")
			},
		},
	}

	for _, tc := range respTestCases {
		t.Run(tc.description, func(t *testing.T) {
			resp := http.Response{
				Body:    io.NopCloser(bytes.NewBufferString(tc.body)),
				Request: &http.Request{URL: &url.URL{Path: tc.path}},
			}

			dump := NewDumpTransport(nil, nil).dumpResponse(&resp)
			tc.assert(t, &resp, string(dump))
		})
	}
}

func Test_redactCookies(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
	}{{
		"empty",
		http.Header{},
	}, {
		"No Set-Cookie header",
		http.Header{"Some-Header": []string{"some-value"}},
	}, {
		"With Set-Cookie header",
		http.Header{
			"Set-Cookie":  []string{"some-cookie=some-value; Path=/; HttpOnly"},
			"Some-Header": []string{"some-value"},
		}}, {
		"Multiple Set-Cookie headers",
		http.Header{
			"Set-Cookie":  []string{"cookie1=value1; Path=/; HttpOnly", "cookie2=value2; Path=/; HttpOnly"},
			"Some-Header": []string{"some-value"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &http.Response{}
			res.Header = tt.headers
			cleanup := redactCookies(res)
			if len(tt.headers.Values(setCookieHeader)) == 0 {
				assert.Empty(t, res.Header.Values(setCookieHeader))
			} else {
				assert.Equal(t, res.Header.Get(setCookieHeader), redactedString)
			}
			cleanup()
			assert.Equal(t, tt.headers, res.Header)
		})
	}
}
