## Interface Changes

It was a product requirement to be fully backwards compatible with the
[Python CLI](https://github.com/cyberark/cyberark-conjur-cli), so the
Golang CLI has maintained that interface wherever possible.

However, the Python CLI has some inconsistencies when it comes to using
nouns and verbs for commands that manipulate policy resources (for example
`conjur policy replace` versus `conjur hostfactory create host`). The Golang
CLI includes additional commands (such as `conjur hostfactory hosts create`) to
make this more consistent and we have deprecated the commands that came from the
Python CLI for future removal.

The CLI package that we're using (cobra) is also incapable of handling
multi-letter flag shorthands, so we had to rework a few Python CLI commands
that make use of those.

Many commands from the Ruby (Docker-based) CLI were excluded from the Python CLI
or reworked to be arguments on other commands. The Golang CLI has restored most
of these commands as it seems like a simpler interface overall. Additional details
about differences from the Ruby CLI can be found in 
[this migration guide](https://github.com/cyberark/conjur-cli-go/blob/master/docs/UPGRADE_from_docker_based.md).

## Testing

We prioritized good unit test coverage and later included integration tests
against Conjur running in Docker. We try to avoid remote connections wherever
possible to ensure tests can be run quickly and in a local environment with
minimal setup.