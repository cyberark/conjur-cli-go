package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var hostfactoryCmd = &cobra.Command{
	Use:   "hostfactory",
	Short: "Commands for managing hostfactories",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

var hostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "Commands for managing hostfactory hosts",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Commands for managing hostfactory tokens",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

var hostsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hostfactory hosts create called")
	},
}

var tokensCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hostfactory tokens create called")
	},
}

var tokensRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hostfactory tokens revoke called")
	},
}

func init() {
	rootCmd.AddCommand(hostfactoryCmd)

	hostfactoryCmd.AddCommand(hostsCmd)
	hostsCmd.AddCommand(hostsCreateCmd)

	hostfactoryCmd.AddCommand(tokensCmd)
	tokensCmd.AddCommand(tokensCreateCmd)
	tokensCmd.AddCommand(tokensRevokeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hostfactoryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hostfactoryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
