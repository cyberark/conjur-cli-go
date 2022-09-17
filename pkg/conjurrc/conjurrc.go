package conjurrc

import (
	"fmt"
	"os"
)

const conjurrcFmt = `---
account: %s
plugins: []
appliance_url: %s
`

func generateConjurrc(account string, applianceURL string) string {
	return fmt.Sprintf(conjurrcFmt, account, applianceURL)
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

	return os.WriteFile(filePath, []byte(fileContents), 0644)
}
