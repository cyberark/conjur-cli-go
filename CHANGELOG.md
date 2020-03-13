# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.2] - 2020-03-13

### Added
- `list` and `show` resource subcommands.
- `policy load` subcommand.
- `variable value` and `variable values add` subcommands.
- Goreleaser can be used now to build static Linux and OSX binaries

### Changed
- Base Golang version bumped to v1.13
- Project converted to use Golang modules
- Updated `conjur-api-go` dependency to v0.6.0
- Internal project structure cleaned up a bit
- New and improved changelog.

## 0.0.1 - 2018-07-22

### Added
- Initial commit of an outline of a Conjur CLI written in GO.

[Unreleased]: https://github.com/cyberark/conjur-cli-go/compare/v0.0.2...HEAD
[0.2.0]: https://github.com/cyberark/conjur-cli-go/compare/v0.1.0...v0.2.0
