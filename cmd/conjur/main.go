package main

import (
	"crypto/fips140"
	"log"

	"github.com/cyberark/conjur-cli-go/pkg/cmd"
)

func main() {
	if !fips140.Enabled() {
		log.Print("WARN: Compliance with FIPS 140-3 is not enabled")
	}

	cmd.Execute()
}
