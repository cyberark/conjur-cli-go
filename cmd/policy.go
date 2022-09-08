package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("policy load called")
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("policy append called")
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("policy replace called")
	},
}

func init() {
	rootCmd.AddCommand(policyCmd)

	policyCmd.PersistentFlags().StringP("branch", "b", "", "The parent policy branch")
        policyCmd.PersistentFlags().StringP("filepath", "f", "", "The policy file to load")

	policyCmd.AddCommand(policyLoadCmd)
	policyCmd.AddCommand(policyAppendCmd)
	policyCmd.AddCommand(policyReplaceCmd)
}
