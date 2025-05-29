package cmd

import (
	"encoding/json"
	"fmt"

	api "github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type issuerClient interface {
	CreateIssuer(issuer api.Issuer) (created api.Issuer, err error)
}

type issuerClientFactoryFunc func(*cobra.Command) (issuerClient, error)

func issuerClientFactory(cmd *cobra.Command) (issuerClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

var issuerCmd = &cobra.Command{
	Use:   "issuer",
	Short: "Manage dynamic secret issuers in Conjur",
	Long: `Manage dynamic secret issuers in Conjur

Examples:
 - conjur issuer create --id my-issuer [options] [args] Creates an issuer whose ID is my-issuer
 - conjur issuer delete --id my-issuer                  Deletes an issuer whose ID is my-issuer
 - conjur issuer get --id my-issuer                     Retrieves an issuer whose ID is my-issuer
 - conjur issuer list                                   Lists all the issuers`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

func newCreateIssuerCmd(clientFactory issuerClientFactoryFunc) *cobra.Command {
	createIssuerCommand := &cobra.Command{
		Use:   "create",
		Short: "Create an issuer",
		Long: `Create an issuer

Examples:
 - conjur issuer create --id my-issuer --max-ttl 3000 --type aws --data '{"access_key_id": "my_key_id", "secret_access_key": "my_key_secret"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			issuerType, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}

			maxTTL, err := cmd.Flags().GetInt("max-ttl")
			if err != nil {
				return err
			}

			dataJSON, err := cmd.Flags().GetString("data")
			if err != nil {
				return err
			}

			data := make(map[string]interface{})
			err = json.Unmarshal([]byte(dataJSON), &data)
			if err != nil {
				return fmt.Errorf("Failed to parse 'data' flag JSON: %w", err)
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.CreateIssuer(api.Issuer{
				ID:     id,
				Type:   issuerType,
				MaxTTL: maxTTL,
				Data:   data,
			})
			if err != nil {
				return err
			}

			prettyResult, err := utils.PrettyPrintToJSON(result)
			if err != nil {
				return err
			}

			cmd.Println(prettyResult)

			return nil
		},
	}

	createIssuerCommand.Flags().StringP(
		"id",
		"i",
		"",
		"Provide an identifier (ID) for the issuer. "+
			"Allowed characters: A-Z, a-z, 0-9, + (plus), - (minus), _ (underscore)",
	)
	createIssuerCommand.MarkFlagRequired("id")

	createIssuerCommand.Flags().StringP(
		"type",
		"t",
		"",
		"The type of issuer - dynamic secrets of this type are created under "+
			"this issuer",
	)
	createIssuerCommand.MarkFlagRequired("type")

	createIssuerCommand.Flags().IntP(
		"max-ttl",
		"m",
		0,
		"Provide the maximum validity (max allowed time-to-live) of secrets "+
			"created by this issuer, in seconds. Minimum: 900; Maximum: 5000",
	)
	createIssuerCommand.MarkFlagRequired("max-ttl")

	createIssuerCommand.Flags().String(
		"data",
		"",
		"Provide a JSON object with key-value pairs. "+
			"For example {\"<parameter>\": \"<value>\"}",
	)
	createIssuerCommand.MarkFlagRequired("data")

	return createIssuerCommand
}

func init() {
	rootCmd.AddCommand(issuerCmd)

	createIssuerCmd := newCreateIssuerCmd(issuerClientFactory)

	issuerCmd.AddCommand(createIssuerCmd)
}
