package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

var resourceExistsCmd = &cobra.Command{
	Use:   "exists",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("resource exists called")
	},
}

var resourceShowCmd = &cobra.Command{
	Use:   "show",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("resource show called")
	},
}

var resourcePermittedRolesCmd = &cobra.Command{
	Use:   "permitted-roles",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("resource permitted-roles called")
	},
}

func init() {
	rootCmd.AddCommand(resourceCmd)

	resourceCmd.AddCommand(resourceExistsCmd)
	resourceCmd.AddCommand(resourceShowCmd)
	resourceCmd.AddCommand(resourcePermittedRolesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resourceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resourceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
