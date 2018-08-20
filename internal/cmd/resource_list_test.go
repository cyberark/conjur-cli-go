package cmd_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

// makeJSONVars returns the requested number of variable ids, as well
// as an array of JSON objects containing the ids
func makeJSONVars(n int) (ids []string, objs []string) {
	for i := 1; i <= n; i++ {
		ids = append(ids, fmt.Sprintf("cucumber:variable:var%d", i))
	}

	for _, o := range ids {
		objs = append(objs, fmt.Sprintf(`{"id": "%s"}`, o))
	}
	return
}

func TestList(t *testing.T) {
	ctrl, mockClient := testSetup(t)
	defer ctrl.Finish()

	expectedIds, expectedObjs := makeJSONVars(2)

	expectedResources := make([]map[string]interface{}, 1)
	err := json.Unmarshal([]byte(fmt.Sprintf("[%s]", strings.Join(expectedObjs, ","))), &expectedResources)
	if err != nil {
		t.Fatalf("Unmarshal failed, %v", err)
	}
	mockClient.EXPECT().Resources(&conjurapi.ResourceFilter{}).Return(expectedResources, nil)

	lister := cmd.NewResourceLister(mockClient)

	resources, err := lister.Do(cmd.ResourceListOptions{})
	if err != nil {
		t.Fatalf("List failed, %v", err)
	}

	expectedJSON, _ := json.MarshalIndent(expectedIds, "", "  ")
	if string(resources) != string(expectedJSON) {
		t.Fatalf("Got %v, want %v", string(resources), string(expectedJSON))
	}
}
