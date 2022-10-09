package prompts

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/manifoldco/promptui"
)

type decoratePromptFunc func(*promptui.Prompt) *promptui.Prompt

func newApplianceURLPrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Enter the URL of your Conjur service",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("URL is required")
			}

			_, err := url.ParseRequestURI(input)
			return err
		},
	}
}

func newAccountPrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Enter your organization account name",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("Account is required")
			}
			return nil
		},
	}
}

func newFileExistsPrompt(filePath string) *promptui.Prompt {
	return &promptui.Prompt{
		Label:     fmt.Sprintf("File %s exists. Overwrite", filePath),
		IsConfirm: true,
	}
}

func AskToOverwriteFile(decoratePrompt decoratePromptFunc, filePath string) error {
	fileExistsPrompt := decoratePrompt(newFileExistsPrompt(filePath))
	_, err := runPrompt(fileExistsPrompt)
	return err
}

func runPrompt(prompt *promptui.Prompt) (userInput string, err error) {
	userInput, err = prompt.Run()
	if err != nil {
		return "", err
	}
	return userInput, nil
}

func newPasswordPrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Please enter your password (it will not be echoed)",
		Mask:  ' ',
	}
}

func newUsernamePrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Enter your username to log into Conjur",
	}
}

// TODO: whenever this is called we should store to .netrc
func AskForCredentials(decoratePrompt decoratePromptFunc, username string, password string) (string, string, error) {
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

// TODO: whenever this is called we should store to .conjurrc
func AskForConnectionDetails(decoratePrompt decoratePromptFunc, account string, applianceURL string) (string, string, error) {
	var err error

	if len(applianceURL) == 0 {
		prompt := decoratePrompt(newApplianceURLPrompt())
		applianceURL, err = runPrompt(prompt)

		if err != nil {
			return "", "", err
		}
	}

	if len(account) == 0 {
		prompt := decoratePrompt(newAccountPrompt())
		account, err = runPrompt(prompt)

		if err != nil {
			return "", "", err
		}
	}

	return account, applianceURL, err
}
