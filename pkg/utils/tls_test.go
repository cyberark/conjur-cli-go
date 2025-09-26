package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetServerCert(t *testing.T) {
	t.Run("Returns an error when the server doesn't exist", func(t *testing.T) {
		_, err := GetServerCert("localhost:444")
		assert.Error(t, err)
	})

	t.Run("Returns an object with SelfSigned field set to true for self-signed certificates", func(t *testing.T) {
		server := startSelfSignedServer(t, 8081)
		defer server.Close()

		res, err := GetServerCert("localhost:8081")
		assert.NoError(t, err)
		assert.Equal(t, true, res.SelfSigned)
	})

	t.Run("Returns the right certificate for self-signed certificates when allowed", func(t *testing.T) {
		server := startSelfSignedServer(t, 8081)
		defer server.Close()

		selfSignedFingerprint := getSha256Fingerprint(server.Certificate().Raw)

		cert, err := GetServerCert("localhost:8081")
		assert.NoError(t, err)

		assert.Equal(t, selfSignedFingerprint, cert.Fingerprint)
	})

	t.Run("Fails for incorrect hostname even when self-signed is allowed", func(t *testing.T) {
		_, err := GetServerCert("https://wrong.host.badssl.com/")
		assert.Error(t, err)
	})

	t.Run("Fails for untrusted root even when self-signed is allowed", func(t *testing.T) {
		_, err := GetServerCert("https://untrusted-root.badssl.com/")
		assert.Error(t, err)
	})
}

func startSelfSignedServer(t *testing.T, port int) *httptest.Server {
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

	return server
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
