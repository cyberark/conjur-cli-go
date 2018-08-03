package action_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/action"
	"github.com/cyberark/conjur-cli-go/action/mocks"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
)

func TestLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockConjurClient(ctrl)
	expectedResponse := conjurapi.PolicyResponse{Version: 1, CreatedRoles: map[string]conjurapi.CreatedRole{}}
	mockClient.EXPECT().LoadPolicy(conjurapi.PolicyModePost, "root", bytes.NewReader([]byte("- !variable var"))).Return(&expectedResponse, nil)

	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "policy.yml", []byte("- !variable var"), 0644); err != nil {
		t.Fatalf("WriteFile failed, %v", err)
	}

	jsonResponse, err := action.Policy{Fs: fs, PolicyID: "root"}.Load(mockClient, "policy.yml", false, false)
	if err != nil {
		t.Fatalf("Policy failed, %v", err)
	}

	marshaledResponse, err := json.MarshalIndent(expectedResponse, "", "  ")
	if err != nil {
		t.Fatalf("failed marshaling policy response, %v", err)
	}
	expectedJSON := string(marshaledResponse)

	if jsonResponse != expectedJSON {
		t.Fatalf("LoadPolicy returned the wrong response")
	}
}
