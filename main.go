package main

import (
	"os"
	"sort"

	"github.com/cyberark/conjur-api-go/conjurapi"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func Api(app *cli.App) *conjurapi.Client {
	return app.Metadata["api"].(*conjurapi.Client)
}

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Usage = "A CLI for Conjur"

	log.SetLevel(log.DebugLevel)

	config := conjurapi.LoadConfig()

	api, err := conjurapi.NewClientFromEnvironment(config)
	if err != nil {
		log.Errorf("Failed creating a Conjur client: %s\n", err.Error())
		os.Exit(1)
	}
	app.Metadata = make(map[string]interface{})
	app.Metadata["api"] = api

	app.Commands = []cli.Command{
		AuthnCommands,
		InitCommands,
		PolicyCommands,
		VariableCommands,
	}
	app.Commands = append(app.Commands, ResourceCommands...)

	sort.Sort(cli.CommandsByName(app.Commands))

	err = app.Run(os.
		Args)
	if err != nil {
		log.Fatal(err)
	}
}
