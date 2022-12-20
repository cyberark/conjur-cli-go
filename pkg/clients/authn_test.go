package clients

import (
	"errors"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/stretchr/testify/assert"
)

type mockOidcConjurClient struct {
	providers   []conjurapi.OidcProvider
	expectedErr error
}

func (m mockOidcConjurClient) ListOidcProviders() ([]conjurapi.OidcProvider, error) {
	if m.expectedErr != nil {
		return nil, m.expectedErr
	}
	return m.providers, nil
}

func TestGetOidcProviderInfo(t *testing.T) {
	testCases := []struct {
		name        string
		providers   []conjurapi.OidcProvider
		serviceID   string
		expectedErr error
		assert      func(t *testing.T, provider *conjurapi.OidcProvider, err error)
	}{
		{
			name: "returns error when provider is not found",
			providers: []conjurapi.OidcProvider{
				{
					ServiceID: "service-id-1",
				},
			},
			serviceID: "service-id-2",
			assert: func(t *testing.T, providers *conjurapi.OidcProvider, err error) {
				assert.EqualError(t, err, "OIDC provider with service ID service-id-2 not found")
			},
		},
		{
			name: "returns provider when provider is found",
			providers: []conjurapi.OidcProvider{
				{
					ServiceID: "service-id-1",
				},
			},
			serviceID: "service-id-1",
			assert: func(t *testing.T, providers *conjurapi.OidcProvider, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "service-id-1", providers.ServiceID)
			},
		},
		{
			name:        "returns error when list providers returns error",
			providers:   []conjurapi.OidcProvider{},
			serviceID:   "service-id-1",
			expectedErr: errors.New("Failed to list OIDC providers"),
			assert: func(t *testing.T, providers *conjurapi.OidcProvider, err error) {
				assert.EqualError(t, err, "Failed to list OIDC providers")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conjurClient := mockOidcConjurClient{providers: tc.providers, expectedErr: tc.expectedErr}
			provider, err := getOidcProviderInfo(conjurClient, tc.serviceID)
			tc.assert(t, provider, err)
		})
	}
}
