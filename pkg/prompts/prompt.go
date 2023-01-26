package prompts

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/AlecAivazis/survey/v2"
)

func newApplianceURLPrompt() *survey.Question {
	return &survey.Question{
		Prompt: &survey.Input{Message: "Enter the URL of your Conjur service:"},
		Validate: func(val interface{}) error {
			input, _ := val.(string)
			if len(input) == 0 {
				return errors.New("URL is required")
			}

			_, err := url.ParseRequestURI(input)
			return err
		},
	}
}

func newAccountPrompt() *survey.Question {
	return &survey.Question{
		Prompt:   &survey.Input{Message: "Enter your organization account name:"},
		Validate: survey.Required,
	}
}

func newFileExistsPrompt(filePath string) *survey.Question {
	return &survey.Question{
		Prompt:   &survey.Confirm{Message: fmt.Sprintf("File %s exists. Overwrite?", filePath)},
		Validate: survey.Required,
	}
}

// AskToOverwriteFile presents a prompt to get confirmation from a user to overwrite a file
func AskToOverwriteFile(filePath string) error {
	var userInput bool

	q := newFileExistsPrompt(filePath)
	err := survey.AskOne(q.Prompt, &userInput, survey.WithValidator(q.Validate), survey.WithShowCursor(true))

	if userInput == false {
		return errors.New("Not trusting cert.")
	}
	return err
}

func newTrustCertPrompt() *survey.Question {
	return &survey.Question{
		Prompt:   &survey.Confirm{Message: "Trust this certificate?"},
		Validate: survey.Required,
	}
}

// AskToTrustCert presents a prompt to get confirmation from a user to trust a certificate
func AskToTrustCert(fingerprint string) error {
	var userInput bool

	warning := fmt.Sprintf("\nThe server's certificate fingerprint is %s.\n", fingerprint) +
		"Please verify this certificate on the appliance using command:\n" +
		"openssl x509 -fingerprint -noout -in ~conjur/etc/ssl/conjur.pem\n\n"
	os.Stdout.Write([]byte(warning))

	q := newTrustCertPrompt()
	err := survey.AskOne(q.Prompt, &userInput, survey.WithValidator(q.Validate), survey.WithShowCursor(true))

	if userInput == false {
		return errors.New("Not trusting cert.")
	}
	return err
}

func newPasswordPrompt() *survey.Question {
	return &survey.Question{
		Prompt:   &survey.Password{Message: "Please enter your password (it will not be echoed):"},
		Validate: survey.Required,
	}
}

func newUsernamePrompt() *survey.Question {
	return &survey.Question{
		Prompt:   &survey.Input{Message: "Enter your username to log into Conjur:"},
		Validate: survey.Required,
	}
}

func newChangePasswordPrompt() *survey.Question {
	return &survey.Question{
		Prompt:   &survey.Password{Message: "Please enter a new password (it will not be echoed):"},
		Validate: survey.Required,
	}
}

// MaybeAskForCredentials optionally presents a prompt to retrieve missing username and/or password from the user
func MaybeAskForCredentials(username string, password string) (string, string, error) {
	var err error
	var userInput string

	if len(username) == 0 {
		q := newUsernamePrompt()
		err := survey.AskOne(q.Prompt, &userInput, survey.WithValidator(q.Validate), survey.WithShowCursor(true))
		if err != nil {
			return "", "", err
		}
		username = userInput
	}

	if len(password) == 0 {
		q := newPasswordPrompt()
		err := survey.AskOne(q.Prompt, &userInput, survey.WithValidator(q.Validate), survey.WithShowCursor(true))
		if err != nil {
			return "", "", err
		}
		password = userInput
	}

	return username, password, err
}

// MaybeAskForChangePassword optionally presents a prompt to retrieve missing new password from the user
func MaybeAskForChangePassword(newPassword string) (string, error) {
	var err error
	var userInput string

	if len(newPassword) == 0 {
		q := newChangePasswordPrompt()
		err := survey.AskOne(q.Prompt, &userInput, survey.WithValidator(q.Validate), survey.WithShowCursor(true))
		if err != nil {
			return "", err
		}
		newPassword = userInput
	}

	return newPassword, err
}

// MaybeAskForConnectionDetails presents a prompt to retrieve missing Conjur account and/or URL from the user
func MaybeAskForConnectionDetails(account string, applianceURL string) (string, string, error) {
	var err error
	var userInput string

	if len(applianceURL) == 0 {
		q := newApplianceURLPrompt()
		err := survey.AskOne(q.Prompt, &userInput, survey.WithValidator(q.Validate), survey.WithShowCursor(true))
		if err != nil {
			return "", "", err
		}
		applianceURL = userInput
	}

	if len(account) == 0 {
		q := newAccountPrompt()
		err := survey.AskOne(q.Prompt, &userInput, survey.WithValidator(q.Validate), survey.WithShowCursor(true))
		if err != nil {
			return "", "", err
		}
		account = userInput
	}

	return account, applianceURL, err
}

// MaybeAskToOverwriteFile checks if a file exists and asks the user if they want to overwrite it. Returns `nil` if
// the file does not exist or if the user confirms they want to overwrite it.
func MaybeAskToOverwriteFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	err := AskToOverwriteFile(filePath)
	if err != nil {
		// TODO: make all the errors lowercase to make Go static check happy, then have something higher up that capitalizes the first letter
		// of errors from commands
		return fmt.Errorf("Not overwriting %s", filePath)
	}

	return nil
}
