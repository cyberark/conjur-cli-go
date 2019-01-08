package cmd

import (
	"encoding/base64"
	"encoding/json"

	"fmt"
)

type AuthnAuthenticatorOptions struct {
	AsHeader bool
}

type AuthnAuthenticator interface {
	Do(options AuthnAuthenticatorOptions) ([]byte, error)
}

func NewAuthnAuthenticator(client AuthnClient) AuthnAuthenticator {
	return authenticatorImpl{
		AuthnClient: client,
	}
}

type authenticatorImpl struct {
	AuthnClient
}

func (l authenticatorImpl) Do(options AuthnAuthenticatorOptions) (token []byte, err error) {
	err = l.RefreshToken()
	if err != nil {
		return
	}

	if options.AsHeader {
		token = []byte(fmt.Sprintf("Authorization: Token token=\"%s\"", base64.StdEncoding.EncodeToString(l.GetAuthToken().Raw())))
		return
	}

	return json.MarshalIndent(l.GetAuthToken(), "", "")
}
