package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"

	"github.com/spf13/cobra"
)

type initEnterpriseCmdFlagValues struct {
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

func getInitEnterpriseCmdFlagValues(cmd *cobra.Command) (initEnterpriseCmdFlagValues, error) {
	account, err := cmd.Flags().GetString("account")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	applianceURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	authnType, err := cmd.Flags().GetString("authn-type")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	serviceID, err := cmd.Flags().GetString("service-id")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	conjurrcFilePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	certFilePath, err := cmd.Flags().GetString("cert-file")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	caCert, err := cmd.Flags().GetString("ca-cert")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	jwtFilePath, err := cmd.Flags().GetString("jwt-file")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	jwtHostID, err := cmd.Flags().GetString("jwt-host-id")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	selfSigned, err := cmd.Flags().GetBool("self-signed")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	insecure, err := cmd.Flags().GetBool("insecure")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	forceFileOverwrite, err := cmd.Flags().GetBool("force")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}
	forceNetrc, err := cmd.Flags().GetBool("force-netrc")
	if err != nil {
		return initEnterpriseCmdFlagValues{}, err
	}

	return initEnterpriseCmdFlagValues{
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

func validateCmdFlags(cmdFlagVals initEnterpriseCmdFlagValues, cmd *cobra.Command) error {
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

func runInitEnterpriseCommand(cmd *cobra.Command, funcs initCmdFuncs) error {
	var err error

	cmdFlagVals, err := getInitEnterpriseCmdFlagValues(cmd)
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
		Environment:  env(cmd.CalledAs()),
	}

	// If using JWT auth, we need to ensure that the JWT file exists and
	// contains a valid JWT. To do this, we'll attempt to authenticate.
	if config.AuthnType == "jwt" {
		client, err := conjurapi.NewClientFromJwt(config)
		if err != nil {
			return err
		}
		err = funcs.JWTAuthenticate(client)
		if err != nil {
			return fmt.Errorf("Unable to authenticate with Secrets Manager using the provided JWT file: %s", err)
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

	err = fetchCertIfNeeded(&config, cmdFlagVals.insecure, cmdFlagVals.selfSigned, cmdFlagVals.forceFileOverwrite, cmdFlagVals.certFilePath)
	if err != nil {
		return err
	}
	if config.SSLCertPath != "" {
		cmd.Printf("Wrote certificate to %s\n", config.SSLCertPath)
	}

	err = writeConjurrc(
		config,
		cmdFlagVals.conjurrcFilePath,
		cmdFlagVals.forceFileOverwrite,
	)
	if err != nil {
		return err
	}

	cmd.Printf("Wrote configuration to %s\n", cmdFlagVals.conjurrcFilePath)
	return nil
}

func env(name string) conjurapi.EnvironmentType {
	switch name {
	case string(conjurapi.EnvironmentSH), "CE", "enterprise":
		return conjurapi.EnvironmentSH
	case "oss", "open-source", "OSS":
		return conjurapi.EnvironmentOSS
	default:
		return conjurapi.EnvironmentSH
	}
}

func newInitEnterpriseCommand(funcs initCmdFuncs) *cobra.Command {
	cmd := &cobra.Command{
		Use:     string(conjurapi.EnvironmentSH),
		Aliases: []string{"CE", "open-source", "oss", "OSS", "enterprise"},
		Short:   "Initialize the Secrets Manager CLI with a Secrets Manager server",
		Long: `Initialize the Secrets Manager CLI with a Secrets Manager server.

The init command creates a configuration file (.conjurrc) that contains the details for connecting to Secrets Manager. This file is located under the user's root directory.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitEnterpriseCommand(cmd, funcs)
		},
	}

	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd.Flags().StringP("account", "a", "", "Secrets Manager organization account name")
	cmd.Flags().StringP("url", "u", "", "URL of the Secrets Manager service. Will prompt if omitted.")
	cmd.Flags().StringP("ca-cert", "c", "", "Secrets Manager SSL certificate (will be obtained from host unless provided by this option)")
	cmd.Flags().StringP("file", "f", defaultConjurRC(userHomeDir), "File to write the configuration to. You must set the CONJURRC environment variable to the same value for this file to be used for further commands.")
	cmd.Flags().String("cert-file", filepath.Join(userHomeDir, "conjur-server.pem"), "File to write the server's certificate to")
	cmd.Flags().StringP("authn-type", "t", "", "Authentication type to use (e.g. LDAP, OIDC, JWT)")
	cmd.Flags().String("service-id", "", "Service ID if using alternative authentication type")
	cmd.Flags().String("jwt-file", "", "Path to the JWT file if using authn-jwt")
	cmd.Flags().String("jwt-host-id", "", "Host ID for authn-jwt (not required if JWT contains host ID)")
	cmd.Flags().BoolP("self-signed", "s", false, "Allow self-signed certificates (insecure)")
	cmd.Flags().BoolP("insecure", "i", false, "Allow non-HTTPS connections (insecure)")
	cmd.Flags().Bool("force-netrc", false, "Use a file-based credential storage rather than OS-native keystore (for compatibility with Summon)")
	cmd.Flags().Bool("force", false, "Force overwrite of existing configuration file")

	return cmd
}
