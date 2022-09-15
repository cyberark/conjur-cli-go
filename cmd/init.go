package cmd

import (
	"fmt"
	"os"
	"io"
	"path/filepath"
	"errors"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/manifoldco/promptui"
)

const conjurrcFmt = `---
account: %s
plugins: []
appliance_url: %s
`

func generateConjurrc(account string, applianceUrl string) string {
	return fmt.Sprintf(conjurrcFmt, account, applianceUrl)
}

func NopReadCloser(r io.Reader) io.ReadCloser {
	return nopCloser{Reader: r}
}

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopCloser{Writer: w}
}

type nopCloser struct {
	io.Reader
	io.Writer
}

func (nopCloser) Close() error { return nil }

func genInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the Conjur configuration",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
	
			account := cmd.Flag("account").Value.String()
			applianceUrl := cmd.Flag("url").Value.String()
			filePath := cmd.Flag("file").Value.String()
	
			if len(applianceUrl) == 0 {
				prompt := promptui.Prompt{
					Label:    "Enter the URL of your Conjur service",
					Stdin: NopReadCloser(cmd.InOrStdin()),
					Stdout: NopWriteCloser(cmd.OutOrStdout()),
					Validate: func(input string) error {
						if len(input) == 0 {
							return errors.New("URL is required")
						}
	
						_, err := url.ParseRequestURI(input)
						return err
					},
				}
				applianceUrl, err = prompt.Run()
				if err != nil {
					return err
				}
			}
	
			if len(account) == 0 {
				prompt := promptui.Prompt{
					Label:    "Enter your organization account name",
					Validate: func(input string) error {
						if len(input) == 0 {
							return errors.New("Account is required")
						}
						return nil
					},
				}
				account, err = prompt.Run()
				if err != nil {
					return err
				}
			}
	
			if _, err := os.Stat(filePath); err == nil {
				prompt := promptui.Prompt{
					Label:     fmt.Sprintf("File %s exists. Overwrite", filePath),
					IsConfirm: true,
				}
			
				_, err := prompt.Run()
				if err != nil {
					return fmt.Errorf("Not overwriting %s", filePath)
				}
			}
	
			fileContents := generateConjurrc(account, applianceUrl)
			err = os.WriteFile(filePath, []byte(fileContents), 0644)
			if err != nil  {
				return err
			}
	
			cmd.Printf("Wrote configuration to %s\n", filePath)
			return nil
		},
	}

	initCmdFlags(cmd)

	return cmd
}


func initCmdFlags(cmd *cobra.Command) {
	// TODO: this can actually return an error
	userHomeDir, _ := os.UserHomeDir()

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	cmd.PersistentFlags().StringP("account", "a", "", "Conjur organization account name")
	cmd.PersistentFlags().StringP("url", "u", "", "URL of the Conjur service")
	cmd.PersistentFlags().BoolP("help", "h", false, "Help for init command") // TODO: maybe change this for everything
	cmd.PersistentFlags().StringP("certificate", "c", "", "Conjur SSL certificate (will be obtained from host unless provided by this option)")
	cmd.PersistentFlags().StringP("file", "f", filepath.Join(userHomeDir, ".conjurrc"), "File to write the configuration to")
	cmd.PersistentFlags().Bool("force", false, "Force overwrite of existing file")
	

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func init() {
	initCmd := genInitCmd()
	rootCmd.AddCommand(initCmd)

	// initCmdFlags(initCmd)
}
