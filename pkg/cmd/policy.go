package cmd

import (
	"encoding/json"
	"io"
	"os"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

func loadPolicyCommandRunner(
	clientFactory policyClientFactoryFunc,
	policyMode conjurapi.PolicyMode,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
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

		conjurClient, err := clientFactory(cmd)
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
}

func newPolicyCommand(clientFactory policyClientFactoryFunc) *cobra.Command {
	policyCmd := &cobra.Command{
		Use:   "policy",
		Short: "Use the policy command to manage Conjur policies",
	}

	policyCmd.PersistentFlags().StringP("branch", "b", "", "The parent policy branch")
	policyCmd.PersistentFlags().StringP("filepath", "f", "", "The policy file to load")
	policyCmd.MarkPersistentFlagRequired("branch")
	policyCmd.MarkPersistentFlagRequired("filepath")

	policyCmd.AddCommand(newPolicyLoadCommand(clientFactory))
	policyCmd.AddCommand(newPolicyAppendCommand(clientFactory))
	policyCmd.AddCommand(newPolicyReplaceCommand(clientFactory))

	return policyCmd
}

func newPolicyLoadCommand(clientFactory policyClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:          "load",
		Short:        "Load a policy and create resources",
		SilenceUsage: true,
		RunE:         loadPolicyCommandRunner(clientFactory, conjurapi.PolicyModePost),
	}
}

func newPolicyAppendCommand(clientFactory policyClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:          "append",
		Short:        "Update existing resources in the policy or create new resources",
		SilenceUsage: true,
		RunE:         loadPolicyCommandRunner(clientFactory, conjurapi.PolicyModePatch),
	}

}

func newPolicyReplaceCommand(clientFactory policyClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:          "replace",
		Short:        "Fully replace an existing policy",
		SilenceUsage: true,
		RunE:         loadPolicyCommandRunner(clientFactory, conjurapi.PolicyModePut),
	}

}

type policyClient interface {
	LoadPolicy(mode conjurapi.PolicyMode, policyBranch string, policySrc io.Reader) (*conjurapi.PolicyResponse, error)
}

type policyClientFactoryFunc func(*cobra.Command) (policyClient, error)

func policyClientFactory(cmd *cobra.Command) (policyClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func init() {
	policyCmd := newPolicyCommand(policyClientFactory)

	rootCmd.AddCommand(policyCmd)
}
