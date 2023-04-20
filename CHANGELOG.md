# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Nothing should go in this section, please add to the latest unreleased version
  (and update the corresponding date), or add a new version.

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

[Unreleased]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.8...HEAD
[8.0.8]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.7...v8.0.8
[8.0.7]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.6...v8.0.7
[8.0.6]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.5...v8.0.6
[8.0.5]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.4...v8.0.5
[8.0.4]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.3...v8.0.4
[8.0.3]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.2...v8.0.3
[8.0.2]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.1...v8.0.2
[8.0.1]: https://github.com/cyberark/conjur-cli-go/compare/v8.0.0...v8.0.1
[8.0.0]: https://github.com/cyberark/conjur-cli-go/releases/tag/v8.0.0
