package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/spf13/cobra"
)

type initCmdFuncs struct {
	JWTAuthenticate func(conjurClient clients.ConjurClient) error
}

var defaultInitCmdFuncs = initCmdFuncs{
	JWTAuthenticate: clients.JWTAuthenticate,
}

type initCmdFlagValues struct {
	account            string
	applianceURL       string
	authnType          string
	serviceID          string
	conjurrcFilePath   string
	certFilePath       string
	caCert             string
	jwtFilePath        string
	jwtHostID          string
	forceFileOverwrite bool
	insecure           bool
	selfSigned         bool
	forceNetrc         bool
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
	certFilePath, err := cmd.Flags().GetString("cert-file")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	caCert, err := cmd.Flags().GetString("ca-cert")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	jwtFilePath, err := cmd.Flags().GetString("jwt-file")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	jwtHostID, err := cmd.Flags().GetString("jwt-host-id")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	selfSigned, err := cmd.Flags().GetBool("self-signed")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	insecure, err := cmd.Flags().GetBool("insecure")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	forceFileOverwrite, err := cmd.Flags().GetBool("force")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	forceNetrc, err := cmd.Flags().GetBool("force-netrc")
	if err != nil {
		return initCmdFlagValues{}, err
	}

	return initCmdFlagValues{
		account:            account,
		applianceURL:       applianceURL,
		authnType:          authnType,
		serviceID:          serviceID,
		conjurrcFilePath:   conjurrcFilePath,
		certFilePath:       certFilePath,
		caCert:             caCert,
		jwtFilePath:        jwtFilePath,
		jwtHostID:          jwtHostID,
		selfSigned:         selfSigned,
		insecure:           insecure,
		forceFileOverwrite: forceFileOverwrite,
		forceNetrc:         forceNetrc,
	}, nil
}

func validateCmdFlags(cmdFlagVals initCmdFlagValues, cmd *cobra.Command) error {
	if cmdFlagVals.insecure && cmdFlagVals.selfSigned {
		return fmt.Errorf("Cannot specify both --insecure and --self-signed")
	}
	if (cmdFlagVals.insecure || cmdFlagVals.selfSigned) && cmdFlagVals.caCert != "" {
		return fmt.Errorf("Cannot specify --ca-cert when using --insecure or --self-signed")
	}

	if cmdFlagVals.selfSigned {
		cmd.PrintErrln("Warning: Using self-signed certificates is not recommended and could lead to exposure of sensitive data")
	}
	if cmdFlagVals.insecure {
		cmd.PrintErrln("Warning: Running the command with '--insecure' makes your system vulnerable to security attacks")
		cmd.PrintErrln("If you prefer to communicate with the server securely you must reinitialize the client in secure mode.")
	}
	return nil
}

