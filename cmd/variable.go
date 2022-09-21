package cmd

import (
	"github.com/spf13/cobra"
)

var variableCmd = &cobra.Command{
	Use:   "variable",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

var variableGetCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := cmd.Flags().GetString("id")
		if err != nil {
			return err
		}

		conjurClient, err := authenticatedConjurClientForCommand(cmd)
		if err != nil {
			return err
		}

		// TODO: update RetrieveSecret to accept version
		data, err := conjurClient.RetrieveSecret(id)
		if err != nil {
			return err
		}

		cmd.Println(string(data))
		return nil
	},
}

var variableSetCmd = &cobra.Command{
	Use:   "set",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		conjurClient, err := authenticatedConjurClientForCommand(cmd)
		if err != nil {
			return err
		}

		err = conjurClient.AddSecret(id, value)
		if err != nil {
			return err
		}

		cmd.Println("Value added")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(variableCmd)

	variableCmd.AddCommand(variableGetCmd)
	variableCmd.AddCommand(variableSetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	variableGetCmd.Flags().StringP("id", "i", "", "Provide variable identifier")
    variableGetCmd.MarkFlagRequired("id")
	variableGetCmd.Flags().StringP("version", "v", "", "Specify the desired version of a single variable value")

	variableSetCmd.Flags().StringP("id", "i", "", "Provide variable identifier")
    variableSetCmd.MarkFlagRequired("id")
	variableSetCmd.Flags().StringP("value", "v", "", "Set the value of the specified variable")
	variableSetCmd.MarkFlagRequired("value")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// variableCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
