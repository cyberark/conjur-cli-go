package cmd

import (
	"errors"
	"fmt"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"strings"
)

type initCmdFuncs struct {
	JWTAuthenticate func(conjurClient clients.ConjurClient) error
}

var defaultInitCmdFuncs = initCmdFuncs{
	JWTAuthenticate: clients.JWTAuthenticate,
}

func writeConjurrc(config conjurapi.Config, conjurrcFilePath string, forceFileOverwrite bool) error {
	fileContents := config.Conjurrc()

	return writeFile(conjurrcFilePath, fileContents, forceFileOverwrite)
}

func writeFile(filePath string, fileContents []byte, forceFileOverwrite bool) error {
	if !forceFileOverwrite {
		err := prompts.MaybeAskToOverwriteFile(filePath)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(filePath, fileContents, 0644)
}

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			// in order to replace the init RunE function with subcommand RunE it have to exist
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// flags are not parsed yet, any flag from subcommand would fail the parsing on the init command
			env := getEnv(args)
			// if the env flag is not set, check if the user provided it as an argument
			env, err := prompts.MaybeAskForEnvironment(env)
			if err != nil {
				return err
			}
			if len(env) == 0 {
				return fmt.Errorf("missing the env name")
			}
			// find the subcommand
			subCmd, _, err := cmd.Traverse([]string{env})
			if err != nil {
				return err
			}
			// add subcommand flags to the main command
			cmd.Flags().AddFlagSet(subCmd.Flags())
			// in order for parse flags to work enable flag parsing
			cmd.DisableFlagParsing = false
			// parse the flags for the subcommand
			err = cmd.ParseFlags(args)
			if err != nil {
				return err
			}
			// help requires special treatment
			if ok, err := cmd.Flags().GetBool("help"); err == nil && ok {
				_ = subCmd.Help()
				return nil
			}
			// set the subcommand to run as the init command
			cmd.RunE = subCmd.RunE
			return nil
		},
		// we have to disable automatic flags parsing before handling the env flag
		DisableFlagParsing: true,
		Short:              "Initialize the Conjur CLI with a Conjur server",
	}
	var env conjurapi.EnvironmentType
	cmd.Flags().Var(&env, "env", "Type of conjur server to connect to (enterprise, oss or cloud)")
	return cmd
}

func getEnv(args []string) string {
	for i, arg := range args {
		if strings.HasPrefix(arg, "--env=") {
			return strings.TrimPrefix(arg, "--env=")
		}
		if arg == "--env" && len(args) > i+1 {
			return args[i+1]
		}
	}
	return ""
}

func fetchCertIfNeeded(config *conjurapi.Config, insecure, selfSigned, forceFileOverwrite bool, certFilePath string) error {
	// If user has specified a cert file, don't fetch it from the server
	if config.SSLCertPath != "" {
		return nil
	}

	// Get TLS certificate from Conjur server
	url, err := url.Parse(config.ApplianceURL)
	if err != nil {
		return err
	}

	// Only fetch certificate if using HTTPS
	if url.Scheme != "https" {
		if insecure {
			return nil
		}
		return fmt.Errorf("Cannot fetch certificate from non-HTTPS URL %s", url)
	}

	cert, err := utils.GetServerCert(url.Host, selfSigned)
	if err != nil {
		errStr := fmt.Sprintf("Unable to retrieve and validate certificate from %s: %s", url.Host, err)
		if !selfSigned {
			errStr += "\nIf you're attempting to use a self-signed certificate, re-run the init command with the `--self-signed` flag\n"
		}
		return errors.New(errStr)
	}

	// Prompt user to accept certificate
	err = prompts.AskToTrustCert(cert.Fingerprint)
	if err != nil {
		return fmt.Errorf("You decided not to trust the certificate")
	}

	certPath := certFilePath

	err = writeFile(certPath, []byte(cert.Cert), forceFileOverwrite)
	if err != nil {
		return err
	}

	config.SSLCert = cert.Cert
	config.SSLCertPath = certPath

	return nil
}

func init() {
	initCmd := newInitCommand()
	initCmd.AddCommand(newCloudInitCmd())
	initCmd.AddCommand(newInitEnterpriseCommand(defaultInitCmdFuncs))
	rootCmd.AddCommand(initCmd)
}
