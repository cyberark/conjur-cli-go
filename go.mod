module github.com/cyberark/conjur-cli-go

go 1.25.5

// Use the replace below for local development with conjur-api-go
// replace github.com/cyberark/conjur-api-go => ./conjur-api-go

require (
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/fang v0.4.2
	github.com/charmbracelet/huh v0.7.1-0.20250908094625-904537be8706
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/charmbracelet/lipgloss/v2 v2.0.0-beta.3.0.20250917201909-41ff0bf215ea
	github.com/cyberark/conjur-api-go v0.13.2 // Run "go get github.com/cyberark/conjur-api-go@main" to update
	github.com/makiuchi-d/gozxing v0.1.1
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/spf13/cobra v1.10.1
	github.com/stretchr/testify v1.11.1
	github.com/wiremock/go-wiremock v1.14.0
	golang.org/x/exp v0.0.0-20250911091902-df9299821621
	golang.org/x/term v0.35.0
)

require (
	al.essio.dev/pkg/shellescape v1.6.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aws/aws-sdk-go-v2 v1.39.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.31.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.14 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.5 // indirect
	github.com/aws/smithy-go v1.23.0 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/catppuccin/go v0.3.0 // indirect
	github.com/charmbracelet/bubbles v0.21.1-0.20250623103423-23b8fd6302d7 // indirect
	github.com/charmbracelet/colorprofile v0.3.2 // indirect
	github.com/charmbracelet/ultraviolet v0.0.0-20250915111650-81d4262876ef // indirect
	github.com/charmbracelet/x/ansi v0.10.1 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13 // indirect
	github.com/charmbracelet/x/exp/charmtone v0.0.0-20250922100529-c9afca5d6f21 // indirect
	github.com/charmbracelet/x/exp/color v0.0.0-20250922100529-c9afca5d6f21 // indirect
	github.com/charmbracelet/x/exp/strings v0.0.0-20250922100529-c9afca5d6f21 // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/mango v0.2.0 // indirect
	github.com/muesli/mango-cobra v1.3.0 // indirect
	github.com/muesli/mango-pflag v0.2.0 // indirect
	github.com/muesli/roff v0.1.0 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/zalando/go-keyring v0.2.6 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
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
