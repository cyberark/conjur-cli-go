# conjur-cli-go

A CLI for Conjur written in Golang using the [Cobra CLI framework](https://github.com/spf13/cobra).

## Development

There's a `docker-compose.yml` file in the top-level directory of this repository. It defines the following development services:
- dev: Runs using the image golang:1.19 and mounts the repository into the working directory whose absolute path is equal to the result of resolving `$PWD` on your host machine.

Running `docker-compose up -d` will start the development services.
Running `docker-compose exec dev bash` will create a bash shell into the `dev` service. The `dev` service is where you can proceed to run `go` commands to run, build, test the project.

## Conjur Server

Run a Conjur server using [conjur-quickstart](https://github.com/cyberark/conjur-quickstart). You can then point the development CLI to that Conjur server, and carry on developing.

To test out login, remember to set the admin password, for example you can set the password using the quickstart CLI
`conjur user update_password -p 'SuperSecret!!!!123'`

## Running

```
go run main.go
```

## Building

```
go build main.go
```

## Adding New Commands

To stub out a new command, [use the cobra-cli tool](https://github.com/spf13/cobra-cli/blob/main/README.md#add-commands-to-a-project).
