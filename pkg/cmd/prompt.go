package cmd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/manifoldco/promptui"
)

type decoratePromptFunc func(*promptui.Prompt) *promptui.Prompt

func newApplianceURLPrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Enter the URL of your Conjur service",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("URL is required")
			}

			_, err := url.ParseRequestURI(input)
			return err
		},
	}
}

func newAccountPrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Enter your organization account name",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("Account is required")
			}
			return nil
		},
	}
}

func newFileExistsPrompt(filePath string) *promptui.Prompt {
	return &promptui.Prompt{
		Label:     fmt.Sprintf("File %s exists. Overwrite", filePath),
		IsConfirm: true,
	}
}

func runPrompt(prompt *promptui.Prompt) (userInput string, err error) {
	userInput, err = prompt.Run()
	if err != nil {
		return "", err
	}
	return userInput, nil
}
