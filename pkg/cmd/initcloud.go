package cmd

import (
	"fmt"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type initCloudCmdFlagValues struct {
	cloudURL           string
	conjurrcFilePath   string
	certFilePath       string
	caCert             string
	proxy              *url.URL
	forceFileOverwrite bool
	selfSigned         bool
	forceNetrc         bool
}

func getInitCloudCmdFlagValues(cmd *cobra.Command) (initCloudCmdFlagValues, error) {
	cloudURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return initCloudCmdFlagValues{}, err
	}
	conjurrcFilePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return initCloudCmdFlagValues{}, err
	}
	certFilePath, err := cmd.Flags().GetString("cert-file")
	if err != nil {
		return initCloudCmdFlagValues{}, err
	}
	caCert, err := cmd.Flags().GetString("ca-cert")
	if err != nil {
		return initCloudCmdFlagValues{}, err
	}
	forceFileOverwrite, err := cmd.Flags().GetBool("force")
	if err != nil {
		return initCloudCmdFlagValues{}, err
	}
	proxyS, err := cmd.Flags().GetString("proxy")
	if err != nil {
		return initCloudCmdFlagValues{}, err
	}
	proxy, err := url.Parse(proxyS)
	if err != nil {
		return initCloudCmdFlagValues{}, fmt.Errorf("invalid proxy URL: %s", proxyS)
	}
	selfSigned, err := cmd.Flags().GetBool("self-signed")
	if err != nil {
		return initCloudCmdFlagValues{}, err
	}
	forceNetrc, err := cmd.Flags().GetBool("force-netrc")
	if err != nil {
		return initCloudCmdFlagValues{}, err
	}

	return initCloudCmdFlagValues{
		cloudURL:           cloudURL,
		conjurrcFilePath:   conjurrcFilePath,
		certFilePath:       certFilePath,
		caCert:             caCert,
		proxy:              proxy,
		forceFileOverwrite: forceFileOverwrite,
		selfSigned:         selfSigned,
		forceNetrc:         forceNetrc,
	}, nil
}

func validateCloudCmdFlags(cmdFlagVals initCloudCmdFlagValues, cmd *cobra.Command) error {
	if !strings.HasPrefix(cmdFlagVals.cloudURL, "https://") {
		return fmt.Errorf("Conjur Cloud URL must use HTTPS")
	}

	if cmdFlagVals.selfSigned {
		cmd.PrintErrln("Warning: Using self-signed certificates is not recommended and could lead to exposure of sensitive data")
	}

	for _, suf := range conjurapi.ConjurCloudSuffixes {
		if strings.Contains(cmdFlagVals.cloudURL, suf) {
			return nil
		}
	}
	return fmt.Errorf("invalid Conjur Cloud URL: %s. Please provide valid", cmdFlagVals.cloudURL)
}

func runInitCloudCommand(cmd *cobra.Command) error {
	var err error

	cmdFlagVals, err := getInitCloudCmdFlagValues(cmd)
	if err != nil {
		return err
	}

	cmdFlagVals.cloudURL, err = prompts.MaybeAskForURL(
		cmdFlagVals.cloudURL,
		cmd,
	)
	if err != nil {
		return err
	}
	cmdFlagVals.cloudURL = normalizeCloudURL(cmdFlagVals.cloudURL)

	err = validateCloudCmdFlags(cmdFlagVals, cmd)
	if err != nil {
		return err
	}

	config := conjurapi.Config{
		Account:      "conjur",
		ApplianceURL: cmdFlagVals.cloudURL,
		AuthnType:    "cloud",
		ServiceID:    "cyberark",
		Proxy:        cmdFlagVals.proxy.String(),
		Environment:  conjurapi.EnvironmentCC,
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

	err = fetchCertIfNeeded(
		&config,
		false,
		cmdFlagVals.selfSigned,
		cmdFlagVals.forceFileOverwrite,
		cmdFlagVals.certFilePath,
	)
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

func normalizeCloudURL(applianceURL string) string {
	result := strings.TrimSuffix(applianceURL, "/")
	if !strings.HasSuffix(result, "/api") {
		return result + "/api"
	}
	return result
}

func newCloudInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cloud",
		Aliases: []string{"CC"},
		Short:   "Initialize the Conjur CLI with a Conjur Cloud server",
		Long: `Initialize the Conjur CLI with a Conjur Cloud server.

The init command creates a configuration file (.conjurrc) that contains the details for connecting to Conjur. This file is located under the user's root directory.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitCloudCommand(cmd)
		},
	}

	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// cmd.Flags().StringP("account", "a", "", "Conjur organization account name")
	cmd.Flags().StringP("url", "u", "", "URL of the Conjur service. Will prompt if omitted.")
	cmd.Flags().StringP("ca-cert", "c", "", "Conjur SSL certificate (will be obtained from host unless provided by this option)")
	cmd.Flags().StringP("file", "f", filepath.Join(userHomeDir, ".conjurrc"), "File to write the configuration to. You must set the CONJURRC environment variable to the same value for this file to be used for further commands.")
	cmd.Flags().String("cert-file", filepath.Join(userHomeDir, "conjur-server.pem"), "File to write the server's certificate to")
	cmd.Flags().StringP("proxy", "p", "", "Proxy URL to use for connecting to Conjur")
	// cmd.Flags().String("service-id", "", "Service ID if using alternative authentication type")
	// cmd.Flags().String("jwt-file", "", "Path to the JWT file if using authn-jwt")
	// cmd.Flags().String("jwt-host-id", "", "Host ID for authn-jwt (not required if JWT contains host ID)")
	cmd.Flags().BoolP("self-signed", "s", false, "Allow self-signed certificates (insecure)")
	// cmd.Flags().BoolP("insecure", "i", false, "Allow non-HTTPS connections (insecure)")
	cmd.Flags().Bool("force-netrc", false, "Use a file-based credential storage rather than OS-native keystore (for compatibility with Summon)")
	cmd.Flags().Bool("force", false, "Force overwrite of existing configuration file")

	return cmd
}
