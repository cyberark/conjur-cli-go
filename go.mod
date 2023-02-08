module github.com/cyberark/conjur-cli-go

go 1.19

// Use the replace below for local development with conjur-api-go
// replace github.com/cyberark/conjur-api-go => ./conjur-api-go

require (
	github.com/AlecAivazis/survey/v2 v2.3.7-0.20221124181729-2363ef255e14 // Run "go get github.com/AlecAivazis/survey/v2@debug-windows" to update (Until https://github.com/go-survey/survey/pull/474 is merged)
	github.com/Netflix/go-expect v0.0.0-20220104043353-73e0943537d2
	github.com/creack/pty v1.1.18
	github.com/cyberark/conjur-api-go v0.10.3-0.20230208145439-074a215fc0de // Run "go get github.com/cyberark/conjur-api-go@main" to update
	github.com/hinshun/vt10x v0.0.0-20220301184237-5011da428d02
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.1
)

require github.com/manifoldco/promptui v0.9.0

require (
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/zalando/go-keyring v0.2.2 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.4.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c => gopkg.in/yaml.v3 v3.0.1
