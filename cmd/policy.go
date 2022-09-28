package cmd

import (
	"encoding/json"
	"io"
	"os"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

func loadPolicyForCommand(policyMode conjurapi.PolicyMode, cmd *cobra.Command, args []string) error {
	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return err
	}

	filepath, err := cmd.Flags().GetString("filepath")
	if err != nil {
		return err
	}

	var inputReader io.Reader = cmd.InOrStdin()
	// the argument received looks like a file, we try to open it
	if filepath != "-" {
		file, err := os.Open(filepath)
		if err != nil {
			return err
		}
		inputReader = file
	}

	conjurClient, err := authenticatedConjurClientForCommand(cmd)
	if err != nil {
		return err
	}

	policyResponse, err := conjurClient.LoadPolicy(
		policyMode,
		branch,
		inputReader,
	)
	if err != nil {
		return err
	}

	data, err := json.Marshal(policyResponse)
	if err != nil {
		return err
	}

	if prettyData, err := utils.PrettyPrintJSON(data); err == nil {
		data = prettyData
	}

	cmd.PrintErrf("Loaded policy '%s'\n", branch)
	cmd.Println(string(data))

	return nil
}

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Commands for managing policy",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

var policyLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return loadPolicyForCommand(conjurapi.PolicyModePost, cmd, args)
	},
}

var policyAppendCmd = &cobra.Command{
	Use:   "append",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return loadPolicyForCommand(conjurapi.PolicyModePatch, cmd, args)
	},
}

var policyReplaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return loadPolicyForCommand(conjurapi.PolicyModePut, cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(policyCmd)

	policyCmd.PersistentFlags().StringP("branch", "b", "", "The parent policy branch")
	policyCmd.PersistentFlags().StringP("filepath", "f", "", "The policy file to load")
	policyCmd.MarkPersistentFlagRequired("branch")
	policyCmd.MarkPersistentFlagRequired("filepath")

	policyCmd.AddCommand(policyLoadCmd)
	policyCmd.AddCommand(policyAppendCmd)
	policyCmd.AddCommand(policyReplaceCmd)
}
