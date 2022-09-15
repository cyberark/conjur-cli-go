package conjurrc

import (
	"os"
	"fmt"
)

const conjurrcFmt = `---
account: %s
plugins: []
appliance_url: %s
`

func generateConjurrc(account string, applianceUrl string) string {
	return fmt.Sprintf(conjurrcFmt, account, applianceUrl)
}

func WriteConjurrc(
	account string, 
	applianceUrl string,
	filePath string, 
	overwriteDecision func(string) error,
) error {
	if _, err := os.Stat(filePath); err == nil {
		err := overwriteDecision(filePath)
		if err != nil {
			return err
		}
	}

	fileContents := generateConjurrc(account, applianceUrl)

	return os.WriteFile(filePath, []byte(fileContents), 0644)
}
