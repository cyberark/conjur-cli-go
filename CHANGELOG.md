# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Nothing should go in this section, please add to the latest unreleased version
  (and update the corresponding date), or add a new version.

## [8.0.19] - 2025-02-03

### Fixed
- Fix false positive on version check
  (CNJR-7876, [cyberark/conjur-cli-go#152](https://github.com/cyberark/conjur-cli-go/issues/152))

## [8.0.18] - 2025-01-10

### Security
- Update multiple dependencies to latest versions

## [8.0.17] - 2024-12-27

### Added
- Add user-friendly error message on timeout and add HTTP timeout flag.
  CNJR-7343 [#150](https://github.com/cyberark/conjur-cli-go/issues/150)

### Fixed
- `role memberships` command now returns all memberships, including inherited memberships (CNJR-5213)
- Checks Conjur server version before using policy dry-run and fetch
  (CNJR-7207, [#149](https://github.com/cyberark/conjur-cli-go/issues/149))

### Security
- Run the CLI as a non-root user in the Docker image (CNJR-6439)
- Use Sha256 certificate fingerprint when confirming server's authenticity (CNJR-6438)
- Change base image to `ubi9/ubi-minimal:9.5` (CNJR-7189)

## [8.0.16] - 2024-07-25

### Added
- Support for `-jwt-file` and `-jwt-host-id` init params for use with JWT authentication (CNJR-4966)
- Publish `latest` tag for Docker image (CNJR-5310)
- Add `--dry-run` support for the policy operations (CNJR-4593)
- Add `fetch` subcommand for policy commands to retrieve policy from Conjur (CNJR-5673)
- Return requested variable duplicate instead of displaying 500 error (CNJR-5006)
- Fixed a json when using batch retrieval of secrets (CNJR-6362)

### Fixed
- Fixed bug causing JSON output to be HTML escaped (CNJR-4574)

### Security
- Updated to Go 1.22 and use Microsoft's Golang image for FIPS compliance (CNJR-4923)

## [8.0.15] - 2024-03-19

### Fixed
- Restored `tar` binary to container to allow for `kubectl cp` (CNJR-3913)

## [8.0.14] - 2024-03-14

### Changed
- Change base image to `ubi9/ubi-minimal:latest` (CNJR-3913)

## [8.0.13] - 2024-03-14

### Added
- OIDC redirect URI supports IPv6 addresses (CNJR-3851)

### Fixed
- FIPS-compliant binaries work on RHEL 7/8 (CNJR-3544)
- FIPS-compliant binaries report correct version when passed `--version` flag
  (CNJR-3547)
- RPM and DEB packages contain correct version in installed package names
  (CNJR-3547)

### Security
- Upgrade golang.org/x/net to v0.17.0, golang.org/x/crypto to v0.14.0,
  golang.org/x/sys to v0.13.0, golang.org/x/text to v0.13.0, and
  golang.org/x/term to v0.13.0 (CNJR-3913)

## [8.0.12] - 2023-10-17

### Fixed
- Update busybox container image to 1.36.1
  [cyberark/conjur-cli-go#147](https://github.com/cyberark/conjur-cli-go/pull/147)

## [8.0.11] - 2023-08-25

### Fixed
- Handle trailing slash on appliance URL
  [cyberark/conjur-cli-go#142](https://github.com/cyberark/conjur-cli-go/pull/142)
- Allow API key rotation for logged-in host
  [cyberark/conjur-cli-go#143](https://github.com/cyberark/conjur-cli-go/pull/143)
- Make `amd64` binary FIPS compliant on FIPS-enabled systems
  [cyberark/conjur-cli-go#145](https://github.com/cyberark/conjur-cli-go/pull/145)

## [8.0.10] - 2023-06-29

### Security
- Upgrade golang.org/x/net to v0.10.0
  [cyberark/conjur-cli-go#139](https://github.com/cyberark/conjur-cli-go/pull/139)
- Upgrade golang.org/x/net to v0.10.0, golang.org/x/crypto to v0.9.0,
  golang.org/x/sys to v0.8.0, golang.org/x/text to v0.9.0, and Go to 1.20
  [cyberark/conjur-cli-go#138](https://github.com/cyberark/conjur-cli-go/pull/138)

### Fixed
- Fixed missing example commands in help output
  [cyberark/conjur-cli-go#134](https://github.com/cyberark/conjur-cli-go/pull/134)

## [8.0.9] - 2023-04-21

### Security
- Redact credentials dumped to logs with `--debug` flag
  [cyberark/conjur-cli-go#130](https://github.com/cyberark/conjur-cli-go/pull/130)

## [8.0.8] - 2023-04-19

### Fixed
- Fixed piping input to `conjur init` confirmation prompts
  [cyberark/conjur-cli-go#127](https://github.com/cyberark/conjur-cli-go/pull/127)
- Made command help text more consistent
  [cyberark/conjur-cli-go#123](https://github.com/cyberark/conjur-cli-go/pull/123)

## [8.0.7] - 2023-04-18

### Fixed
- Fixed not using hosts file on Windows
  [cyberark/conjur-cli-go#121](https://github.com/cyberark/conjur-cli-go/pull/121)

## [8.0.6] - 2023-04-17

### Fixed
- Improved error message when using self-signed certificates
  [cyberark/conjur-cli-go#119](https://github.com/cyberark/conjur-cli-go/pull/119)
- Fix double prompt in Windows
  [cyberark/conjur-cli-go#120](https://github.com/cyberark/conjur-cli-go/pull/120)

## [8.0.5] - 2023-03-24

### Changed
- OIDC login now supports a custom redirect URL port
  [cyberark/conjur-cli-go#117](https://github.com/cyberark/conjur-cli-go/pull/117)

### Fixed
- Reject OIDC login if configured port is in use
  [cyberark/conjur-cli-go#117](https://github.com/cyberark/conjur-cli-go/pull/117)

## [8.0.4] - 2023-03-03

### Fixed
- Allow hostfactory cidrs to specify a subnet
  [cyberark/conjur-cli-go#113](https://github.com/cyberark/conjur-cli-go/pull/113)
- Update variable get to retrieve multiple variables
  [cyberark/conjur-cli-go#114](https://github.com/cyberark/conjur-cli-go/pull/114)

## [8.0.3] - 2023-02-21

### Fixed
- Fix rotating api key of the logged-in user
  [cyberark/conjur-cli-go#107](https://github.com/cyberark/conjur-cli-go/pull/107)
- Support fully- and partially-qualified IDs for API key rotation
  [cyberark/conjur-cli-go#111](https://github.com/cyberark/conjur-cli-go/pull/111)

### Added
- Added Transition guide from the Docker based CLI
  [cyberark/conjur-cli-go#86](https://github.com/cyberark/conjur-cli-go/pull/86)

## [8.0.2] - 2023-02-16

### Fixed
- Fix default value of `hostfactory tokens create --cidr` parameter.
  [cyberark/conjur-cli-go#105](https://github.com/cyberark/conjur-cli-go/pull/105)

## [8.0.1] - 2023-02-15

### Added
- Update Hostfactory help as the fully qualified hostfactory name is no longer required.
  [cyberark/conjur-cli-go#104](https://github.com/cyberark/conjur-cli-go/pull/104)

## [8.0.0] - 2023-02-09

### Added
- Initial release of Conjur CLI written in Golang

## [0.0.0] - 2023-01-01

### Added
- Placeholder version to capture the reset of the repository

[Unreleased]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.16...HEAD
[8.0.16]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.15...v8.0.16
[8.0.15]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.14...v8.0.15
[8.0.14]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.13...v8.0.14
[8.0.13]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.12...v8.0.13
[8.0.12]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.11...v8.0.12
[8.0.11]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.10...v8.0.11
[8.0.10]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.9...v8.0.10
[8.0.9]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.8...v8.0.9
[8.0.8]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.7...v8.0.8
[8.0.7]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.6...v8.0.7
[8.0.6]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.5...v8.0.6
[8.0.5]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.4...v8.0.5
[8.0.4]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.3...v8.0.4
[8.0.3]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.2...v8.0.3
[8.0.2]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.1...v8.0.2
[8.0.1]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.0...v8.0.1
[8.0.0]: https://github.com/cyberark/conjur-cli-go/releases/tag/v8.0.0
