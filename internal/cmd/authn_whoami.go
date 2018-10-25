package cmd

import (
	"encoding/json"
)

// WhoAmIResponse is the response of WhoAmI command.
type WhoAmIResponse struct {
	Account string `json:"account"`
}

// NewWhoAmIer returns a WhoAmIer implementation.
func NewWhoAmIer(client AuthnClient) WhoAmIer {
	return whoAmIImpl{
		AuthnClient: client,
	}
}

// WhoAmIer uses the client to return the WhoAmIResponse.
type WhoAmIer interface {
	Do() (resp []byte, err error)
}

type whoAmIImpl struct {
	AuthnClient
}

func (r whoAmIImpl) Do() (resp []byte, err error) {
	err = r.RefreshToken()
	if err != nil {
		return
	}

	response := WhoAmIResponse{
		Account: r.GetConfig().Account,
	}

	return json.Marshal(response)
}
