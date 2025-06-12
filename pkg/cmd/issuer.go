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
	DeleteIssuer(issuerID string, keepSecrets bool) (err error)
	Issuer(issuerID string) (issuer api.Issuer, err error)
	Issuers() (issuers []api.Issuer, err error)
	UpdateIssuer(issuerID string, issuerUpdate api.IssuerUpdate) (updated api.Issuer, err error)
}

type issuerClientFactoryFunc func(*cobra.Command) (issuerClient, error)

func issuerClientFactory(cmd *cobra.Command) (issuerClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

var getBoolFlagFunc = getBoolFlag
var getIntFlagFunc = getIntFlag
var getStringFlagFunc = getStringFlag

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
			id, err := getStringFlagFunc(cmd, "id")
			if err != nil {
				return err
			}

			issuerType, err := getStringFlagFunc(cmd, "type")
			if err != nil {
				return err
			}

			maxTTL, err := getIntFlagFunc(cmd, "max-ttl")
			if err != nil {
				return err
			}

			dataJSON, err := getStringFlagFunc(cmd, "data")
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

func newDeleteIssuerCmd(clientFactory issuerClientFactoryFunc) *cobra.Command {
	deleteIssuerCommand := &cobra.Command{
		Use:   "delete",
		Short: "Delete an issuer",
		Long: `Delete an issuer
Examples:
 - conjur issuer delete --id my-issuer`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getStringFlagFunc(cmd, "id")
			if err != nil {
				return err
			}

			keepSecrets, err := getBoolFlagFunc(cmd, "keep-secrets")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			err = client.DeleteIssuer(id, keepSecrets)
			if err != nil {
				return err
			}

			cmd.Printf("The issuer '%s' was deleted\n", id)

			return nil
		},
	}

	deleteIssuerCommand.Flags().StringP(
		"id",
		"i",
		"",
		"Provide the issuer's identifier (ID)",
	)
	deleteIssuerCommand.MarkFlagRequired("id")

	deleteIssuerCommand.Flags().BoolP(
		"keep-secrets",
		"k",
		false,
		"If set, dynamic secrets associated with this issuer aren't deleted when "+
			"the issuer is deleted",
	)

	return deleteIssuerCommand
}

func newGetIssuerCmd(clientFactory issuerClientFactoryFunc) *cobra.Command {
	getIssuerCommand := &cobra.Command{
		Use:   "get",
		Short: "Retrieve information about an issuer",
		Long: `Retrieve information about an issuer
Examples:
 - conjur issuer get --id my-issuer`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getStringFlagFunc(cmd, "id")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.Issuer(id)
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

	getIssuerCommand.Flags().StringP(
		"id",
		"i",
		"",
		"Provide the issuer's identifier (ID)",
	)
	getIssuerCommand.MarkFlagRequired("id")

	return getIssuerCommand
}

func newListIssuersCmd(clientFactory issuerClientFactoryFunc) *cobra.Command {
	listIssuersCommand := &cobra.Command{
		Use:   "list",
		Short: "List all the issuers",
		Long: `List all the issuers
Examples:
 - conjur issuer list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.Issuers()
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

	return listIssuersCommand
}

func newUpdateIssuerCmd(clientFactory issuerClientFactoryFunc) *cobra.Command {
	updateIssuerCommand := &cobra.Command{
		Use:   "update",
		Short: "Update an issuer",
		Long: `Update an issuer.
You can update the issuer's maximum validity (max-ttl) and/or the data argument's JSON object.
Examples:
 - conjur issuer update --id my-issuer --data '{"access_key_id": "new_key_id", "secret_access_key": "new_key_secret"}'
 - conjur issuer update --id my-issuer --max-ttl 5000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getStringFlagFunc(cmd, "id")
			if err != nil {
				return err
			}

			issuerUpdate := api.IssuerUpdate{}

			maxTTL, err := getIntFlagFunc(cmd, "max-ttl")
			if err != nil {
				return err
			}

			if maxTTL > 0 {
				issuerUpdate.MaxTTL = &maxTTL
			}

			dataJSON, err := getStringFlagFunc(cmd, "data")
			if err != nil {
				return err
			}

			if dataJSON != "" {
				data := make(map[string]interface{})
				err = json.Unmarshal([]byte(dataJSON), &data)
				if err != nil {
					return fmt.Errorf("Failed to parse 'data' flag JSON: %w", err)
				}

				issuerUpdate.Data = data
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.UpdateIssuer(id, issuerUpdate)
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

	updateIssuerCommand.Flags().StringP(
		"id",
		"i",
		"",
		"Provide the issuer's identifier (ID)",
	)
	updateIssuerCommand.MarkFlagRequired("id")

	updateIssuerCommand.Flags().IntP(
		"max-ttl",
		"m",
		0,
		"Provide the maximum validity (time-to-live) for the issuer. The new "+
			"value must be greater than the current value. Maximum: 5000",
	)

	updateIssuerCommand.Flags().String(
		"data",
		"",
		"Update the JSON object",
	)

	return updateIssuerCommand
}

func getBoolFlag(cmd *cobra.Command, key string) (bool, error) {
	return cmd.Flags().GetBool(key)
}

func getIntFlag(cmd *cobra.Command, key string) (int, error) {
	return cmd.Flags().GetInt(key)
}

func getStringFlag(cmd *cobra.Command, key string) (string, error) {
	return cmd.Flags().GetString(key)
}

func init() {
	rootCmd.AddCommand(issuerCmd)

	createIssuerCmd := newCreateIssuerCmd(issuerClientFactory)
	updateIssuerCmd := newUpdateIssuerCmd(issuerClientFactory)
	deleteIssuerCmd := newDeleteIssuerCmd(issuerClientFactory)
	getIssuerCmd := newGetIssuerCmd(issuerClientFactory)
	listIssuersCmd := newListIssuersCmd(issuerClientFactory)

	issuerCmd.AddCommand(createIssuerCmd)
	issuerCmd.AddCommand(updateIssuerCmd)
	issuerCmd.AddCommand(deleteIssuerCmd)
	issuerCmd.AddCommand(getIssuerCmd)
	issuerCmd.AddCommand(listIssuersCmd)
}
