package storage

import (
	"errors"
	"os"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/cyberark/conjur-api-go/conjurapi"
)

func getMachineName(config conjurapi.Config) string {
	return config.ApplianceURL + "/authn"
}

// StoreCredentials stores credentials to the specified .netrc file
func StoreCredentials(config conjurapi.Config, login string, apiKey string) error {
	machineName := getMachineName(config)
	filePath := config.NetRCPath

	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.WriteFile(filePath, []byte{}, 0644)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	nrc, err := netrc.ParseFile(filePath)
	if err != nil {
		return err
	}

	m := nrc.FindMachine(machineName)
	if m == nil || m.IsDefault() {
		_ = nrc.NewMachine(machineName, login, apiKey, "")
	} else {
		m.UpdateLogin(login)
		m.UpdatePassword(apiKey)
	}

	data, err := nrc.MarshalText()
	if err != nil {
		return err
	}

	if data[len(data)-1] != byte('\n') {
		data = append(data, byte('\n'))
	}

	return os.WriteFile(filePath, data, 0644)
}

// PurgeCredentials purges credentials from the specified .netrc file
func PurgeCredentials(config conjurapi.Config) error {
	// Remove cached access token (in case user logged in with OIDC which saves the access token to a file)
	if config.OidcTokenPath == "" {
		config.OidcTokenPath = os.ExpandEnv("$HOME/.conjur/oidc_token")
	}
	os.Remove(config.OidcTokenPath)

	// Remove cached credentials (username, api key) from .netrc
	machineName := getMachineName(config)
	filePath := config.NetRCPath

	nrc, err := netrc.ParseFile(filePath)
	if err != nil {
		// If the .netrc file doesn't exist, we don't need to do anything
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		// Any other error should be returned
		return err
	}

	nrc.RemoveMachine(machineName)

	data, err := nrc.MarshalText()
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// TODO: Should we stat for PurgeCredentials and StoreCredentials
//  and make sure we retain the permissions the file originally had ?
