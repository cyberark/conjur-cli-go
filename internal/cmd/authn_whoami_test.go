package cmd_test

import (
	"encoding/json"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

func TestWhoAmI(t *testing.T) {
	ctrl, mockClient := testSetup(t)
	defer ctrl.Finish()

	mockClient.EXPECT().RefreshToken()

	configResponse := conjurapi.Config{
		Account: "cucumber",
	}
	mockClient.EXPECT().GetConfig().Return(configResponse)

	whoamier := cmd.NewWhoAmIer(mockClient)
	jsonResponse, err := whoamier.Do()
	if err != nil {
		t.Fatalf("WhoAmI failed, %v", err)
	}

	expectedResponse := cmd.WhoAmIResponse{
		Account: "cucumber",
	}
	marshaledResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		t.Fatalf("failed marshaling policy response, %v", err)
	}

	if string(jsonResponse) != string(marshaledResponse) {
		t.Fatalf("WhoAmI returned the wrong response")
	}
}
