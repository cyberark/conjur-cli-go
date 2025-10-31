# conjur-cli-go

A CLI for CyberArk Secrets Manager written in Golang using the [Cobra CLI framework](https://github.com/spf13/cobra).

## Certification Level

![](https://img.shields.io/badge/Certification%20Level-Certified-6C757D?link=https://github.com/cyberark/community/blob/main/Conjur/conventions/certification-levels.md)

This repo is a **Certified** level project. It's been reviewed by CyberArk to
verify that it will securely work with CyberArk Enterprise as documented. In
addition, CyberArk offers Enterprise-level support for these features. For more
detailed information on our certification levels, see
[our community guidelines](https://github.com/cyberark/community/blob/main/Conjur/conventions/certification-levels.md#certified).

## Development

See the [dev](dev/) directory for more details.

## Running

```
go run ./cmd/conjur
```

## Building

```
go build ./cmd/conjur
```

## Adding New Commands

To stub out a new command, [use the cobra-cli tool](https://github.com/spf13/cobra-cli/blob/main/README.md#add-commands-to-a-project).
