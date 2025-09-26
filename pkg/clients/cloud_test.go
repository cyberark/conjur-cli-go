package clients

import (
	"fmt"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi/response"
	"github.com/stretchr/testify/assert"
)

func Test_extractHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{{
		"401",
		&response.ConjurError{
			Code: 401,
		},
		401,
	}, {
		"no status code",
		fmt.Errorf("some other error"),
		0,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, extractHTTPStatusCode(tt.err), "extractHTTPStatusCode(%v)", tt.err)
		})
	}
}

func TestParseCloudURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *CloudURL
		wantErr assert.ErrorAssertionFunc
	}{{
		"valid URL",
		"https://mytenant.secretsmgr.cyberark.cloud/api",
		&CloudURL{Tenant: "mytenant", Suffix: ".cyberark.cloud"},
		assert.NoError,
	}, {
		"invalid subdomain",
		"https://mytenant.secretsmgr.invalid.domain/api",
		&CloudURL{Tenant: "mytenant", Suffix: ".invalid.domain"},
		func(t assert.TestingT, err error, i ...interface{}) bool {
			return assert.Error(t, err) && assert.Contains(t, err.Error(), "unrecognized domain suffix")
		},
	}, {
		"missing secretsmgr",
		"https://mytenant.cyberark.cloud/api",
		&CloudURL{Tenant: "mytenant", Suffix: ".cyberark.cloud"},
		func(t assert.TestingT, err error, i ...interface{}) bool {
			return assert.Error(t, err) && assert.Contains(t, err.Error(), "https://<tenant>.secretsmgr.cyberark.cloud/api")
		},
	}, {
		"missing api",
		"https://mytenant.secretsmgr.cyberark.cloud",
		nil,
		assert.Error,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCloudURL(tt.url)
			if !tt.wantErr(t, err, fmt.Sprintf("ParseCloudURL(%v)", tt.url)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ParseCloudURL(%v)", tt.url)
		})
	}
}
