package prompts

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
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

func newEnvironmentPrompt() *survey.Select {
	return &survey.Select{
		Message: "Select the environment you want to use:",
		Options: conjurapi.SupportedEnvironments,
		Default: string(conjurapi.EnvironmentCE),
		Help:    "The environment you want to use. The default is enterprise.",
		VimMode: true,
	}
}

func stringToBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "y":
		return true
	case "yes":
		return true
	default:
		return false
	}
}

func basicConfirm(message string) bool {
	fmt.Fprintf(os.Stdout, "%s ", message)
	var answer string
	fmt.Scanln(&answer)
	fmt.Fprintln(os.Stdout, "")
	return stringToBool(answer)
}

func confirm(message string) (bool, error) {
	var err error
	var userInput bool

	// When handling Pipe-based Stdin, we can't use Survey to gather responses.
	// https://github.com/go-survey/survey/issues/394
	// Instead, use standard fmt funcs to prompt to Stdout and gather from Stdin.
	if isatty.IsTerminal(os.Stdin.Fd()) {
		q := &survey.Question{
			Prompt:   &survey.Confirm{Message: message},
			Validate: survey.Required,
		}
		err = survey.AskOne(q.Prompt, &userInput, survey.WithValidator(q.Validate), survey.WithShowCursor(true))
	} else {
		userInput = basicConfirm(message)
	}

	return userInput, err
}

// AskToOverwriteFile presents a prompt to get confirmation from a user to overwrite a file
func AskToOverwriteFile(filePath string) error {
	userInput, err := confirm(fmt.Sprintf("File %s exists. Overwrite?", filePath))

	if !userInput {
		return fmt.Errorf("Not overwriting %s", filePath)
	}
	return err
}

// AskToTrustCert presents a prompt to get confirmation from a user to trust a certificate
func AskToTrustCert(fingerprint string) error {
	warning := fmt.Sprintf("\nThe server's certificate Sha256 fingerprint is %s.\n", fingerprint) +
		"Please verify this certificate on the appliance using command:\n" +
		"openssl x509 -fingerprint -sha256 -noout -in ~conjur/etc/ssl/conjur.pem\n\n"
	os.Stdout.Write([]byte(warning))

	userInput, err := confirm("Trust this certificate?")

	if !userInput {
		return errors.New("You decided not to trust the certificate.")
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
func MaybeAskForConnectionDetails(account string, applianceURL string, cmd *cobra.Command) (string, string, error) {
	var err error
	var userInput string

	applianceURL, err = MaybeAskForURL(applianceURL, cmd)
	if err != nil {
		return "", "", err
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

// MaybeAskForURL presents a prompt to retrieve missing Conjur URL from the user
func MaybeAskForURL(url string, _ *cobra.Command) (string, error) {
	var err error
	var userInput string

	if len(url) == 0 {
		q := newApplianceURLPrompt()
		err := survey.AskOne(q.Prompt, &userInput, survey.WithValidator(q.Validate), survey.WithShowCursor(true))
		if err != nil {
			return "", err
		}
		url = userInput
	}

	return url, err
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

// MaybeAskForEnvironment presents a prompt to retrieve missing environment from the user
func MaybeAskForEnvironment(env string) (string, error) {
	var err error
	var userInput string

	if len(env) == 0 {
		q := newEnvironmentPrompt()
		if err := survey.AskOne(q, &userInput, survey.WithShowCursor(true)); err != nil {
			return "", err
		}
		env = userInput
	}
	return env, err
}
