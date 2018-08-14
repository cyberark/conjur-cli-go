package cmd

import (
	"encoding/json"
)

// ResourceShowOptions are the options that control the resource shown.
type ResourceShowOptions struct {
	ResourceID string
}

// NewResourceShower returns a ResourceShower implementation.
func NewResourceShower(client ResourceClient) ResourceShower {
	return showImpl{
		ResourceClient: client,
	}
}

// ResourceShower uses the client to return the single resource described by options.
type ResourceShower interface {
	Do(options ResourceShowOptions) (resp []byte, err error)
}

type showImpl struct {
	ResourceClient
}

func (r showImpl) Do(options ResourceShowOptions) (resp []byte, err error) {
	resource, err := r.Resource(options.ResourceID)
	if err != nil {
		return
	}

	return json.MarshalIndent(resource, "", "  ")
}
