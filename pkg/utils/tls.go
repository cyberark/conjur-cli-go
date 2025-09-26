package utils

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ServerCert represents a TLS certificate and its fingerprint
type ServerCert struct {
	Fingerprint    string
	Issuer         pkix.Name
	CreationDate   string
	ExpirationDate string
	Cert           string
	SelfSigned     bool
	UntrustedCA    bool
}

func isSelfSigned(cert *x509.Certificate) bool {
	return cert.Subject.CommonName == cert.Issuer.CommonName
}

// GetServerCert returns the TLS certificate and fingerprint for a given host.
// The host should be in the format hostname:port. If the port is not specified,
// 443 is used.
func GetServerCert(host string) (ServerCert, error) {
	// Split host into hostname and port
	hostParts := strings.Split(host, ":")
	hostname := hostParts[0]
	port := "443"
	if len(hostParts) > 1 {
		port = hostParts[1]
	}

	conn, err := tls.Dial("tcp", hostname+":"+port, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return ServerCert{}, err
	}
	defer conn.Close()

	// Get the server's certificate
	peerCerts := conn.ConnectionState().PeerCertificates
	if len(peerCerts) == 0 {
		return ServerCert{}, errors.New("no peer certificates found")
	}
	cert := peerCerts[0]
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	selfSigned := isSelfSigned(cert)

	if selfSigned {
		rootCAs.AddCert(cert)
	}

	opts := x509.VerifyOptions{
		Roots:         rootCAs,
		DNSName:       hostname,
		Intermediates: x509.NewCertPool(),
	}

	for _, cert := range peerCerts[1:] {
		opts.Intermediates.AddCert(cert)
	}

	var untrustedCA bool

	_, err = cert.Verify(opts)
	if err != nil {
		fmt.Println("Error verifying certificate:", err)
		// check if error is due to unknowh authority
		var unknownAuthorityError x509.UnknownAuthorityError
		if errors.As(err, &unknownAuthorityError) {
			untrustedCA = true
		} else {
			return ServerCert{}, err
		}
	}

	// Calculate the fingerprint of the cert
	fingerprint := getSha256Fingerprint(cert.Raw)
	pem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	creationDate := cert.NotBefore.Format(time.RFC1123)
	expirationDate := cert.NotAfter.Format(time.RFC1123)
	return ServerCert{Fingerprint: fingerprint, Issuer: cert.Issuer, CreationDate: creationDate, ExpirationDate: expirationDate, Cert: string(pem), SelfSigned: selfSigned, UntrustedCA: untrustedCA}, nil
}

func getSha256Fingerprint(cert []byte) string {
	sum := sha256.Sum256(cert)
	return strings.ToUpper(fmt.Sprintf("%x", sum))
}
