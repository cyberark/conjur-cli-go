package cmd

import (
	"strconv"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/spf13/cobra"
)

func newVariableCmd(
	getClientFactory variableGetClientFactoryFunc,
	setClientFactory variableSetClientFactoryFunc,
) *cobra.Command {
	variableCmd := &cobra.Command{
		Use:   "variable",
		Short: "Use the variable command to manage Conjur variables",
	}

	variableGetCmd := newVariableGetCmd(getClientFactory)
	variableSetCmd := newVariableSetCmd(setClientFactory)

	variableCmd.AddCommand(variableGetCmd)
	variableCmd.AddCommand(variableSetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	variableGetCmd.Flags().StringP("id", "i", "", "Provide variable identifier")
	variableGetCmd.MarkFlagRequired("id")

	// Variable versions start from 1 and then increase so we don't know
	// the version of the latest value and can not provide a good default if
	// this flag is an integer. Use a string to provide a default of "".
	variableGetCmd.Flags().StringP("version", "v", "", "Specify the desired version of a single variable value")

	variableSetCmd.Flags().StringP("id", "i", "", "Provide variable identifier")
	variableSetCmd.MarkFlagRequired("id")
	variableSetCmd.Flags().StringP("value", "v", "", "Set the value of the specified variable")
	variableSetCmd.MarkFlagRequired("value")

	return variableCmd
}

type variableGetClient interface {
	RetrieveSecret(string) ([]byte, error)
	RetrieveSecretWithVersion(string, int) ([]byte, error)
}
type variableSetClient interface {
	AddSecret(string, string) error
}

type variableGetClientFactoryFunc func(*cobra.Command) (variableGetClient, error)
type variableSetClientFactoryFunc func(*cobra.Command) (variableSetClient, error)

func variableGetClientFactory(cmd *cobra.Command) (variableGetClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func variableSetClientFactory(cmd *cobra.Command) (variableSetClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func newVariableGetCmd(clientFactory variableGetClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:          "get",
		Short:        "Use the get subcommand to get the value of one or more Conjur variables",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			versionStr, err := cmd.Flags().GetString("version")
			if err != nil {
				return err
			}

			var data []byte

			if versionStr == "" {
				data, err = client.RetrieveSecret(id)
			} else {
				version, err := strconv.Atoi(versionStr)
				if err != nil {
					return err
				}

				data, err = client.RetrieveSecretWithVersion(id, version)
			}

			if err != nil {
				return err
			}

			cmd.Println(string(data))

			return nil
		},
	}
}

func newVariableSetCmd(clientFactory variableSetClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:          "set",
		Short:        "Use the set subcommand to set the value of a Conjur variable",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			value, err := cmd.Flags().GetString("value")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			err = client.AddSecret(id, value)
			if err != nil {
				return err
			}

			cmd.Println("Value added")
			return nil
		},
	}
}

func init() {
	variableCmd := newVariableCmd(variableGetClientFactory, variableSetClientFactory)
	rootCmd.AddCommand(variableCmd)
}
