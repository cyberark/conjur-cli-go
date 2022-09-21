package cmd

import (
	"github.com/manifoldco/promptui"
)

func newPasswordPrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Please enter your password (it will not be echoed)",
		Mask:  '*',
	}
}

func newUsernamePrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Enter your username to log into Conjur",
	}
}

func mightString(v string, e error) string {
	if e != nil {
		return ""
	}

	return v
}

func mightBool(v bool, e error) bool {
	if e != nil {
		return false
	}

	return v
}

// TODO: whenever this is called we should store to .netrc
func requestCredentials(decoratePrompt decoratePromptFunc, username string, password string) (string, string, error) {
	var err error

	if len(username) == 0 {
		usernamePrompt := decoratePrompt(newUsernamePrompt())
		username, err = runPrompt(usernamePrompt)
		if err != nil {
			return "", "", err
		}
	}

	if len(password) == 0 {
		passwordPrompt := decoratePrompt(newPasswordPrompt())
		password, err = runPrompt(passwordPrompt)
		if err != nil {
			return "", "", err
		}
	}

	return username, password, err
}
