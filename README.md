# conjur-cli-go

Golang-based CLI for [Conjur OSS v5+](https://cyberark.github.io/conjur) and
[DAP](https://docs.cyberark.com/Product-Doc/OnlineHelp/AAM-DAP/Latest/en/Content/Get%20Started/WhatIsConjur.html)

**THIS REPO IS UNDER DEVELOPMENT AND CODE MAY CHANGE SIGNIFFICANTY**

#### This repo's metrics:
![Tests](https://github.com/cyberark/conjur-oss-suite-release/workflows/Tests/badge.svg)

## Quick start

- Go to the [releases](https://github.com/repos/cyberark/conjur-cli-go/releases/latest) page
  and download the appropriate package for your system.
- Copy the binary to your `/usr/bin` directory and make it executable
```sh-session
$ sudo cp conjur-cli-go /usr/bin/conjur-cli-go
$ chmod +x /usr/bin/conjur-cli-go
```
- Start using it!
```sh-session
$ conjur-cli-go -version
conjur-cli-go version 0.0.2
```

## Usage

Currently the tool supports these high0level commands:
```sh-session
COMMANDS:
     authn     Login and logout
     init      Initialize the Conjur configuration
     list      List objects
     policy    Manage policies
     show      Show an object
     variable  Manage variables
     help, h   Shows a list of commands or help for one command
```

## Development

We welcome contributions of all kinds to this project. For instructions on how to
get started, instructions for using the suite release tooling in this project, and
descriptions of our development workflows - please see our [contributing guide](CONTRIBUTING.md).

## License

This repository is licensed under Apache License 2.0 - see [`LICENSE`](LICENSE)
for more details.
