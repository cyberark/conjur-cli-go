package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type InitOptions struct {
	Account     string
	Certificate string
	File        string
	Force       bool
	URL         string
}

type Initer interface {
	Do(options InitOptions) error
}

func NewIniter(fs afero.Fs) Initer {
	return initerImpl{
		fs: fs,
	}
}

type initerImpl struct {
	fs      afero.Fs
	options InitOptions
}

func (i initerImpl) Do(options InitOptions) (err error) {
	if err = i.checkOptions(options); err != nil {
		return
	}
	i.options = options

	var conjurrc *os.File
	if conjurrc, err = i.openConjurrc(); err != nil {
		return
	}
	defer conjurrc.Close()

	return i.writeConjurrc(conjurrc)
}

func (i initerImpl) checkOptions(options InitOptions) error {
	if options.URL == "" {
		return fmt.Errorf("appliance URL is required")
	}

	if options.Account == "" {
		return fmt.Errorf("account must is required")
	}

	if options.File == "" {
		options.File = path.Join(os.Getenv("HOME"), ".conjurrc")
	}

	if _, err := i.fs.Stat(options.File); err == nil && !options.Force {
		return fmt.Errorf("%s exists, but force not specified", options.File)
	}

	return nil
}

func (i initerImpl) openConjurrc() (conjurrc *os.File, err error) {
	flags := os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	file, err := i.fs.OpenFile(i.options.File, flags, 0644)
	return file.(*os.File), err
}

func (i initerImpl) writeConjurrc(conjurrc *os.File) (err error) {
	config := conjurapi.Config{
		Account:      i.options.Account,
		ApplianceURL: i.options.URL,
		SSLCert:      i.options.Certificate,
		V4:           false,
	}

	bytes, err := yaml.Marshal(config)
	if err != nil {
		return
	}

	buf := bufio.NewWriter(conjurrc)
	if _, err = buf.Write(bytes); err != nil {
		return
	}
	return buf.Flush()
}
