package prompts

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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

// AskToOverwriteFile presents a prompt to get confirmation from a user to overwrite a file
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

// PromptDecoratorForCommand redirects the stdin and stdout from the command onto the prompt. This is to
// ensure that command and prompt stay synchronized
func PromptDecoratorForCommand(cmd *cobra.Command) func(*promptui.Prompt) *promptui.Prompt {
	return func(prompt *promptui.Prompt) *promptui.Prompt {
		prompt.Stdin = utils.NoopReadCloser(cmd.InOrStdin())
		prompt.Stdout = utils.NoopWriteCloser(cmd.OutOrStdout())

		return prompt
	}

}

// MaybeAskForCredentials optionally presents a prompt to retrieve missing username and/or password from the user
func MaybeAskForCredentials(decoratePrompt decoratePromptFunc, username string, password string) (string, string, error) {
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

// MaybeAskForConnectionDetails presents a prompt to retrieve missing Conjur account and/or URL from the user
func MaybeAskForConnectionDetails(decoratePrompt decoratePromptFunc, account string, applianceURL string) (string, string, error) {
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