func runInitCommand(cmd *cobra.Command, funcs initCmdFuncs) error {
	var err error

	cmdFlagVals, err := getInitCmdFlagValues(cmd)
	if err != nil {
		return err
	}

	err = validateCmdFlags(cmdFlagVals, cmd)
	if err != nil {
		return err
	}

	account, applianceURL, err := prompts.MaybeAskForConnectionDetails(
		cmdFlagVals.account,
		cmdFlagVals.applianceURL,
		cmd,
	)
	if err != nil {
		return err
	}

	config := conjurapi.Config{
		Account:      account,
		ApplianceURL: applianceURL,
		AuthnType:    cmdFlagVals.authnType,
		ServiceID:    cmdFlagVals.serviceID,
		JWTFilePath:  cmdFlagVals.jwtFilePath,
		JWTHostID:    cmdFlagVals.jwtHostID,
	}

	// If using JWT auth, we need to ensure that the JWT file exists and
	// contains a valid JWT. To do this, we'll attempt to authenticate.
	if config.AuthnType == "jwt" {
		client, err := conjurapi.NewClient(config)
		if err != nil {
			return err
		}
		err = funcs.JWTAuthenticate(client)
		if err != nil {
			return fmt.Errorf("Unable to authenticate with Conjur using the provided JWT file: %s", err)
		}
	}

	// If the user has specified the --force-netrc flag, don't try to use the native keychain
	if cmdFlagVals.forceNetrc {
		config.CredentialStorage = conjurapi.CredentialStorageFile
	}

	// If user has specified a cert file, read it and set it on the config
	if cmdFlagVals.caCert != "" {
		path, err := filepath.Abs(cmdFlagVals.caCert)
		if err != nil {
			return err
		}
		config.SSLCertPath = path
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	err = fetchCertIfNeeded(&config, cmdFlagVals)
	if err != nil {
		return err
	}
	if config.SSLCertPath != "" {
		cmd.Printf("Wrote certificate to %s\n", config.SSLCertPath)
	}

	err = writeConjurrc(
		config,
		cmdFlagVals,
	)
	if err != nil {
		return err
	}

	cmd.Printf("Wrote configuration to %s\n", cmdFlagVals.conjurrcFilePath)
	return nil
}

func fetchCertIfNeeded(config *conjurapi.Config, cmdFlagVals initCmdFlagValues) error {
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
		if cmdFlagVals.insecure {
			return nil
		}
		return fmt.Errorf("Cannot fetch certificate from non-HTTPS URL %s", url)
	}

	cert, err := utils.GetServerCert(url.Host, cmdFlagVals.selfSigned)
	if err != nil {
		errStr := fmt.Sprintf("Unable to retrieve and validate certificate from %s: %s", url.Host, err)
		if !cmdFlagVals.selfSigned {
			errStr += "\nIf you're attempting to use a self-signed certificate, re-run the init command with the `--self-signed` flag\n"
		}
		return errors.New(errStr)
	}

	// Prompt user to accept certificate
	err = prompts.AskToTrustCert(cert.Fingerprint)
	if err != nil {
		return fmt.Errorf("You decided not to trust the certificate")
	}

	certPath := cmdFlagVals.certFilePath

	err = writeFile(certPath, []byte(cert.Cert), cmdFlagVals.forceFileOverwrite)
	if err != nil {
		return err
	}

	config.SSLCert = cert.Cert
	config.SSLCertPath = certPath

	return nil
}

func writeConjurrc(config conjurapi.Config, cmdFlagVals initCmdFlagValues) error {
	filePath := cmdFlagVals.conjurrcFilePath
	fileContents := config.Conjurrc()

	return writeFile(filePath, fileContents, cmdFlagVals.forceFileOverwrite)
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

func newInitCommand(funcs initCmdFuncs) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the Conjur CLI with a Conjur server",
		Long: `Initialize the Conjur CLI with a Conjur server.

The init command creates a configuration file (.conjurrc) that contains the details for connecting to Conjur. This file is located under the user's root directory.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitCommand(cmd, funcs)
		},
	}

	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd.Flags().StringP("account", "a", "", "Conjur organization account name")
	cmd.Flags().StringP("url", "u", "", "URL of the Conjur service")
	cmd.Flags().StringP("ca-cert", "c", "", "Conjur SSL certificate (will be obtained from host unless provided by this option)")
	cmd.Flags().StringP("file", "f", filepath.Join(userHomeDir, ".conjurrc"), "File to write the configuration to. You must set the CONJURRC environment variable to the same value for this file to be used for further commands.")
	cmd.Flags().String("cert-file", filepath.Join(userHomeDir, "conjur-server.pem"), "File to write the server's certificate to")
	cmd.Flags().StringP("authn-type", "t", "", "Authentication type to use")
	cmd.Flags().String("service-id", "", "Service ID if using alternative authentication type")
	cmd.Flags().String("jwt-file", "", "Path to the JWT file if using authn-jwt")
	cmd.Flags().String("jwt-host-id", "", "Host ID for authn-jwt (not required if JWT contains host ID)")
	cmd.Flags().BoolP("self-signed", "s", false, "Allow self-signed certificates (insecure)")
	cmd.Flags().BoolP("insecure", "i", false, "Allow non-HTTPS connections (insecure)")
	cmd.Flags().Bool("force-netrc", false, "Use a file-based credential storage rather than OS-native keystore (for compatibility with Summon)")
	cmd.Flags().Bool("force", false, "Force overwrite of existing configuration file")

	return cmd
}

func init() {
	initCmd := newInitCommand(defaultInitCmdFuncs)
	rootCmd.AddCommand(initCmd)
}
