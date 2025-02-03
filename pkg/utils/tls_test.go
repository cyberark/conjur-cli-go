package utils

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetServerCert(t *testing.T) {
	t.Run("Returns an error when the server doesn't exist", func(t *testing.T) {
		_, err := GetServerCert("localhost:444", false)
		assert.Error(t, err)
	})

	t.Run("Returns an error for self-signed certificates", func(t *testing.T) {
		server := startSelfSignedServer(t, 8080)
		defer server.Close()

		_, err := GetServerCert("localhost:8080", false)
		assert.Error(t, err)
	})

	t.Run("Returns the right certificate for self-signed certificates when allowed", func(t *testing.T) {
		server := startSelfSignedServer(t, 8080)
		defer server.Close()

		selfSignedFingerprint := getSha256Fingerprint(server.Certificate().Raw)

		cert, err := GetServerCert("localhost:8080", true)
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

func startSelfSignedServer(t *testing.T, port int) *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		assert.NoError(t, err, "unabled to start test server")
	}

	server.Listener = l
	server.StartTLS()

	return server
}
