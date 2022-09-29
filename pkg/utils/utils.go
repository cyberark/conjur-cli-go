package utils

import "github.com/manifoldco/promptui"

type DecoratePromptFunc func(*promptui.Prompt) *promptui.Prompt

func RunPrompt(prompt *promptui.Prompt) (userInput string, err error) {
	userInput, err = prompt.Run()
	if err != nil {
		return "", err
	}
	return userInput, nil
}

func MightString(v string, e error) string {
	if e != nil {
		return ""
	}

	return v
}

func MightBool(v bool, e error) bool {
	if e != nil {
		return false
	}

	return v
}
