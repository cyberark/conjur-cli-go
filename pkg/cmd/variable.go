package cmd

import (
	"fmt"
	"strconv"
	"strings"

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
	ids := make([]string, 0)
	variableGetCmd.Flags().StringSliceP("id", "i", ids, "Provide variable identifiers")
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
	RetrieveBatchSecretsSafe(variableIDs []string) (map[string][]byte, error)
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

func printMultilineResults(cmd *cobra.Command, secrets map[string][]byte) error {
	if len(secrets) > 1 {
		cmd.Println("{")
		for fullID, value := range secrets {
			id := strings.Split(string(fullID), ":")
			cmd.Printf("    \"%s\": \"%s\"\n", id[len(id)-1], value)
		}
		cmd.Println("}")
	} else {
		for _, v := range secrets {
			cmd.Println(string(v))
		}
	}
	return nil
}

func newVariableGetCmd(clientFactory variableGetClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Use the get subcommand to get the value of one or more Conjur variables",
		Long: `Use the get subcommand to get the value of one or more Conjur variables

Examples:
- conjur variable get -i secret
- conjur variable get -i secret,secret2
- conjur variable get -i secret -v 1
		`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmd.Flags().GetStringSlice("id")
			if err != nil {
				return err
			}
			singleID := ""
			if len(id) == 1 {
				singleID = id[0]
			}
			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			versionStr, err := cmd.Flags().GetString("version")
			if err != nil {
				return err
			}

			data := map[string][]byte{}

			if versionStr == "" {
				data, err = client.RetrieveBatchSecretsSafe(id)
			} else {
				if len(id) > 1 {
					return fmt.Errorf("version can not be used with multiple variables")
				}
				var version int
				version, err = strconv.Atoi(versionStr)
				if err != nil {
					return err
				}
				data[singleID], err = client.RetrieveSecretWithVersion(singleID, version)
			}
			if err != nil {
				return err
			}
			return printMultilineResults(cmd, data)
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
