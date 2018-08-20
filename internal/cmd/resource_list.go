package cmd

import (
	"encoding/json"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

// ResourceListOptions are the options that control the resources returned.
type ResourceListOptions struct {
	Inspect bool
	Kind    string
}

// NewResourceLister returns a ResourceLister implementation.
func NewResourceLister(client ResourceClient) ResourceLister {
	return listerImpl{
		ResourceClient: client,
	}
}

// ResourceLister uses the client to return the resources described by options.
type ResourceLister interface {
	Do(options ResourceListOptions) (resp []byte, err error)
}

type listerImpl struct {
	ResourceClient
}

func (r listerImpl) Do(options ResourceListOptions) (resp []byte, err error) {
	filter := &conjurapi.ResourceFilter{
		Kind: options.Kind,
	}

	resources, err := r.Resources(filter)
	if err != nil {
		return
	}

	if !options.Inspect {
		var extractedIds []string
		for _, r := range resources {
			extractedIds = append(extractedIds, r["id"].(string))
		}
		return json.MarshalIndent(extractedIds, "", "  ")
	}

	return json.MarshalIndent(resources, "", "  ")
}
