package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

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
		dryrun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		branch, err := cmd.Flags().GetString("branch")
		if err != nil {
			return err
		}

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		var inputReader io.Reader = cmd.InOrStdin()
		// the argument received looks like a file, we try to open it
		if file != "-" {
			file, err := os.Open(file)
			if err != nil {
				return err
			}
			inputReader = file
		}

		conjurClient, err := clientFactory(cmd)
		if err != nil {
			return err
		}

		data, err := DryRunOrLoadPolicy(conjurClient, dryrun, policyMode, branch, inputReader)
		if err != nil {
			return err
		}

		if prettyData, err := utils.PrettyPrintJSON(data); err == nil {
			data = prettyData
		}

		cmd.PrintErrf("%s policy '%s'\n", cmdMessage(dryrun), branch)
		cmd.Println(string(data))

		return nil
	}
}

func fetchPolicyCommandRunner(
	clientFactory policyClientFactoryFunc,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		branch, err := cmd.Flags().GetString("branch")
		if err != nil {
			return err
		}

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}

		// validate output arg
		if output != "yaml" && output != "json" {
			return errors.New("output format must be 'yaml' or 'json'")
		}
		var returnJSON bool
		if output == "json" {
			returnJSON = true
		}

		depth, err := cmd.Flags().GetUint("depth")
		if err != nil {
			return err
		}
		if depth > 64 {
			return errors.New("depth must be less than or equal to 64")
		}

		limit, err := cmd.Flags().GetUint("limit")
		if err != nil {
			return err
		}
		if limit > 100000 {
			return errors.New("limit must be less than or equal to 100000")
		}

		// validate file arg
		if file != "" {
			if err := validateFilePath(file); err != nil {
				return err
			}
		}

		conjurClient, err := clientFactory(cmd)
		if err != nil {
			return err
		}

		data, err := fetchPolicy(conjurClient, branch, returnJSON, depth, limit)
		if err != nil {
			return err
		}

		if output == "json" {
			if prettyData, err := utils.PrettyPrintJSON(data); err == nil {
				data = prettyData
			}
		}

		if file == "" {
			cmd.Println(string(data))
		} else {
			// Write data to a file
			err := os.WriteFile(file, data, 0644)
			if err != nil {
				return err
			}
			cmd.Println("Policy has been fetched and saved to " + file)
		}
		warningMsg := "\nWarning: The effective policy's output may not fully replicate " +
			"the policy defined in Conjur. If you try to upload the output to Conjur, the upload may fail."
		cmd.Println(warningMsg)

		return nil
	}
}

func validateFilePath(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		return errors.New("directory " + dir + " does not exist")
	}
	return nil
}

func cmdMessage(dryrun bool) string {
	if dryrun {
		return "Dry run"
	}
	return "Loaded"
}

func DryRunOrLoadPolicy(conjurClient policyClient, dryrun bool, policyMode conjurapi.PolicyMode, branch string, inputReader io.Reader) ([]byte, error) {
	if dryrun {
		return DryRunPolicy(conjurClient, policyMode, branch, inputReader)
	}
	return LoadPolicy(conjurClient, policyMode, branch, inputReader)
}

