module github.com/cyberark/conjur-cli-go

go 1.19

// Use the replace below for local development with conjur-api-go
// replace github.com/cyberark/conjur-api-go => ./conjur-api-go

require (
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d
	github.com/cyberark/conjur-api-go v0.10.3-0.20221201205217-6df567cf841f // Run "go get github.com/cyberark/conjur-api-go@main" to update
	github.com/manifoldco/promptui v0.9.0
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8

require (
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.0.0-20211214234402-4825e8c3871d // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c => gopkg.in/yaml.v3 v3.0.1
