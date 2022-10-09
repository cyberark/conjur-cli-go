package storage

import (
	"errors"
	"io/ioutil"
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
			err = ioutil.WriteFile(filePath, []byte{}, 0644)
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

	return ioutil.WriteFile(filePath, data, 0644)
}

// PurgeCredentials purges credentials from the specified .netrc file
func PurgeCredentials(config conjurapi.Config) error {
	machineName := getMachineName(config)
	filePath := config.NetRCPath

	nrc, err := netrc.ParseFile(filePath)
	if err != nil {
		return err
	}

	nrc.RemoveMachine(machineName)

	data, err := nrc.MarshalText()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, data, 0644)
}

// TODO: Should we stat for PurgeCredentials and StoreCredentials
//  and make sure we retain the permissions the file originally had ?
