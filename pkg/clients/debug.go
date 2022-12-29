package clients

import (
	"net/http"

	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/spf13/cobra"
)

// MaybeDebugLoggingForClient optionally carries out debug logging of HTTP requests and responses on a Conjur client
func MaybeDebugLoggingForClient(
	debug bool,
	cmd *cobra.Command,
	client ConjurClient,
) {
	if !debug {
		return
	}

	if client == nil {
		return
	}

	httpClient := client.GetHttpClient()
	if httpClient == nil {
		return
	}

	// TODO: check to see if the transport has already been decorated to make this idempotent
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
