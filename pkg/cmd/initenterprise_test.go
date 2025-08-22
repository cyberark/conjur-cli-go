package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/stretchr/testify/assert"
)

type mockInitClient struct {
	t               *testing.T
	jwtAuthenticate func(t *testing.T, client clients.ConjurClient) error
}

func (m mockInitClient) JWTAuthenticate(client clients.ConjurClient) error {
	return m.jwtAuthenticate(m.t, client)
}

var initEnterpriseCmdTestCases = []struct {
	name string
	// NOTE: -f defaults to conjurrcInTmpDir in the args slice.
	// It can always be overwritten by appending another value to the args slice.
	args            []string
	out             string
	promptResponses []promptResponse
	// Being unable to pipe responses to prompts is a known shortcoming of Survey.
	// https://github.com/go-survey/survey/issues/394
	// This flag is used to enable Pipe-based, and not PTY-based, tests.
	pipe            bool
	beforeTest      func(t *testing.T, conjurrcInTmpDir string) func()
	assert          func(t *testing.T, conjurrcInTmpDir string, stdout string)
	jwtAuthenticate func(t *testing.T, client clients.ConjurClient) error
}{
	{
		name: "help",
		args: []string{"init", "enterprise", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "help (self-hosted)",
		args: []string{"init", "self-hosted", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "writes conjurrc",
		args: []string{"init", "enterprise", "-u=http://host", "-a=test-account", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: http://host
environment: self-hosted
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
			// Shouldn't write certificate for HTTP url
			assert.NotContains(t, stdout, "Wrote certificate to")
		},
	},
	{
		name: "writes conjurrc for ldap",
		args: []string{"init", "enterprise", "-u=http://host", "-a=test-account", "-t=ldap", "--service-id=test", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: http://host
authn_type: ldap
service_id: test
environment: self-hosted
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "writes conjurrc for jwt",
		args: []string{"init", "enterprise", "-u=http://host", "-a=test-account", "-t=jwt", "--service-id=test", "--jwt-file=/path/to/jwt", "--jwt-host-id=host-id", "-i"},
		jwtAuthenticate: func(t *testing.T, client clients.ConjurClient) error {
			// Just return nil to simulate successful JWT authentication
			return nil
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: http://host
authn_type: jwt
service_id: test
jwt_host_id: host-id
jwt_file: /path/to/jwt
environment: self-hosted
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "fails when jwt authentication fails",
		args: []string{"init", "enterprise", "-u=http://host", "-a=test-account", "-t=jwt", "--service-id=test", "--jwt-file=/path/to/jwt", "--jwt-host-id=host-id", "-i"},
		jwtAuthenticate: func(t *testing.T, client clients.ConjurClient) error {
			return fmt.Errorf("jwt authentication failed")
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Unable to authenticate with Secrets Manager using the provided JWT file: jwt authentication failed")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "prompts for account and URL",
		args: []string{"init", "enterprise", "-i"},
		promptResponses: []promptResponse{
			{
				prompt:   "Enter the URL of your Secrets Manager service:",
				response: "http://conjur",
			},
			{
				prompt:   "Enter your organization account name:",
				response: "dev",
			},
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)

			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: dev
appliance_url: http://conjur
environment: self-hosted
`

			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
	{
		name: "prompts for overwrite, reject",
		args: []string{"init", "enterprise", "-u=http://host", "-a=other-test-account", "-i"},
		promptResponses: []promptResponse{
			{
				prompt:   ".conjurrc exists. Overwrite?",
				response: "N",
			},
		},
		pipe: true,
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
			return nil
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			// Assert that file is not overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			assert.Equal(t, "something", string(data))
		},
	},
	{
		name: "prompts for overwrite, accept",
		args: []string{"init", "enterprise", "-u=http://host", "-a=other-test-account", "-i"},
		promptResponses: []promptResponse{
			{
				prompt:   ".conjurrc exists. Overwrite?",
				response: "y",
			},
		},
		pipe: true,
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
			return nil
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			// Assert that file is overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: other-test-account
appliance_url: http://host
environment: self-hosted
`
			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
	{
		name: "writes conjurrc with force netrc",
		args: []string{"init", "enterprise", "-u=http://host", "-a=test-account", "--force-netrc", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: http://host
credential_storage: file
environment: self-hosted
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "force overwrite",
		args: []string{"init", "enterprise", "-u=http://host", "-a=yet-another-test-account", "--force", "-i"},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
			return nil
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			// Assert that file is overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: yet-another-test-account
appliance_url: http://host
environment: self-hosted
`
			assert.Equal(t, expectedConjurrc, string(data))

			// Assert on output
			assert.NotContains(t, stdout, ".conjurrc exists. Overwrite?")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "errors on missing conjurrc file directory",
		args: []string{"init", "enterprise", "-u=http://host", "-a=test-account", "-f=/no/such/dir/file", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "no such file or directory")
		},
	},
	{
		name: "prompts to trust self-signed certificate, reject",
		args: []string{"init", "enterprise", "-u=https://localhost:8080", "-a=test-account"},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			return startSelfSignedServer(t, 8080)
		},
		promptResponses: []promptResponse{
			{
				prompt:   "Trust this certificate?",
				response: "N",
			},
		},
		pipe: true,
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "You decided not to trust the certificate")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails if can't retrieve server certificate",
		args: []string{"init", "enterprise", "-u=https://nohost.example.com", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Unable to retrieve and validate certificate")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "succeeds for self-signed certificate with --self-signed flag",
		args: []string{"init", "enterprise", "-u=https://localhost:8080", "-a=test-account", "--self-signed"},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			return startSelfSignedServer(t, 8080)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Warning: Using self-signed certificates is not recommended and could lead to exposure of sensitive data")
			assertCertWritten(t, conjurrcInTmpDir, stdout)
		},
	},
	{
		name: "succeeds for self-signed certificate without --self-signed flag",
		args: []string{"init", "enterprise", "-u=https://localhost:8080", "-a=test-account"},
		promptResponses: []promptResponse{
			{
				prompt:   "Trust this certificate?",
				response: "y",
			},
		},
		pipe: true,
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			return startSelfSignedServer(t, 8080)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Trust this certificate? [y/N]")
			assertCertWritten(t, conjurrcInTmpDir, stdout)
		},
	},
	{
		name: "fails for http urls",
		args: []string{"init", "enterprise", "-u=http://example.com", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Cannot fetch certificate from non-HTTPS URL")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails urls without scheme",
		args: []string{"init", "enterprise", "-u=invalid-url", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Error: Cannot fetch certificate from non-HTTPS URL")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails for invalid urls",
		args: []string{"init", "enterprise", "-u=https://invalid:url:test", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Error: parse")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "allows http urls when specified",
		args: []string{"init", "enterprise", "-u=http://example.com", "-a=test-account", "--insecure"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Warning: Running the command with '--insecure' makes your system vulnerable to security attacks")
			assert.Contains(t, stdout, "If you prefer to communicate with the server securely you must reinitialize the client in secure mode.")
		},
	},
	{
		name: "fails if both --insecure and --self-signed are specified",
		args: []string{"init", "enterprise", "-u=http://example.com", "-a=test-account", "--insecure", "--self-signed"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Cannot specify both --insecure and --self-signed")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails if --insecure and --ca-cert are specified",
		args: []string{"init", "enterprise", "-u=http://example.com", "-a=test-account", "--insecure", "--ca-cert=cert.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Cannot specify --ca-cert when using --insecure or --self-signed")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails if --self-signed and --ca-cert are specified",
		args: []string{"init", "enterprise", "-u=http://example.com", "-a=test-account", "--self-signed", "--ca-cert=cert.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Cannot specify --ca-cert when using --insecure or --self-signed")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "allows cert specified by --ca-cert",
		args: []string{"init", "enterprise", "-u=https://example.com", "-a=test-account", "--ca-cert=custom-cert.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			pwd, _ := os.Getwd()
			expectedConjurrc := `account: test-account
appliance_url: https://example.com
cert_file: ` + pwd + `/custom-cert.pem
environment: self-hosted
`
			assert.Equal(t, expectedConjurrc, string(data))
		},
	}, {
		name: "works for open source",
		args: []string{"init", "open-source", "-u=http://host", "-a=test-account", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: http://host
environment: oss
`
			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
}

func TestInitEnterpriseCmd(t *testing.T) {
	for _, tc := range initEnterpriseCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// conjurrcInTmpDir is the .conjurrc test file location, it is
			// cleaned up before every test
			tempDir := t.TempDir()
			conjurrcInTmpDir := tempDir + "/.conjurrc"

			if tc.beforeTest != nil {
				cleanup := tc.beforeTest(t, conjurrcInTmpDir)
				if cleanup != nil {
					defer cleanup()
				}
			}

			// --file default to conjurrcInTmpDir. It can always be overwritten in each test case
			args := []string{
				"--file=" + tempDir + "/.conjurrc",
				"--cert-file=" + tempDir + "/conjur-server.pem",
			}
			args = append(args, tc.args...)

			// Create command tree for init
			mockClient := mockInitClient{t: t, jwtAuthenticate: tc.jwtAuthenticate}

			cmd := newInitEnterpriseCommand(initCmdFuncs{
				JWTAuthenticate: mockClient.JWTAuthenticate,
			})
			rootCmd := newRootCommand()
			initCmd := newInitCommand()
			initCmd.AddCommand(cmd)
			rootCmd.AddCommand(initCmd)
			rootCmd.SetArgs(args)

			var out string
			if tc.pipe {
				var content string
				for _, pr := range tc.promptResponses {
					content = fmt.Sprintf("%s%s\n", content, pr.response)
				}
				out, _ = executeCommandForTestWithPipeResponses(t, rootCmd, content)
			} else {
				out, _ = executeCommandForTestWithPromptResponses(t, rootCmd, tc.promptResponses)
			}

			if tc.assert != nil {
				tc.assert(t, conjurrcInTmpDir, out)
			}
		},
		)
	}

	// Other tests
	t.Run("default flags", func(t *testing.T) {

		cmd := newInitEnterpriseCommand(defaultInitCmdFuncs)
		rootCmd := newRootCommand()
		rootCmd.AddCommand(cmd)
		rootCmd.SetArgs([]string{"init"})
		executeCommandForTest(t, rootCmd)

		f, err := cmd.Flags().GetString("file")
		assert.NoError(t, err)
		homeDir, err := os.UserHomeDir()
		assert.NoError(t, err)
		assert.Equal(t, homeDir+"/.conjurrc", f)

		f, err = cmd.Flags().GetString("cert-file")
		assert.NoError(t, err)
		assert.Equal(t, homeDir+"/conjur-server.pem", f)
	})
}

func assertFetchCertFailed(t *testing.T, conjurrcInTmpDir string) {
	// Assert that conjurrc and certificate were not written
	expectedCertPath := filepath.Dir(conjurrcInTmpDir) + "/conjur-server.pem"
	_, err := os.Stat(conjurrcInTmpDir)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(expectedCertPath)
	assert.True(t, os.IsNotExist(err))
}

func assertCertWritten(t *testing.T, conjurrcInTmpDir string, stdout string) {
	expectedCertPath := filepath.Dir(conjurrcInTmpDir) + "/conjur-server.pem"

	// Assert that the certificate path is written to conjurrc
	data, _ := os.ReadFile(conjurrcInTmpDir)
	assert.Contains(t, string(data), "cert_file: "+expectedCertPath)

	// Assert that certificate is written
	assert.Contains(t, stdout, "Wrote certificate to "+expectedCertPath)
	data, _ = os.ReadFile(expectedCertPath)
	assert.Contains(t, string(data), "-----BEGIN CERTIFICATE-----")
}

func startSelfSignedServer(t *testing.T, port int) func() {
	err := generateSelfSignedCert()
	assert.NoError(t, err, "failed to generate self-signed certificate")
	// Load your custom cert and key
	cert, err := tls.LoadX509KeyPair("localhost.crt", "localhost.key")
	assert.NoError(t, err, "failed to load cert/key")

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	assert.NoError(t, err, "unable to start test server")

	// Set custom TLS config
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server.Listener = l
	server.StartTLS()

	return func() { server.Close() }
}

func generateSelfSignedCert() error {
	// Generate a private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	// Create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Localhost"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add localhost to the SAN
	template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	template.DNSNames = []string{"localhost"}

	// Create the certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// Save the certificate to a file
	certFile, err := os.Create("localhost.crt")
	if err != nil {
		return err
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return err
	}

	// Save the private key to a file
	keyFile, err := os.Create("localhost.key")
	if err != nil {
		return err
	}
	defer keyFile.Close()

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}

	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return err
	}

	return nil
}

func Test_env(t *testing.T) {
	tests := []struct {
		name string
		want conjurapi.EnvironmentType
	}{{
		name: "self-hosted",
		want: conjurapi.EnvironmentSH,
	}, {
		name: "CE",
		want: conjurapi.EnvironmentSH,
	}, {
		name: "default",
		want: conjurapi.EnvironmentSH,
	}, {
		name: "oss",
		want: conjurapi.EnvironmentOSS,
	}, {
		name: "open-source",
		want: conjurapi.EnvironmentOSS,
	}, {
		name: "OSS",
		want: conjurapi.EnvironmentOSS,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, env(tt.name), "env(%v)", tt.name)
		})
	}
}
