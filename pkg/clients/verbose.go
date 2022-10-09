package clients

import (
	"net/http"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/spf13/cobra"
)

// MaybeVerboseLoggingForClient optionally carries out verbose logging of HTTP requests and responses on a Conjur client
func MaybeVerboseLoggingForClient(
	verbose bool,
	cmd *cobra.Command,
	client *conjurapi.Client,
) {
	if !verbose {
		return
	}

	if client == nil {
		return
	}
	// TODO: check to see if the transport has already been decorated to make this idempotent

	httpClient := client.GetHttpClient()
	transport := httpClient.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	httpClient.Transport = utils.NewDumpTransport(
		transport,
		func(dump []byte) {
			cmd.PrintErrln(string(dump))
			cmd.PrintErrln()
		},
	)
}
