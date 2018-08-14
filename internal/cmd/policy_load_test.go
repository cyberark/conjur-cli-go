package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/internal/cmd"
	"github.com/cyberark/conjur-cli-go/internal/cmd/mocks"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
)

func TestLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockConjurClient(ctrl)
	expectedResponse := conjurapi.PolicyResponse{
		Version:      1,
		CreatedRoles: map[string]conjurapi.CreatedRole{},
	}
	policy := []byte("[!variable var]")
	mockClient.EXPECT().LoadPolicy(conjurapi.PolicyModePost, "root", bytes.NewReader(policy)).Return(&expectedResponse, nil)

	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "policy.yml", policy, 0644); err != nil {
		t.Fatalf("WriteFile failed, %v", err)
	}

	options := cmd.PolicyLoadOptions{
		Delete:   false,
		Filename: "policy.yml",
		PolicyID: "root",
		Replace:  false,
	}

	loader := cmd.NewPolicyLoader(mockClient, fs)

	jsonResponse, err := loader.Do(options)
	if err != nil {
		t.Fatalf("Policy failed, %v", err)
	}

	marshaledResponse, err := json.MarshalIndent(expectedResponse, "", "  ")
	if err != nil {
		t.Fatalf("failed marshaling policy response, %v", err)
	}

	if string(jsonResponse) != string(marshaledResponse) {
		t.Fatalf("LoadPolicy returned the wrong response")
	}
}
