# conjur-cli-go

A CLI for Conjur written in Golang using the [Cobra CLI framework](https://github.com/spf13/cobra).

## Certification level

![](https://img.shields.io/badge/Certification%20Level-Trusted-28A745?link=https://github.com/cyberark/community/blob/master/Conjur/conventions/certification-levels.md)

This repo is a **Trusted** level project. It's is supported by CyberArk and has
been verified to work with Conjur Open Source. For more detailed information on
our certification levels, see
[our community guidelines](https://github.com/cyberark/community/blob/master/Conjur/conventions/certification-levels.md#trusted).

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

## Transitioning from the docker based CLI to 8.x

See the [transition guide](docs/UPGRADE_from_docker_based.md)

## FIPS Compatibility

The `amd64` binaries are built using RedHat's patched Go compiler and with
`GOEXPERIMENT=boringcrypto`. When run on a FIPS-enabled system, the binary will
use the OpenSSL FIPS module provided by the system. On non-FIPS systems, the
binary will fall back to BoringCrypto.
