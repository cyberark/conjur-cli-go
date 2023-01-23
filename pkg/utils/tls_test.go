package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetServerCert(t *testing.T) {
	// Note: this will change every few years when GitHub renews their TLS certificate
	githubFingerprint := "1E:16:CC:3F:84:2F:65:FC:C0:AB:93:2D:63:8A:C6:4A:95:C9:1B:7A"
	githubCert := `-----BEGIN CERTIFICATE-----
MIIFajCCBPCgAwIBAgIQBRiaVOvox+kD4KsNklVF3jAKBggqhkjOPQQDAzBWMQsw
CQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMTAwLgYDVQQDEydEaWdp
Q2VydCBUTFMgSHlicmlkIEVDQyBTSEEzODQgMjAyMCBDQTEwHhcNMjIwMzE1MDAw
MDAwWhcNMjMwMzE1MjM1OTU5WjBmMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2Fs
aWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEVMBMGA1UEChMMR2l0SHVi
LCBJbmMuMRMwEQYDVQQDEwpnaXRodWIuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0D
AQcDQgAESrCTcYUh7GI/y3TARsjnANwnSjJLitVRgwgRI1JlxZ1kdZQQn5ltP3v7
KTtYuDdUeEu3PRx3fpDdu2cjMlyA0aOCA44wggOKMB8GA1UdIwQYMBaAFAq8CCkX
jKU5bXoOzjPHLrPt+8N6MB0GA1UdDgQWBBR4qnLGcWloFLVZsZ6LbitAh0I7HjAl
BgNVHREEHjAcggpnaXRodWIuY29tgg53d3cuZ2l0aHViLmNvbTAOBgNVHQ8BAf8E
BAMCB4AwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIGbBgNVHR8EgZMw
gZAwRqBEoEKGQGh0dHA6Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5
YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5jcmwwRqBEoEKGQGh0dHA6Ly9jcmw0LmRp
Z2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5j
cmwwPgYDVR0gBDcwNTAzBgZngQwBAgIwKTAnBggrBgEFBQcCARYbaHR0cDovL3d3
dy5kaWdpY2VydC5jb20vQ1BTMIGFBggrBgEFBQcBAQR5MHcwJAYIKwYBBQUHMAGG
GGh0dHA6Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBPBggrBgEFBQcwAoZDaHR0cDovL2Nh
Y2VydHMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0VExTSHlicmlkRUNDU0hBMzg0MjAy
MENBMS0xLmNydDAJBgNVHRMEAjAAMIIBfwYKKwYBBAHWeQIEAgSCAW8EggFrAWkA
dgCt9776fP8QyIudPZwePhhqtGcpXc+xDCTKhYY069yCigAAAX+Oi8SRAAAEAwBH
MEUCIAR9cNnvYkZeKs9JElpeXwztYB2yLhtc8bB0rY2ke98nAiEAjiML8HZ7aeVE
P/DkUltwIS4c73VVrG9JguoRrII7gWMAdwA1zxkbv7FsV78PrUxtQsu7ticgJlHq
P+Eq76gDwzvWTAAAAX+Oi8R7AAAEAwBIMEYCIQDNckqvBhup7GpANMf0WPueytL8
u/PBaIAObzNZeNMpOgIhAMjfEtE6AJ2fTjYCFh/BNVKk1mkTwBTavJlGmWomQyaB
AHYAs3N3B+GEUPhjhtYFqdwRCUp5LbFnDAuH3PADDnk2pZoAAAF/jovErAAABAMA
RzBFAiEA9Uj5Ed/XjQpj/MxQRQjzG0UFQLmgWlc73nnt3CJ7vskCICqHfBKlDz7R
EHdV5Vk8bLMBW1Q6S7Ga2SbFuoVXs6zFMAoGCCqGSM49BAMDA2gAMGUCMCiVhqft
7L/stBmv1XqSRNfE/jG/AqKIbmjGTocNbuQ7kt1Cs7kRg+b3b3C9Ipu5FQIxAM7c
tGKrYDGt0pH8iF6rzbp9Q4HQXMZXkNxg+brjWxnaOVGTDNwNH7048+s/hT9bUQ==
-----END CERTIFICATE-----`

	selfSignedFingerprint := "31:CF:9F:34:65:7B:F1:A7:A2:AA:B0:4F:46:48:19:42:83:6D:84:E2"

	t.Run("Returns the right certificate from github.com", func(t *testing.T) {
		cert, err := GetServerCert("github.com", false)
		assert.NoError(t, err)

		assert.Equal(t, githubCert, strings.TrimSpace(cert.Cert))
		assert.Equal(t, githubFingerprint, cert.Fingerprint)
	})

	t.Run("Returns the right certificate from github.com:443", func(t *testing.T) {
		cert, err := GetServerCert("github.com:443", false)
		assert.NoError(t, err)

		assert.Equal(t, githubCert, strings.TrimSpace(cert.Cert))
		assert.Equal(t, githubFingerprint, cert.Fingerprint)
	})

	t.Run("Returns the right certificate from github.com when self-signed is allowed", func(t *testing.T) {
		// Ensure that allowing self-signed certs doesn't break support for normal certs
		cert, err := GetServerCert("github.com", true)
		assert.NoError(t, err)

		assert.Equal(t, githubCert, strings.TrimSpace(cert.Cert))
		assert.Equal(t, githubFingerprint, cert.Fingerprint)
	})

	t.Run("Returns an error when the server doesn't exist", func(t *testing.T) {
		_, err := GetServerCert("example.com:444", false)
		assert.Error(t, err)
	})

	t.Run("Returns an error for self-signed certificates", func(t *testing.T) {
		_, err := GetServerCert("self-signed.badssl.com", false)
		assert.Error(t, err)
	})

	t.Run("Returns the right certificate for self-signed certificates when allowed", func(t *testing.T) {
		cert, err := GetServerCert("self-signed.badssl.com", true)
		assert.NoError(t, err)

		assert.Equal(t, selfSignedFingerprint, cert.Fingerprint)
	})

	t.Run("Fails for incorrect hostname even when self-signed is allowed", func(t *testing.T) {
		_, err := GetServerCert("https://wrong.host.badssl.com/", true)
		assert.Error(t, err)
	})

	t.Run("Fails for untrusted root even when self-signed is allowed", func(t *testing.T) {
		_, err := GetServerCert("https://untrusted-root.badssl.com/", true)
		assert.Error(t, err)
	})
}
