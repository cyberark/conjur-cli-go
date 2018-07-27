package action

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/cyberark/conjur-api-go/conjurapi"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// PolicyClient specifies the Conjur API methods required to implement the policy actions.
type PolicyClient interface {
	LoadPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.PolicyResponse, error)
}

// Policy contains the information required to perform policy actions.
type Policy struct {
	Fs       afero.Fs
	PolicyID string
}

func (p Policy) readFrom(filename string) (body []byte, err error) {
	if filename == "-" {
		return ioutil.ReadAll(os.Stdin)
	}

	in, err := p.Fs.Open(filename)
	if err != nil {
		return
	}
	defer func() {
		if cerr := in.Close(); cerr != nil {
			log.Warnf("Failed closing %s, %s", cerr.Error())
		}
	}()
	return ioutil.ReadAll(in)
}

// Load uses the Conjur API to load a policy.
func (p Policy) Load(client PolicyClient, filename string, delete, replace bool) (policyResp string, err error) {
	mode := conjurapi.PolicyModePost
	switch {
	case delete:
		mode = conjurapi.PolicyModePatch
	case replace:
		mode = conjurapi.PolicyModePut
	}

	body, err := p.readFrom(filename)
	if err != nil {
		return
	}

	resp, err := client.LoadPolicy(mode, p.PolicyID, bytes.NewReader(body))
	if err != nil {
		return
	}

	jsonResp, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return
	}

	policyResp = string(jsonResp)
	return
}
