package cmd

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func newPasswordPrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Please enter your password (it will not be echoed)",
		Mask:  '*',
	}
}

func newUsernamePrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Enter your username to log into Conjur",
	}
}

func mightString(v string, e error) string {
	if e != nil {
		return ""
	}

	return v
}

func mightBool(v bool, e error) bool {
	if e != nil {
		return false
	}

	return v
}

func authenticatedConjurClientForCommand(cmd *cobra.Command) (*conjurapi.Client, error) {
	var err error

	setCommandStreamsOnPrompt := func(prompt *promptui.Prompt) *promptui.Prompt {
		prompt.Stdin = utils.NoopReadCloser(cmd.InOrStdin())
		prompt.Stdout = utils.NoopWriteCloser(cmd.OutOrStdout())

		return prompt
	}

	// TODO: This is called multiple time because each operation potentially uses a new HTTP client bound to the
	// temporary Conjur client being created at that point in time. We should really not be creating so many Conjur clients
	decorateConjurClient := func(conjurClient *conjurapi.Client) {}

	if mightBool(cmd.Flags().GetBool("verbose")) {
		decorateConjurClient = func(conjurClient *conjurapi.Client) {
			if conjurClient == nil {
				return
			}
			// TODO: check to see if the transport has already been decorated to make this idempotent

			httpClient := conjurClient.GetHttpClient()
			transport := httpClient.Transport
			if transport == nil {
				transport = http.DefaultTransport
			}
			httpClient.Transport = utils.NewDumpTransport(
				transport,
				func(dump []byte) {
					cmd.Println(string(dump))
					cmd.Println()
				},
			)
		}
	}

	// TODO: extract this common code for gathering configuring into a seperate package
	// Some of the code is in conjur-api-go and needs to be made configurable so that you can pass a custom path to .conjurrc

	config, err := conjurapi.LoadConfig()
	if err != nil {
		return nil, err
	}

	var authenticator conjurapi.Authenticator
	conjurClient, err := conjurapi.NewClientFromEnvironment(config)
	decorateConjurClient(conjurClient)
	if err == nil {
		authenticator = conjurClient.GetAuthenticator()
	} else {
		cmd.Printf("warn: %s\n", err)
	}

	if authenticator == nil {
		username, password, err := askForCredentials(setCommandStreamsOnPrompt, "", "")
		if err != nil {
			return nil, err
		}

		loginPair := authn.LoginPair{Login: username, APIKey: password}

		conjurClient, err = conjurapi.NewClient(config)
		if err != nil {
			return nil, err
		}

		decorateConjurClient(conjurClient)

		// TODO: maybe have specific struct for login
		data, err := conjurClient.Login(loginPair)
		if err != nil {
			return nil, err
		}

		authenticatePair := authn.LoginPair{Login: username, APIKey: string(data)}
		err = storeCredentials(config, authenticatePair.Login, authenticatePair.APIKey)
		if err != nil {
			return nil, err
		}

		conjurClient, err = conjurapi.NewClientFromKey(config, authenticatePair)
		decorateConjurClient(conjurClient)
		if err != nil {
			return nil, err
		}
	}

	return conjurClient, nil
}

// TODO: whenever this is called we should store to .netrc
func askForCredentials(decoratePrompt decoratePromptFunc, username string, password string) (string, string, error) {
	var err error

	if len(username) == 0 {
		usernamePrompt := decoratePrompt(newUsernamePrompt())
		username, err = runPrompt(usernamePrompt)
		if err != nil {
			return "", "", err
		}
	}

	if len(password) == 0 {
		passwordPrompt := decoratePrompt(newPasswordPrompt())
		password, err = runPrompt(passwordPrompt)
		if err != nil {
			return "", "", err
		}
	}

	return username, password, err
}

func storeCredentials(config conjurapi.Config, login string, apiKey string) error {
	machineName := config.ApplianceURL + "/authn"
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

	// TODO: Should we stat and make sure we retain the permissions the file originally had ?
	if data[len(data)-1] != byte('\n') {
		data = append(data, byte('\n'))
	}

	return ioutil.WriteFile(filePath, data, 0644)
}

func purgeCredentials(config conjurapi.Config) error {
	machineName := config.ApplianceURL + "/authn"
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

	// TODO: Should we stat and make sure we retain the permissions the file originally had ?
	return ioutil.WriteFile(filePath, data, 0644)
}
