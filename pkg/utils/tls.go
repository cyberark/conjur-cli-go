package utils

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
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

func GetServerCertViaHTTPProxy(ctx context.Context, proxyRaw, host string, timeout time.Duration) (ServerCert, error) {
	hostname, port := splitHostPortDefault(host, "443")
	targetAddr := net.JoinHostPort(hostname, port)

	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	proxyURL, err := url.Parse(proxyRaw)
	if err != nil {
		return ServerCert{}, fmt.Errorf("parse proxy url: %w", err)
	}
	if proxyURL.Scheme != "http" {
		return ServerCert{}, fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
	}
	if proxyURL.Host == "" {
		return ServerCert{}, errors.New("proxy host is empty")
	}

	d := &net.Dialer{Timeout: timeout}
	raw, err := d.DialContext(ctx, "tcp", proxyURL.Host)
	if err != nil {
		return ServerCert{}, fmt.Errorf("dial proxy: %w", err)
	}
	defer raw.Close()

	_ = raw.SetDeadline(time.Now().Add(timeout))

	req := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: targetAddr},
		Host:   targetAddr,
		Header: make(http.Header),
	}
	if proxyURL.User != nil {
		user := proxyURL.User.Username()
		pass, _ := proxyURL.User.Password()
		token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
		req.Header.Set("Proxy-Authorization", "Basic "+token)
	}
	req.Header.Set("Proxy-Connection", "keep-alive")

	if err := req.Write(raw); err != nil {
		return ServerCert{}, fmt.Errorf("write CONNECT: %w", err)
	}

	br := bufio.NewReader(raw)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		return ServerCert{}, fmt.Errorf("read CONNECT response: %w", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ServerCert{}, fmt.Errorf("proxy CONNECT failed: %s", resp.Status)
	}

	tlsConn := tls.Client(raw, &tls.Config{
		ServerName:         hostname, // SNI + hostname verification
		InsecureSkipVerify: true,     // we will do verification ourselves below
		MinVersion:         tls.VersionTLS12,
	})

	_ = tlsConn.SetDeadline(time.Now().Add(timeout))
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return ServerCert{}, fmt.Errorf("tls handshake: %w", err)
	}
	_ = tlsConn.SetDeadline(time.Time{})

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return ServerCert{}, errors.New("no peer certificates found")
	}
	leaf := state.PeerCertificates[0]

	// Ensure the cert matches the hostname even though we skipped verify in tls.Config.
	if err := leaf.VerifyHostname(hostname); err != nil {
		return ServerCert{}, fmt.Errorf("certificate does not match hostname: %w", err)
	}

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	selfSigned := isSelfSigned(leaf)
	if selfSigned {
		rootCAs.AddCert(leaf)
	}

	opts := x509.VerifyOptions{
		Roots:         rootCAs,
		DNSName:       hostname,
		Intermediates: x509.NewCertPool(),
	}
	for _, c := range state.PeerCertificates[1:] {
		opts.Intermediates.AddCert(c)
	}

	untrustedCA := false
	if _, err := leaf.Verify(opts); err != nil {
		var unknownAuthorityError x509.UnknownAuthorityError
		if errors.As(err, &unknownAuthorityError) {
			untrustedCA = true
		} else {
			return ServerCert{}, fmt.Errorf("verify certificate: %w", err)
		}
	}

	fp := getSha256Fingerprint(leaf.Raw)
	p := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leaf.Raw})

	return ServerCert{
		Fingerprint:    fp,
		Issuer:         leaf.Issuer,
		CreationDate:   leaf.NotBefore.Format(time.RFC1123),
		ExpirationDate: leaf.NotAfter.Format(time.RFC1123),
		Cert:           string(p),
		SelfSigned:     selfSigned,
		UntrustedCA:    untrustedCA,
	}, nil
}

// splitHostPortDefault returns host and port from an input string.
// If no port is present, defaultPort is returned.
// It also tolerates URL inputs (e.g. https://example.com:443/path) by extracting the host.
func splitHostPortDefault(host, defaultPort string) (string, string) {
	s := strings.TrimSpace(host)
	if s == "" {
		return "", defaultPort
	}

	// If a URL is provided, extract its host.
	if u, err := url.Parse(s); err == nil && u.Host != "" {
		s = strings.TrimSpace(u.Host)
	}

	// Try strict host:port parsing (handles IPv6 in brackets).
	if h, p, err := net.SplitHostPort(s); err == nil {
		if p == "" {
			return h, defaultPort
		}
		return h, p
	}

	// If there is a colon, but it isn't a valid host:port (often raw IPv6 without brackets),
	// keep it as host and apply the default port.
	if strings.Contains(s, ":") && !strings.HasPrefix(s, "[") {
		return s, defaultPort
	}

	return s, defaultPort
}
