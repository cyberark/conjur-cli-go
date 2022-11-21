package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/conjurrc"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"

	"github.com/spf13/cobra"
)

type initCmdFlagValues struct {
	account            string
	applianceURL       string
	authnType          string
	serviceID          string
	conjurrcFilePath   string
	forceFileOverwrite bool
}

func getInitCmdFlagValues(cmd *cobra.Command) (initCmdFlagValues, error) {
	account, err := cmd.Flags().GetString("account")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	applianceURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	authnType, err := cmd.Flags().GetString("authn-type")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	serviceID, err := cmd.Flags().GetString("service-id")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	conjurrcFilePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	forceFileOverwrite, err := cmd.Flags().GetBool("force")
	if err != nil {
		return initCmdFlagValues{}, err
	}

	return initCmdFlagValues{
		account:            account,
		applianceURL:       applianceURL,
		authnType:          authnType,
		serviceID:          serviceID,
		conjurrcFilePath:   conjurrcFilePath,
		forceFileOverwrite: forceFileOverwrite,
	}, nil
}

func runInitCommand(cmd *cobra.Command, args []string) error {
	var err error

	cmdFlagVals, err := getInitCmdFlagValues(cmd)
	if err != nil {
		return err
	}

	setCommandStreamsOnPrompt := prompts.PromptDecoratorForCommand(cmd)

	account, applianceURL, err := prompts.MaybeAskForConnectionDetails(
		setCommandStreamsOnPrompt,
		cmdFlagVals.account,
		cmdFlagVals.applianceURL,
	)
	if err != nil {
		return err
	}

	config := conjurapi.Config{
		Account:      account,
		ApplianceURL: applianceURL,
		AuthnType:    cmdFlagVals.authnType,
		ServiceID:    cmdFlagVals.serviceID,
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	err = writeConjurrc(
		config,
		cmdFlagVals,
		setCommandStreamsOnPrompt,
	)
	if err != nil {
		return err
	}

	cmd.Printf("Wrote configuration to %s\n", cmdFlagVals.conjurrcFilePath)
	return nil
}

func writeConjurrc(config conjurapi.Config, cmdFlagVals initCmdFlagValues, setCommandStreamsOnPrompt prompts.DecoratePromptFunc) error {
	return conjurrc.WriteConjurrc(
		config,
		cmdFlagVals.conjurrcFilePath,
		func(filePath string) error {
			if cmdFlagVals.forceFileOverwrite {
				return nil
			}

			err := prompts.AskToOverwriteFile(setCommandStreamsOnPrompt, filePath)
			if err != nil {
				// TODO: make all the errors lowercase to make Go static check happy, then have something higher up that capitalizes the first letter
				// of errors from commands
				return fmt.Errorf("Not overwriting %s", filePath)
			}

			return nil
		})
}

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Use the init command to initialize the Conjur CLI with a Conjur endpoint.",
		Long: `Use the init command to initialize the Conjur CLI with a Conjur endpoint.

The init command creates a configuration file (.conjurrc) that contains the details for connecting to Conjur. This file is located under the user's root directory.`,
		SilenceUsage: true,
		RunE:         runInitCommand,
	}

	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd.Flags().StringP("account", "a", "", "Conjur organization account name")
	cmd.Flags().StringP("url", "u", "", "URL of the Conjur service")
	cmd.Flags().StringP("certificate", "c", "", "Conjur SSL certificate (will be obtained from host unless provided by this option)")
	cmd.Flags().StringP("file", "f", filepath.Join(userHomeDir, ".conjurrc"), "File to write the configuration to")
	cmd.Flags().StringP("authn-type", "t", "", "Authentication type to use")
	cmd.Flags().StringP("service-id", "", "", "Service ID if using alternative authentication type")
	cmd.Flags().Bool("force", false, "Force overwrite of existing file")

	return cmd
}

func init() {
	initCmd := newInitCommand()
	rootCmd.AddCommand(initCmd)
}
