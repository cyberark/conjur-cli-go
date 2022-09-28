package conjurrc

import (
	"os"

	"gopkg.in/yaml.v3"
)

// conjurrc uses the YAML format
func generateConjurrc(account string, applianceUrl string) []byte {
	conjurrcMap := map[string]interface{}{
		"account":       account,
		"appliance_url": applianceUrl,
		"plugins":       []string{},
	}

	data, _ := yaml.Marshal(&conjurrcMap)
	return data
}

// WriteConjurrc writes Conjur connection info to a file.
func WriteConjurrc(
	account string,
	applianceURL string,
	filePath string,
	overwriteDecision func(string) error,
) error {
	if _, err := os.Stat(filePath); err == nil {
		err := overwriteDecision(filePath)
		if err != nil {
			return err
		}
	}

	fileContents := generateConjurrc(account, applianceURL)

	return os.WriteFile(filePath, fileContents, 0644)
}
