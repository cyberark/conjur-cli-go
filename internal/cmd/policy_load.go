package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/afero"
)

// PolicyLoadOptions are the options used to load a policy.
type PolicyLoadOptions struct {
	Delete   bool
	Filename string
	PolicyID string
	Replace  bool
}

// PolicyLoader performs the operations necessary to load the policy as described by options.
type PolicyLoader interface {
	Do(options PolicyLoadOptions) ([]byte, error)
}

// NewPolicyLoader returns a PolicyLoader that uses client and fs to load a policy.
func NewPolicyLoader(client PolicyClient, fs afero.Fs) PolicyLoader {
	return loaderImpl{
		PolicyClient: client.(PolicyClient),
		Fs:           fs,
	}
}

type loaderImpl struct {
	PolicyClient
	afero.Fs
}

func (p loaderImpl) readFrom(filename string) (body []byte, err error) {
	if filename == "-" {
		return ioutil.ReadAll(os.Stdin)
	}

	in, err := p.Open(filename)
	if err != nil {
		return
	}
	defer in.Close()

	return ioutil.ReadAll(in)
}

func (p loaderImpl) Do(options PolicyLoadOptions) (resp []byte, err error) {
	mode := conjurapi.PolicyModePost
	switch {
	case options.Delete:
		mode = conjurapi.PolicyModePatch
	case options.Replace:
		mode = conjurapi.PolicyModePut
	}

	body, err := p.readFrom(options.Filename)
	if err != nil {
		return
	}

	loadResp, err := p.LoadPolicy(mode, options.PolicyID, bytes.NewReader(body))
	if err != nil {
		return
	}

	return json.MarshalIndent(loadResp, "", "  ")
}
