package conjurrc

import (
	"os"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

// WriteConjurrc writes Conjur connection info to a file.
func WriteConjurrc(
	config conjurapi.Config,
	filePath string,
	overwriteDecision func(string) error,
) error {
	if _, err := os.Stat(filePath); err == nil {
		err := overwriteDecision(filePath)
		if err != nil {
			return err
		}
	}

	fileContents := config.Conjurrc()

	return os.WriteFile(filePath, fileContents, 0644)
}
