package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetServerCert(t *testing.T) {
	// Note: this will change every few years when GitHub renews their TLS certificate
	githubFingerprint := "A3:B5:9E:5F:E8:84:EE:1F:34:D9:8E:EF:85:8E:3F:B6:62:AC:10:4A"
	githubCert := `-----BEGIN CERTIFICATE-----
MIIFajCCBPGgAwIBAgIQDNCovsYyz+ZF7KCpsIT7HDAKBggqhkjOPQQDAzBWMQsw
CQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMTAwLgYDVQQDEydEaWdp
Q2VydCBUTFMgSHlicmlkIEVDQyBTSEEzODQgMjAyMCBDQTEwHhcNMjMwMjE0MDAw
MDAwWhcNMjQwMzE0MjM1OTU5WjBmMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2Fs
aWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEVMBMGA1UEChMMR2l0SHVi
LCBJbmMuMRMwEQYDVQQDEwpnaXRodWIuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0D
AQcDQgAEo6QDRgPfRlFWy8k5qyLN52xZlnqToPu5QByQMog2xgl2nFD1Vfd2Xmgg
nO4i7YMMFTAQQUReMqyQodWq8uVDs6OCA48wggOLMB8GA1UdIwQYMBaAFAq8CCkX
jKU5bXoOzjPHLrPt+8N6MB0GA1UdDgQWBBTHByd4hfKdM8lMXlZ9XNaOcmfr3jAl
BgNVHREEHjAcggpnaXRodWIuY29tgg53d3cuZ2l0aHViLmNvbTAOBgNVHQ8BAf8E
BAMCB4AwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIGbBgNVHR8EgZMw
gZAwRqBEoEKGQGh0dHA6Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5
YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5jcmwwRqBEoEKGQGh0dHA6Ly9jcmw0LmRp
Z2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5j
cmwwPgYDVR0gBDcwNTAzBgZngQwBAgIwKTAnBggrBgEFBQcCARYbaHR0cDovL3d3
dy5kaWdpY2VydC5jb20vQ1BTMIGFBggrBgEFBQcBAQR5MHcwJAYIKwYBBQUHMAGG
GGh0dHA6Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBPBggrBgEFBQcwAoZDaHR0cDovL2Nh
Y2VydHMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0VExTSHlicmlkRUNDU0hBMzg0MjAy
MENBMS0xLmNydDAJBgNVHRMEAjAAMIIBgAYKKwYBBAHWeQIEAgSCAXAEggFsAWoA
dwDuzdBk1dsazsVct520zROiModGfLzs3sNRSFlGcR+1mwAAAYZQ3Rv6AAAEAwBI
MEYCIQDkFq7T4iy6gp+pefJLxpRS7U3gh8xQymmxtI8FdzqU6wIhALWfw/nLD63Q
YPIwG3EFchINvWUfB6mcU0t2lRIEpr8uAHYASLDja9qmRzQP5WoC+p0w6xxSActW
3SyB2bu/qznYhHMAAAGGUN0cKwAABAMARzBFAiAePGAyfiBR9dbhr31N9ZfESC5G
V2uGBTcyTyUENrH3twIhAPwJfsB8A4MmNr2nW+sdE1n2YiCObW+3DTHr2/UR7lvU
AHcAO1N3dT4tuYBOizBbBv5AO2fYT8P0x70ADS1yb+H61BcAAAGGUN0cOgAABAMA
SDBGAiEAzOBr9OZ0+6OSZyFTiywN64PysN0FLeLRyL5jmEsYrDYCIQDu0jtgWiMI
KU6CM0dKcqUWLkaFE23c2iWAhYAHqrFRRzAKBggqhkjOPQQDAwNnADBkAjAE3A3U
3jSZCpwfqOHBdlxi9ASgKTU+wg0qw3FqtfQ31OwLYFdxh0MlNk/HwkjRSWgCMFbQ
vMkXEPvNvv4t30K6xtpG26qmZ+6OiISBIIXMljWnsiYR1gyZnTzIg3AQSw4Vmw==
-----END CERTIFICATE-----`

	// Note: this will change whenever certs are renewed for self-signed.badssl.com
	selfSignedFingerprint := "42:B0:D7:0D:41:C3:7C:E7:09:9F:55:97:56:BC:51:E5:D0:34:24:51"

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
