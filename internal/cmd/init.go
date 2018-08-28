package cmd

import (
	"bufio"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

// InitOptions are the options used to initialize the API configuration
type InitOptions struct {
	Account     string
	Certificate string
	File        string
	Force       bool
	URL         string
}

// Initer performs the operations necessary to initialize the API configuration
type Initer interface {
	Do(options InitOptions) error
}

// NewIniter returns an Initer that uses fs to write the API configuration
func NewIniter(fs afero.Fs) Initer {
	return initerImpl{
		fs: fs,
	}
}

type initerImpl struct {
	fs afero.Fs

	options InitOptions
}

func (i initerImpl) Do(options InitOptions) (err error) {
	i.options = options

	if options.URL == "" {
		return fmt.Errorf("appliance URL is required")
	}
	var serverURL *url.URL
	if serverURL, err = url.Parse(options.URL); err != nil {
		return
	}

	if options.Account == "" {
		return fmt.Errorf("account is required")
	}

	// Determine paths for rc and pem files
	var rcPath string
	if options.File != "" {
		rcPath = options.File
	} else {
		rcPath = path.Join(os.Getenv("HOME"), ".conjurrc")
	}

	dir := path.Dir(rcPath)
	var pemPath string
	if serverURL.Scheme == "https" {
		pemPath = path.Join(dir, fmt.Sprintf("conjur-%s.pem", options.Account))
	}

	// Write rc file
	var rcFile *os.File
	if rcFile, err = i.safeOpen(rcPath); err == nil {
		defer checkClose(rcFile)

		// ok if pemPath is zero-valued
		if err = i.writeConjurrc(rcFile, pemPath); err != nil {
			return
		}
	}

	// Write pem file, if appropriate
	if pemPath != "" {
		if pemFile, err := i.safeOpen(pemPath); err == nil {
			defer checkClose(pemFile)

			return i.writePem(pemFile)
		}
	}

	return
}

func (i initerImpl) safeOpen(path string) (*os.File, error) {
	// Ensure path doesn't exist (or --force is specified)
	if _, err := i.fs.Stat(path); err == nil && !i.options.Force {
		return nil, fmt.Errorf("%s exists, but force not specified", path)
	}

	flags := os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	file, err := i.fs.OpenFile(path, flags, 0644)
	return file.(*os.File), err
}

func checkClose(file *os.File) {
	if err := file.Close(); err != nil {
		logrus.Warn(err)
	}
}

func (i initerImpl) writeConjurrc(rc *os.File, pemPath string) (err error) {
	config := conjurapi.Config{
		Account:      i.options.Account,
		ApplianceURL: i.options.URL,
		SSLCertPath:  pemPath,
		V4:           false,
	}

	bytes, err := yaml.Marshal(config)
	if err != nil {
		return
	}

	buf := bufio.NewWriter(rc)
	if _, err = buf.Write(bytes); err != nil {
		return
	}
	return buf.Flush()
}

func (i initerImpl) writePem(pemFile *os.File) (err error) {
	certs, err := i.chooseCerts()
	if err != nil {
		return
	}

	buf := bufio.NewWriter(pemFile)
	for _, cert := range certs {
		block := &pem.Block{
			Type:    "CERTIFICATE",
			Headers: nil,
			Bytes:   cert.Raw,
		}
		if err = pem.Encode(buf, block); err != nil {
			return
		}
	}
	return buf.Flush()
}

func (i initerImpl) chooseCerts() (certs []*x509.Certificate, err error) {
	if i.options.Certificate != "" {
		// Use provided cert, ensure it can be parsed
		if block, rest := pem.Decode([]byte(i.options.Certificate)); string(rest) == "" {
			if certs, err = x509.ParseCertificates(block.Bytes); err != nil {
				return
			}
		} else {
			return nil, fmt.Errorf("extra characters at end of certificate: '%s'", rest)
		}
	} else {
		// Fetch certs from the server
		if serverURL, err := url.Parse(i.options.URL); err == nil {
			return fetchServerCerts(serverURL.Hostname(), serverURL.Port())
		}
	}

	return
}

func fetchServerCerts(hostname, port string) ([]*x509.Certificate, error) {
	conf := &tls.Config{
		// It's up to the user to verify the cert's fingerprint.
		InsecureSkipVerify: true,
	}

	if port == "" {
		port = "443"
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", hostname, port), conf)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	hexSHA := sha1.Sum(certs[0].Raw)

	logrus.Infof("Fetched certificates from https://%s:%s", hostname, port)

	// Match OpenSSL's formatting
	fingerprint := strings.Replace(fmt.Sprintf("% X", hexSHA), " ", ":", -1)
	fmt.Printf(`
SHA1 Fingerprint=%s
`, fingerprint)

	// Match the Ruby CLI's formatting
	fmt.Printf(`
Please verify this certificate on the appliance using command:
              openssl x509 -fingerprint -noout -in ~conjur/etc/ssl/conjur.pem

`,
	)

	return certs, nil
}
