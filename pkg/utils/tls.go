package utils

import (
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

// ServerCert represents a TLS certificate and its fingerprint
type ServerCert struct {
	Fingerprint string
	Cert        string
}

// GetServerCert returns the TLS certificate and fingerprint for a given host.
// The host should be in the format hostname:port. If the port is not specified,
// 443 is used.
func GetServerCert(host string, allowSelfSigned bool) (ServerCert, error) {
	// Split host into hostname and port
	hostParts := strings.Split(host, ":")
	hostname := hostParts[0]
	port := "443"
	if len(hostParts) > 1 {
		port = hostParts[1]
	}

	conn, err := tls.Dial("tcp", hostname+":"+port, &tls.Config{
		InsecureSkipVerify: allowSelfSigned,
	})
	if err != nil {
		return ServerCert{}, err
	}
	defer conn.Close()

	// Get the server's certificate
	cert := conn.ConnectionState().PeerCertificates[0]

	if allowSelfSigned {
		// If allowing self-signed certificates, we need to verify the certificate manually because
		// InsecureSkipVerify is set to true which bypasses the automatic verification.
		// We load the system root CAs and add the server's certificate to the pool, then verify.
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		rootCAs.AddCert(cert)
		_, err = cert.Verify(x509.VerifyOptions{Roots: rootCAs})
		if err != nil {
			return ServerCert{}, err
		}
	}

	// Calculate the fingerprint of the cert
	fingerprint := getSha1Fingerprint(cert.Raw)
	pem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	return ServerCert{Fingerprint: fingerprint, Cert: string(pem)}, nil
}

func getSha1Fingerprint(cert []byte) string {
	sha1sum := sha1.Sum(cert)
	return strings.ToUpper(fmt.Sprintf("%x", sha1sum))
}
