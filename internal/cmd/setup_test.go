package cmd_test

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cyberark/conjur-cli-go/internal/cmd/mocks"
)

func testSetup(t *testing.T) (ctrl *gomock.Controller, mockClient *mocks.MockConjurClient) {
	ctrl = gomock.NewController(t)

	mockClient = mocks.NewMockConjurClient(ctrl)

	return
}
