package cmd_test

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
	"github.com/cyberark/conjur-cli-go/internal/cmd/mocks"
)

func TestValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := "var"
	expectedValue := string("value")
	mockClient := mocks.NewMockConjurClient(ctrl)
	mockClient.EXPECT().RetrieveSecret(id).Return([]byte(expectedValue), nil)

	options := cmd.VariableValueOptions{
		ID: id,
	}
	valuer := cmd.NewVariableValuer(mockClient)

	value, err := valuer.Do(options)
	if err != nil {
		t.Fatalf("Value failed, %v", err)
	}
	if string(value) != expectedValue {
		t.Fatalf("Got '%v', want '%v'", value, expectedValue)
	}
}
