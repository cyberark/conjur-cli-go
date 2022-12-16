package cmd

import (
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

	variableGetCmd.Flags().IntP("version", "v", 0, "Specify the desired version of a single variable value")

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
		Use:          "get <variable-id> [-v]",
		Short:        "Retrieve the value of a Conjur variable",
		Long:         `Retrieve the value of a Conjur variables given a <variable-id> and optional [version].

Examples:
- conjur variable get ci-staging-password
- conjur variable get ci-staging-password -v 2`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.PrintErr("Error: Must specify <variable-id>")
				return nil
			}

			variableID := args[0]
			
			version, err := cmd.Flags().GetInt("version")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			var data []byte

			if version != 0 {
				data, err = client.RetrieveSecretWithVersion(variableID, version)
			} else {
				data, err = client.RetrieveSecret(variableID)
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
		Use:          "set <variable-id> <value>",
		Short:        "Set the value of a Conjur variable given a <variable-id> and <value>",
		Long:         `Set the value of a Conjur variable given a <variable-id> and <value>.

Examples:
- conjur variable set ci-staging-password secret-value`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				cmd.PrintErr("Error: Must specify <variable-id> and <value>")
				return nil
			}

			variableID := args[0]
			value := args[1]
			
			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			err = client.AddSecret(variableID, value)
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
