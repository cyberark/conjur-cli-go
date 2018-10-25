package cmd

import (
	"fmt"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/spf13/afero"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

// AuthnLoginOptions are the options used to authentification.
type AuthnLoginOptions struct {
	Config   conjurapi.Config
	Username string
	Password string
}

// AuthnLoginer performs the operations necessary to authentification as described by options.
type AuthnLoginer interface {
	Do(options AuthnLoginOptions) error
}

// NewAuthnLoginer returns a AuthnLoginer that uses client to Authentification.
func NewAuthnLoginer(client AuthnClient, fs afero.Fs) AuthnLoginer {
	return loginerImpl{
		AuthnClient: client,
		Fs:          fs,
	}
}

type loginerImpl struct {
	AuthnClient
	afero.Fs
}

func (l loginerImpl) Do(options AuthnLoginOptions) (err error) {
	login := options.Username
	apiKey, err := l.Login(login, options.Password)
	if err != nil {
		return
	}

	config := l.GetConfig()
	rc := new(netrc.Netrc)
	if rcFile, err := l.Open(config.NetRCPath); err == nil {
		defer rcFile.Close()

		var parseErr error
		rc, parseErr = netrc.Parse(rcFile)
		if err != nil {
			return parseErr
		}
	}

	url := config.ApplianceURL + "/authn"
	m := rc.FindMachine(url)
	if m != nil {
		m.UpdateLogin(login)
		m.UpdatePassword(string(apiKey))
	} else {
		rc.NewMachine(url, login, string(apiKey), "")
	}

	newRc, err := rc.MarshalText()
	if err != nil {
		return
	}

	return afero.WriteFile(l, string(config.NetRCPath), []byte(fmt.Sprintf("%s", newRc)), 0600)
}
