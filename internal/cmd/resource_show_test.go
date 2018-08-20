package cmd_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

func TestShow(t *testing.T) {
	ctrl, mockClient := testSetup(t)
	defer ctrl.Finish()

	id := "cucumber:variable:var1"
	expectedResource := make(map[string]interface{}, 1)
	err := json.Unmarshal([]byte(fmt.Sprintf(`{"id": "%s"}`, id)), &expectedResource)
	if err != nil {
		t.Fatal(err)
	}
	mockClient.EXPECT().Resource(id).Return(expectedResource, nil)

	shower := cmd.NewResourceShower(mockClient)
	options := cmd.ResourceShowOptions{
		ResourceID: id,
	}
	resources, err := shower.Do(options)
	if err != nil {
		t.Fatalf("List failed, %v", err)
	}

	expectedJSON, _ := json.MarshalIndent(expectedResource, "", "  ")
	if string(resources) != string(expectedJSON) {
		t.Fatalf("Got %v, want %v", string(resources), string(expectedJSON))
	}
}
