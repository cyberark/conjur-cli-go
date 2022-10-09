package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/spf13/cobra"
)

func NewVariableCmd(
	getClientFactory variableGetClientFactoryFunc,
	setClientFactory variableSetClientFactoryFunc,
) *cobra.Command {
	variableCmd := &cobra.Command{
		Use:   "variable",
		Short: "Use the variable command to manage Conjur variables",
	}

	variableGetCmd := NewVariableGetCmd(getClientFactory)
	variableSetCmd := NewVariableSetCmd(setClientFactory)

	variableCmd.AddCommand(variableGetCmd)
	variableCmd.AddCommand(variableSetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	variableGetCmd.Flags().StringP("id", "i", "", "Provide variable identifier")
	variableGetCmd.MarkFlagRequired("id")
	variableGetCmd.Flags().StringP("version", "v", "", "Specify the desired version of a single variable value")

	variableSetCmd.Flags().StringP("id", "i", "", "Provide variable identifier")
	variableSetCmd.MarkFlagRequired("id")
	variableSetCmd.Flags().StringP("value", "v", "", "Set the value of the specified variable")
	variableSetCmd.MarkFlagRequired("value")

	return variableCmd
}

type variableGetClient interface {
	RetrieveSecret(string) ([]byte, error)
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

func NewVariableGetCmd(clientFactory variableGetClientFactoryFunc) *cobra.Command {
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

			// TODO: update RetrieveSecret to accept version
			data, err := client.RetrieveSecret(id)
			if err != nil {
				return err
			}

			cmd.Println(string(data))
			return nil
		},
	}
}

func NewVariableSetCmd(clientFactory variableSetClientFactoryFunc) *cobra.Command {
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
	variableCmd := NewVariableCmd(variableGetClientFactory, variableSetClientFactory)
	rootCmd.AddCommand(variableCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// variableCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