func DryRunPolicy(conjurClient policyClient, policyMode conjurapi.PolicyMode, branch string, inputReader io.Reader) ([]byte, error) {
	policyResponse, err := conjurClient.DryRunPolicy(
		policyMode,
		branch,
		inputReader,
	)

	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(policyResponse)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func LoadPolicy(conjurClient policyClient, policyMode conjurapi.PolicyMode, branch string, inputReader io.Reader) ([]byte, error) {
	policyResponse, err := conjurClient.LoadPolicy(
		policyMode,
		branch,
		inputReader,
	)

	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(policyResponse)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func fetchPolicy(conjurClient policyClient, branch string, returnJSON bool, policyTreeDepth uint, sizeLimit uint) ([]byte, error) {
	return conjurClient.FetchPolicy(
		branch,
		returnJSON,
		policyTreeDepth,
		sizeLimit,
	)
}

func newPolicyCommand(clientFactory policyClientFactoryFunc) *cobra.Command {
	policyCmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage Conjur policies",
	}

	policyCmd.PersistentFlags().StringP("branch", "b", "", "The parent policy branch")
	policyCmd.MarkPersistentFlagRequired("branch")

	policyCmd.AddCommand(newPolicyFetchCommand(clientFactory))
	policyCmd.AddCommand(newPolicyLoadCommand(clientFactory))
	policyCmd.AddCommand(newPolicyUpdateCommand(clientFactory))
	policyCmd.AddCommand(newPolicyReplaceCommand(clientFactory))

	return policyCmd
}

func newPolicyLoadCommand(clientFactory policyClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load a policy and create resources",
		Long: `Load a policy and create resources.

Examples:
- conjur policy load -b staging -f /policy/staging.yml`,
		SilenceUsage: true,
		RunE:         loadPolicyCommandRunner(clientFactory, conjurapi.PolicyModePost),
	}
	cmd.PersistentFlags().StringP("file", "f", "", "The policy file to load")
	cmd.PersistentFlags().BoolP("dry-run", "", false, "Dry run mode (input policy will be validated without applying the changes)")

	cmd.MarkPersistentFlagRequired("file")

	return cmd
}

func newPolicyFetchCommand(clientFactory policyClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch effective policy",
		Long: `Fetch effective policy.

Examples:
- conjur policy fetch -b staging -o json`,
		SilenceUsage: true,
		RunE:         fetchPolicyCommandRunner(clientFactory),
	}

	cmd.PersistentFlags().UintP("depth", "D", 64, "The max depth of fetched policy")
	cmd.PersistentFlags().UintP("limit", "l", 100000, "The max size of the output")
	cmd.PersistentFlags().StringP("output", "o", "yaml", "The output format")
	cmd.PersistentFlags().StringP("file", "f", "", "The file to save output to")

	return cmd
}

func newPolicyUpdateCommand(clientFactory policyClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update existing resources in the policy or create new resources",
		Long: `Update existing resources in the policy or create new resources.

Examples:
- conjur policy update -b staging -f /policy/staging.yml`,
		SilenceUsage: true,
		RunE:         loadPolicyCommandRunner(clientFactory, conjurapi.PolicyModePatch),
	}

	cmd.PersistentFlags().StringP("file", "f", "", "The policy file to load")
	cmd.PersistentFlags().BoolP("dry-run", "", false, "Dry run mode (input policy will be validated without applying the changes)")

	cmd.MarkPersistentFlagRequired("file")

	return cmd
}

func newPolicyReplaceCommand(clientFactory policyClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace",
		Short: "Fully replace an existing policy",
		Long: `Fully replace an existing policy.

Examples:
- conjur policy replace -b staging -f /policy/staging.yml`,
		SilenceUsage: true,
		RunE:         loadPolicyCommandRunner(clientFactory, conjurapi.PolicyModePut),
	}

	cmd.PersistentFlags().StringP("file", "f", "", "The policy file to load")
	cmd.PersistentFlags().BoolP("dry-run", "", false, "Dry run mode (input policy will be validated without applying the changes)")

	cmd.MarkPersistentFlagRequired("file")

	return cmd

}

type policyClient interface {
	LoadPolicy(mode conjurapi.PolicyMode, policyBranch string, policySrc io.Reader) (*conjurapi.PolicyResponse, error)
	DryRunPolicy(mode conjurapi.PolicyMode, policyBranch string, policySrc io.Reader) (*conjurapi.DryRunPolicyResponse, error)
	FetchPolicy(policyBranch string, returnJSON bool, policyTreeDepth uint, sizeLimit uint) ([]byte, error)
}

type policyClientFactoryFunc func(*cobra.Command) (policyClient, error)

func policyClientFactory(cmd *cobra.Command) (policyClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func init() {
	policyCmd := newPolicyCommand(policyClientFactory)

	rootCmd.AddCommand(policyCmd)
}
