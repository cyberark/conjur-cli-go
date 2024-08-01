package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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
			path:        "/authn-xyz/account/login",
			body:        "some-body",
			assert: func(t *testing.T, req *http.Request, dump string) {
				assert.Contains(t, dump, redactedString)
				assert.NotContains(t, dump, "some-body")

				reqBody, err := io.ReadAll(req.Body)
				assert.Nil(t, err)
				assert.Equal(t, string(reqBody), "some-body")
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
				Body: io.NopCloser(bytes.NewBufferString(tc.body)),
			}

			dump := NewDumpTransport(nil, nil).dumpResponse(&resp)
			tc.assert(t, &resp, string(dump))
		})
	}
}
