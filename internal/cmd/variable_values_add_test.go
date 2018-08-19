package cmd_test

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
	"github.com/cyberark/conjur-cli-go/internal/cmd/mocks"
)

func TestValuesAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := "var"
	newValue := "value"

	mockClient := mocks.NewMockConjurClient(ctrl)
	mockClient.EXPECT().AddSecret(id, newValue).Return(nil)

	options := cmd.VariableValuesAddOptions{
		ID:    id,
		Value: newValue,
	}
	adder := cmd.NewVariableValuesAdder(mockClient)

	err := adder.Do(options)
	if err != nil {
		t.Fatalf("ValuesAdd failed, %v", err)
	}
}
