package prompts

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// DecoratePromptFunc is a function that decorates a prompt with additional functionality.
// For example, it may be used to redirect the stdin and stdout from a command onto a prompt.
type DecoratePromptFunc func(*promptui.Prompt) *promptui.Prompt

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
func AskToOverwriteFile(decoratePrompt DecoratePromptFunc, filePath string) error {
	fileExistsPrompt := decoratePrompt(newFileExistsPrompt(filePath))
	_, err := runPrompt(fileExistsPrompt)
	return err
}

func newTrustCertPrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label:     "Trust this certificate",
		IsConfirm: true,
	}
}

// AskToTrustCert presents a prompt to get confirmation from a user to trust a certificate
func AskToTrustCert(decoratePrompt DecoratePromptFunc, fingerprint string) error {
	warning := fmt.Sprintf("\nThe server's certificate fingerprint is %s.\n", fingerprint) +
		"Please verify this certificate on the appliance using command:\n" +
		"openssl x509 -fingerprint -noout -in ~conjur/etc/ssl/conjur.pem\n\n"

	trustCertPrompt := decoratePrompt(newTrustCertPrompt())
	// The prompt label doesn't support newlines, so we write the warning to stdout directly
	trustCertPrompt.Stdout.Write([]byte(warning))
	_, err := runPrompt(trustCertPrompt)
	return err
}

func runPrompt(prompt *promptui.Prompt) (userInput string, err error) {
	// To avoid loss of contents described in the documentation for NewByteBufferedReader
	// we wrap the stdin reader.
	prompt.Stdin = utils.NoopReadCloser(
		utils.NewByteBufferedReader(prompt.Stdin),
	)

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
func PromptDecoratorForCommand(cmd *cobra.Command) DecoratePromptFunc {
	return func(prompt *promptui.Prompt) *promptui.Prompt {
		prompt.Stdin = utils.NoopReadCloser(cmd.InOrStdin())
		prompt.Stdout = utils.NoopWriteCloser(cmd.OutOrStdout())

		return prompt
	}

}

// MaybeAskForCredentials optionally presents a prompt to retrieve missing username and/or password from the user
func MaybeAskForCredentials(decoratePrompt DecoratePromptFunc, username string, password string) (string, string, error) {
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
func MaybeAskForConnectionDetails(decoratePrompt DecoratePromptFunc, account string, applianceURL string) (string, string, error) {
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

// MaybeAskToOverwriteFile checks if a file exists and asks the user if they want to overwrite it. Returns `nil` if
// the file does not exist or if the user confirms they want to overwrite it.
func MaybeAskToOverwriteFile(decoratePrompt DecoratePromptFunc, filePath string, forceOverwrite bool) error {
	if forceOverwrite {
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	err := AskToOverwriteFile(decoratePrompt, filePath)
	if err != nil {
		// TODO: make all the errors lowercase to make Go static check happy, then have something higher up that capitalizes the first letter
		// of errors from commands
		return fmt.Errorf("Not overwriting %s", filePath)
	}

	return nil
}
