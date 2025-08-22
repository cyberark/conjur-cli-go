package prompts

import (
	"context"
	"errors"
	"fmt"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"net/url"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/cmd/style"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// AskForPrompt presents a prompt to retrieve a custom message from the user
func AskForPrompt(ctx context.Context, message string, timeout time.Duration) (string, error) {
	var userInput string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				Value(&userInput).
				Validate(huh.ValidateNotEmpty()),
		),
	).
		WithTheme(style.GetTheme()).
		WithAccessible(useAccessibleForm())

	//TODO: accessible mode ignores ctx
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := form.RunWithContext(ctxWithTimeout)
	if errors.Is(ctx.Err(), context.Canceled) {
		// if context was canceled it means that the challenge was satisfied without user input
		// and the best method to cleanup is to make huh thinks user canceled the input
		form.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		return "", nil
	}
	return strings.TrimSpace(userInput), err
}

// AskForMFAMechanism presents a prompt to select MFA mechanism to use
func AskForMFAMechanism(options []string) (string, error) {
	return option(
		options,
		"Select MFA Mechanism",
		"Please select the multi factor authentication mechanism you want to use.",
	)
}

// MaybeAskForCredentials optionally presents a prompt to retrieve missing username and/or password from the user
func MaybeAskForCredentials(username, password string) (string, string, error) {
	var err error
	username, err = MaybeAskForUsername(username)
	if err != nil {
		return "", "", err
	}
	password, err = MaybeAskForPassword(password)
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}

// MaybeAskForUsername optionally presents a prompt to retrieve missing username from the user
func MaybeAskForUsername(username string) (string, error) {
	if len(username) > 0 {
		return username, nil
	}
	return input("Enter your username to log into Secrets Manager:")
}

// MaybeAskForPassword optionally presents a prompt to retrieve missing password from the user
func MaybeAskForPassword(password string) (string, error) {
	if len(password) > 0 {
		return password, nil
	}
	return passwordInput("Please enter your password (it will not be echoed):")
}

// MaybeAskForChangePassword optionally presents a prompt to retrieve missing new password from the user
func MaybeAskForChangePassword(newPassword string) (string, error) {
	if len(newPassword) > 0 {
		return newPassword, nil
	}
	return passwordInput("Please enter a new password (it will not be echoed):")
}

// MaybeAskForConnectionDetails presents a prompt to retrieve missing Conjur account and/or URL from the user
func MaybeAskForConnectionDetails(account, applianceURL string, cmd *cobra.Command) (string, string, error) {
	var err error
	applianceURL, err = MaybeAskForURL(applianceURL, cmd)
	if err != nil {
		return "", "", err
	}
	if len(account) > 0 {
		return account, applianceURL, nil
	}
	account, err = input("Enter your organization account name:")
	return account, applianceURL, err
}

// MaybeAskForURL presents a prompt to retrieve missing Conjur URL from the user
func MaybeAskForURL(applianceURL string, _ *cobra.Command) (string, error) {
	if len(applianceURL) > 0 {
		return applianceURL, nil
	}
	var err error
	applianceURL, err = input("Enter the URL of your Secrets Manager service:")
	if err != nil {
		return "", err
	}
	_, err = url.ParseRequestURI(applianceURL)
	return applianceURL, err
}

// MaybeAskToOverwriteFile checks if a file exists and asks the user if they want to overwrite it.
func MaybeAskToOverwriteFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	return AskToOverwriteFile(filePath)
}

// AskToOverwriteFile presents a prompt to get confirmation from a user to overwrite a file
func AskToOverwriteFile(filePath string) error {
	ok, err := confirm(fmt.Sprintf("File %s exists. Overwrite?", filePath), "")
	if !ok {
		return fmt.Errorf("not overwriting %s", filePath)
	}
	return err
}

// AskToTrustCert presents a prompt to get confirmation from a user to trust a certificate
func AskToTrustCert(cert utils.ServerCert) error {
	var warning string
	if cert.SelfSigned {
		warning += "\nWARNING: This is a self-signed certificate. " +
			"Please verify that this is the correct certificate before trusting it.\n"
	}
	if cert.UntrustedCA {
		warning += "\nWARNING: This certificate is signed by an untrusted CA. " +
			"Please verify that this is the correct certificate before trusting it.\n"
	}

	warning += fmt.Sprintf("\nThe server's certificate Sha256 fingerprint is %s.\n", cert.Fingerprint) +
		"Please verify this certificate on the Secrets Manager using command:\n" +
		"openssl x509 -fingerprint -sha256 -noout -in ~conjur/etc/ssl/conjur.pem\n\n" +
		fmt.Sprintf("Certificate issuer organization: %s\n", strings.Join(cert.Issuer.Organization, ",")) +
		fmt.Sprintf("Certificate issuer common name: %s\n", cert.Issuer.CommonName) +
		fmt.Sprintf("Certificate creation date: %s\n", cert.CreationDate) +
		fmt.Sprintf("Certificate expiration date: %s\n", cert.ExpirationDate)

	ok, err := confirm("Trust this certificate?", warning)
	if !ok {
		return errors.New("you decided not to trust the certificate")
	}
	return err
}

// MaybeAskForEnvironment presents a prompt to retrieve missing environment from the user
func MaybeAskForEnvironment(env string) (string, error) {
	if len(env) > 0 {
		return env, nil
	}
	return option(
		conjurapi.SupportedEnvironments,
		"Select the environment you want to use:",
		"The environment you want to use. The default is self-hosted.",
	)
}

func passwordInput(title string) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("user password cannot be requested in non-interactive mode")
	}
	var password string
	err := huh.NewForm(huh.NewGroup(huh.NewInput().
		EchoMode(huh.EchoModeNone).
		Title(title).
		Value(&password).
		Validate(huh.ValidateNotEmpty()),
	)).
		WithTheme(style.GetTheme()).
		WithAccessible(useAccessibleForm()).
		Run()
	return strings.TrimSpace(password), err
}

func input(title string) (string, error) {
	var userInput string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(title).
			Value(&userInput).
			Validate(huh.ValidateNotEmpty()),
	)).
		WithTheme(style.GetTheme()).
		WithAccessible(useAccessibleForm()).
		Run()
	return strings.TrimSpace(userInput), err
}

// option presents a prompt to select an option from a list
func option(options []string, title string, description string) (string, error) {
	var selected string
	err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title(title).
			Options(huh.NewOptions(options...)...).
			Value(&selected).
			Description(description),
	)).
		WithTheme(style.GetTheme()).
		WithAccessible(useAccessibleForm()).
		Run()
	return selected, err
}

// confirm presents a yes/no prompt
func confirm(message, description string) (bool, error) {
	var userInput bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(message).
			Value(&userInput).
			Description(description),
	)).
		WithTheme(style.GetTheme()).
		WithAccessible(useAccessibleForm()).
		Run()
	return userInput, err
}

func useAccessibleForm() bool {
	return !term.IsTerminal(int(os.Stdout.Fd()))
}
