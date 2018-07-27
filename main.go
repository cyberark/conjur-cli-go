package main

import (
	"os"
	"sort"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/action"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// AppClient retrieves the Conjur client from the App's metadata.
func AppClient(app *cli.App) action.ConjurClient {
	return app.Metadata["api"].(action.ConjurClient)
}

var commands = [][]cli.Command{
	AuthnCommands,
	InitCommands,
	PolicyCommands,
	VariableCommands,
}

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Usage = "A CLI for Conjur"

	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true, DisableLevelTruncation: true})

	config := conjurapi.LoadConfig()

	client, err := conjurapi.NewClientFromEnvironment(config)
	if err != nil {
		log.Errorf("Failed creating a Conjur client: %s\n", err.Error())
		os.Exit(1)
	}
	app.Metadata = make(map[string]interface{})
	app.Metadata["api"] = action.ConjurClient(client)

	for _, cmds := range commands {
		app.Commands = append(app.Commands, cmds...)
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
