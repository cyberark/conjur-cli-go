package action_test

import (
	"testing"

	"github.com/cyberark/conjur-cli-go/action"
	"github.com/cyberark/conjur-cli-go/action/mocks"
	"github.com/golang/mock/gomock"
)

func TestValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expected := string("value")
	mockClient := mocks.NewMockConjurClient(ctrl)
	mockClient.EXPECT().RetrieveSecret("var").Return([]byte(expected), nil)

	value, err := action.Variable{Name: "var"}.Value(mockClient)
	if err != nil {
		t.Fatalf("Value failed, %v", err)
	}
	if string(value) != expected {
		t.Fatalf("Got '%v', want '%v'", value, expected)
	}
}

func TestValuesAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	newValue := "value"

	mockClient := mocks.NewMockConjurClient(ctrl)
	mockClient.EXPECT().AddSecret("var", newValue).Return(nil)

	err := action.Variable{Name: "var"}.ValuesAdd(mockClient, newValue)
	if err != nil {
		t.Fatalf("ValuesAdd failed, %v", err)
	}
}
