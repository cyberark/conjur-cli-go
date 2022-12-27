package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

var roleExistsCmd = &cobra.Command{
	Use:   "exists",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("role exists called")
	},
}

var roleMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("role members called")
	},
}

var roleMembershipsCmd = &cobra.Command{
	Use:   "memberships",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("role memberships called")
	},
}

func init() {
	rootCmd.AddCommand(roleCmd)

	roleCmd.AddCommand(roleExistsCmd)
	roleCmd.AddCommand(roleMembersCmd)
	roleCmd.AddCommand(roleMembershipsCmd)
}
