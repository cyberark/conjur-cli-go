package utils

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
)

const redactedString = "[REDACTED]"

type dumpTransport struct {
	roundTripper http.RoundTripper
	logRequest   func([]byte)
	logResponse  func([]byte)
}

// redactAuthz purges Authorization headers from a given request,
// and returns a function to restore them.
func redactAuthz(req *http.Request) (restore func()) {
	restore = func() {}

	origAuthz := req.Header.Get("Authorization")
	if origAuthz != "" {
		req.Header.Set("Authorization", redactedString)
		restore = func() {
			req.Header.Set("Authorization", origAuthz)
		}
	}

	return
}

// bodyMatch determines whether a given ReadCloser should be redacted
// given a RegEx, and returns a copy of the ReadCloser.
func bodyMatch(rc io.ReadCloser, rx *regexp.Regexp) (bool, io.ReadCloser, error) {
	if rc == nil || rc == http.NoBody {
		return true, http.NoBody, nil
	}

	var content bytes.Buffer
	if _, err := content.ReadFrom(rc); err != nil {
		return false, rc, err
	}
	if err := rc.Close(); err != nil {
		return false, rc, err
	}

	return rx.Match(content.Bytes()), io.NopCloser(&content), nil
}

// redactBody replaces a given ReadCloser with a redacted one if it
// is found to match the given RegEx, and returns a function to
// restore the ReadCloser.
func redactBody(rc *io.ReadCloser, contentLen *int64, rx *regexp.Regexp) (restore func()) {
	restore = func() {}

	redact, origBody, _ := bodyMatch(*rc, rx)

	if redact {
		origLength := *contentLen

		redactedReader := io.NopCloser(strings.NewReader(redactedString))
		*rc = redactedReader
		*contentLen = int64(len(redactedString))

		restore = func() {
			*rc = origBody
			*contentLen = origLength
		}
	} else {
		*rc = origBody
	}

	return
}

// dumpRequest logs the contents of a given HTTP request, but first:
// 1. sanitizes the Authorization header
// 2. sanitizes the request body if the request is for authentication
func (d *dumpTransport) dumpRequest(req *http.Request) []byte {
	restoreAuthz := redactAuthz(req)
	defer restoreAuthz()
	// This regex matches requests to create an issuer which include an AWS secret access key.
	rx := `"secret_access_key":\w?"[A-Za-z0-9\/+]{40}"`

	// We redact any request for authorization
	if strings.Contains(req.URL.Path, "/authn") {
		rx = ".*"
	}

	restoreBody := redactBody(
		&req.Body,
		&req.ContentLength,
		regexp.MustCompile(rx),
	)
	defer restoreBody()

	dump, _ := httputil.DumpRequestOut(req, true)
	return dump
}

// dumpResponse logs the contents of a given HTTP response, but first
// sanitizes the response body if it includes a Conjur token.
func (d *dumpTransport) dumpResponse(res *http.Response) []byte {
	restoreBody := redactBody(
		&res.Body,
		&res.ContentLength,
		regexp.MustCompile("{\"protected\":\".*\",\"payload\":\".*\",\"signature\":\".*\"}"),
	)
	defer restoreBody()

	dump, _ := httputil.DumpResponse(res, true)
	return dump
}

func (d *dumpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump := d.dumpRequest(req)
	if d.logRequest != nil {
		d.logRequest(dump)
	}

	res, err := d.roundTripper.RoundTrip(req)
	if err != nil {
		if d.logResponse != nil {
			d.logResponse([]byte(err.Error()))
		}
		return res, err
	}

	dump = d.dumpResponse(res)
	if d.logResponse != nil {
		d.logResponse(dump)
	}

	return res, err
}

// NewDumpTransport creates a RoundTripper that can log the dumps of requests and responses
func NewDumpTransport(roundTripper http.RoundTripper, logFunc func([]byte)) *dumpTransport {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}

	return &dumpTransport{
		roundTripper: roundTripper,
		logRequest:   logFunc,
		logResponse:  logFunc,
	}
}
