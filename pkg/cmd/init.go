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
	"path/filepath"
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

	return os.WriteFile(filePath, fileContents, 0600)
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
		Short:              "Initialize the Secrets Manager CLI with a Secrets Manager server",
	}
	var env conjurapi.EnvironmentType
	cmd.Flags().Var(&env, "env", "Type of Secrets Manager server to connect to (self-hosted, oss or saas)")
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

	cert, err := utils.GetServerCert(url.Host)
	if err != nil {
		errStr := fmt.Sprintf("Unable to retrieve and validate certificate from %s: %s", url.Host, err)
		return errors.New(errStr)
	}

	var persistCert bool
	// Prompt user to trust the certificate if it is self-signed and --self-signed flag is not set
	if cert.SelfSigned || cert.UntrustedCA {
		persistCert = true
		if !selfSigned {
			err = prompts.AskToTrustCert(cert)
			if err != nil {
				return fmt.Errorf("Based on your selection, this certificate will not be trusted")
			}
		}
	}

	// Initialize combined certificate content
	combinedCert := cert.Cert
	// Check if the host contains ".secretsmgr" which indicates CC and fetch certificate from the identity host
	if strings.Contains(url.Host, ".secretsmgr") {
		combinedCert, err = appendConjurCloudCert(url, selfSigned, combinedCert)
		if err != nil {
			return err
		}
	}

	if persistCert {
		err = persistCertInConfig(config, certFilePath, combinedCert, forceFileOverwrite)
		if err != nil {
			return err
		}
	}

	return nil
}

func persistCertInConfig(config *conjurapi.Config, certFilePath string, combinedCert string, forceFileOverwrite bool) error {
	err := writeFile(certFilePath, []byte(combinedCert), forceFileOverwrite)
	if err != nil {
		return err
	}

	config.SSLCert = combinedCert
	config.SSLCertPath = certFilePath
	return nil
}

func appendConjurCloudCert(url *url.URL, selfSigned bool, combinedCert string) (string, error) {
	identityHost := strings.Replace(url.Host, ".secretsmgr", "", 1)
	identityCert, err := utils.GetServerCert(identityHost)
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve and validate certificate from %s: %s", identityHost, err)
	}

	// Prompt user to trust the identity certificate if it is self-signed and --self-signed flag is not set
	if (identityCert.SelfSigned || identityCert.UntrustedCA) && !selfSigned {
		err = prompts.AskToTrustCert(identityCert)
		if err != nil {
			return "", fmt.Errorf("Based on your selection, this certificate will not be trusted from %s", identityHost)
		}
	}

	// Append the identity certificate to the combined content
	combinedCert += "\n" + identityCert.Cert
	return combinedCert, nil
}

func defaultConjurRC(homeDir string) string {
	defaultConjurRC := filepath.Join(homeDir, ".conjurrc")
	if conjurRC, ok := os.LookupEnv("CONJURRC"); ok {
		defaultConjurRC = conjurRC
	}
	return defaultConjurRC
}

func init() {
	initCmd := newInitCommand()
	initCmd.AddCommand(newCloudInitCmd())
	initCmd.AddCommand(newInitEnterpriseCommand(defaultInitCmdFuncs))
	rootCmd.AddCommand(initCmd)
}
