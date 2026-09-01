package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

func newVariableCmd(
	getClientFactory variableGetClientFactoryFunc,
	setClientFactory variableSetClientFactoryFunc,
) *cobra.Command {
	variableCmd := &cobra.Command{
		Use:   "variable",
		Short: "Manage Idira Secrets Manager variables",
	}

	variableGetCmd := newVariableGetCmd(getClientFactory)
	variableSetCmd := newVariableSetCmd(setClientFactory)

	variableCmd.AddCommand(variableGetCmd)
	variableCmd.AddCommand(variableSetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	ids := make([]string, 0)
	variableGetCmd.Flags().StringSliceP("id", "i", ids, "(Required) Provide variable identifiers")
	variableGetCmd.MarkFlagRequired("id")

	// Variable versions start from 1 and then increase so we don't know
	// the version of the latest value and can not provide a good default if
	// this flag is an integer. Use a string to provide a default of "".
	variableGetCmd.Flags().StringP("version", "v", "", "Specify the desired version of a single variable value")

	variableSetCmd.Flags().StringP("id", "i", "", "(Required) Provide variable identifier")
	variableSetCmd.MarkFlagRequired("id")
	variableSetCmd.Flags().StringP("value", "v", "", "Set the value of the specified variable")
	variableSetCmd.Flags().StringP("file", "f", "", "Set the value of the specified variable based on file contents. Use \"-\" to read from stdin")
	variableSetCmd.Flags().Bool("trim-whitespace", false, "Trim leading and trailing whitespace from the resolved value")

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
	// Create a new map to store the transformed data
	formattedSecrets := make(map[string]string)

	for fullID, value := range secrets {
		id := strings.Split(string(fullID), ":")
		formattedSecrets[id[len(id)-1]] = string(value)
	}

	if len(formattedSecrets) > 1 {
		return output.Plain(cmd, formattedSecrets, func() error {
			jsonData, err := json.MarshalIndent(formattedSecrets, "", "    ")
			if err != nil {
				return err
			}
			cmd.Println(string(jsonData))
			return nil
		})
	}

	var single string
	for _, v := range formattedSecrets {
		single = v
	}

	return output.Plain(cmd, map[string]any{"value": single}, func() error {
		return output.PlainBare(cmd, single)
	})
}

func newVariableGetCmd(clientFactory variableGetClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get the value of one or more Idira Secrets Manager variables",
		Long: `Get the value of one or more Idira Secrets Manager variables.

Examples:
- conjur variable get -i secret
- conjur variable get -i secret,secret2
- conjur variable get -i secret -v 1
		`,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return promptForMissingRequiredFlags(cmd, "id")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmd.Flags().GetStringSlice("id")
			if err != nil {
				return err
			}
			// Remove Duplicates
			uniqueIDs := make([]string, 0, len(id))
			idSet := make(map[string]struct{})
			for _, variableID := range id {
				if _, exists := idSet[variableID]; !exists {
					idSet[variableID] = struct{}{}
					uniqueIDs = append(uniqueIDs, variableID)
				}
			}

			singleID := ""
			if len(uniqueIDs) == 1 {
				singleID = uniqueIDs[0]
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
				data, err = client.RetrieveBatchSecretsSafe(uniqueIDs)
			} else {
				if len(uniqueIDs) > 1 {
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
		Use:   "set",
		Short: "Set the value of an Idira Secrets Manager variable",
		Long: `Set the value of an Idira Secrets Manager variable.

If neither --value nor --file is given and --interactive is set, you will
be prompted for the value. The prompt input is masked and not displayed.

Examples:
- conjur variable set -i secret -v value
- conjur variable set -i secret -f file.txt
- echo -n "value" | conjur variable set -i secret -f -
		`,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return promptForMissingRequiredFlags(cmd, "id")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			value, err := cmd.Flags().GetString("value")
			if err != nil {
				return err
			}

			file, err := cmd.Flags().GetString("file")
			if err != nil {
				return err
			}

			if value == "" && file == "" {
				interactive, err := interactivePromptingEnabled(cmd)
				if err != nil {
					return err
				}
				if interactive {
					value, err = askForSecretValue(id)
					if err != nil {
						return err
					}
				}
			}

			// Validate that exactly one of --value or --file is provided
			if (value == "" && file == "") || (value != "" && file != "") {
				return fmt.Errorf("must specify exactly one of --value or --file")
			}

			if file != "" {
				if file == "-" {
					valueBytes, err := io.ReadAll(cmd.InOrStdin())
					if err != nil {
						return fmt.Errorf("failed to read value from stdin: %w", err)
					}
					value = string(valueBytes)
				} else {
					valueBytes, err := os.ReadFile(file)
					if err != nil {
						return fmt.Errorf("failed to read file %s: %w", file, err)
					}
					value = string(valueBytes)
				}
			}

			trimWhitespace, err := cmd.Flags().GetBool("trim-whitespace")
			if err != nil {
				return err
			}
			if trimWhitespace {
				value = strings.TrimSpace(value)
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			err = client.AddSecret(id, value)
			if err != nil {
				return err
			}

			return output.Plain(cmd, map[string]any{"id": id, "status": "set"}, func() error {
				cmd.Println("Value added")
				return nil
			})
		},
	}
}

func init() {
	variableCmd := newVariableCmd(variableGetClientFactory, variableSetClientFactory)
	rootCmd.AddCommand(variableCmd)
}
