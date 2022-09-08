package cmd

import (
 	"fmt"

	"github.com/spf13/cobra"
)

var authnCmd = &cobra.Command{
	Use:   "authn",
	Short: "Commands for managing authentication",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

var authnLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("authn login called")
	},
}

var authnLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("authn logout called")
	},
}

var authnWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("authn whoami called")
	},
}

func init() {
	rootCmd.AddCommand(authnCmd)

	authnCmd.AddCommand(authnLoginCmd)
	authnLoginCmd.Flags().StringP("username", "u", "", "")
	authnLoginCmd.Flags().StringP("password", "p", "", "")

        authnCmd.AddCommand(authnLogoutCmd)

        authnCmd.AddCommand(authnWhoamiCmd)
}
