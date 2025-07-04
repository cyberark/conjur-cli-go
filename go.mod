module github.com/cyberark/conjur-cli-go

go 1.24.2

// Use the replace below for local development with conjur-api-go
// replace github.com/cyberark/conjur-api-go => ./conjur-api-go

require (
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/fang v0.4.2
	github.com/charmbracelet/huh v0.7.1-0.20250603124601-31a1db2cbc39
	github.com/charmbracelet/lipgloss/v2 v2.0.0-beta1
	github.com/cyberark/conjur-api-go v0.12.14 // Run "go get github.com/cyberark/conjur-api-go@main" to update
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394
	golang.org/x/term v0.30.0
)

require github.com/charmbracelet/lipgloss v1.1.0

require (
	al.essio.dev/pkg/shellescape v1.6.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/catppuccin/go v0.3.0 // indirect
	github.com/charmbracelet/bubbles v0.21.0 // indirect
	github.com/charmbracelet/colorprofile v0.3.2 // indirect
	github.com/charmbracelet/x/ansi v0.10.1 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13 // indirect
	github.com/charmbracelet/x/exp/charmtone v0.0.0-20250603201427-c31516f43444 // indirect
	github.com/charmbracelet/x/exp/strings v0.0.0-20240722160745-212f7b056ed0 // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/mango v0.1.0 // indirect
	github.com/muesli/mango-cobra v1.2.0 // indirect
	github.com/muesli/mango-pflag v0.1.0 // indirect
	github.com/muesli/roff v0.1.0 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/zalando/go-keyring v0.2.6 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.24.0 // indirect
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
