module github.com/cyberark/conjur-cli-go

go 1.23.7

// Use the replace below for local development with conjur-api-go
// replace github.com/cyberark/conjur-api-go => ./conjur-api-go

require (
	github.com/AlecAivazis/survey/v2 v2.3.7 // Run "go get github.com/AlecAivazis/survey/v2@debug-windows" to update (Until https://github.com/go-survey/survey/pull/474 is merged)
	github.com/Netflix/go-expect v0.0.0-20220104043353-73e0943537d2
	github.com/creack/pty v1.1.24
	github.com/cyberark/conjur-api-go v0.12.13 // Run "go get github.com/cyberark/conjur-api-go@main" to update
	github.com/hinshun/vt10x v0.0.0-20220301184237-5011da428d02
	github.com/mattn/go-isatty v0.0.20
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
)

require (
	al.essio.dev/pkg/shellescape v1.5.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/zalando/go-keyring v0.2.6 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/term v0.27.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c => gopkg.in/yaml.v3 v3.0.1

replace golang.org/x/net v0.0.0-20190620200207-3b0461eec859 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.0.0-20220722155237-a158d28d115b => golang.org/x/net v0.33.0

replace golang.org/x/net v0.6.0 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.10.0 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.15.0 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.21.0 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.25.0 => golang.org/x/net v0.33.0

// DO NOT REMOVE: WE WANT THIS LINE TO PREVENT ACCIDENTALLY COMMITTING A VERSION OF conjur-api-go WHEN UPDATING DEPENDENCIES
replace github.com/cyberark/conjur-api-go => github.com/cyberark/conjur-api-go latest
