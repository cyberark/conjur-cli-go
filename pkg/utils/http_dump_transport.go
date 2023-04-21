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

// redactBody determines whether a given request or response body should
// be redacted, and returns a copy of the body to reattach to the
// request or response in question.
func redactBody(rc io.ReadCloser, rx *regexp.Regexp) (bool, io.ReadCloser, error) {
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

func redactRequestBody(req *http.Request) (restore func()) {
	restore = func() {}

	redactedReader := io.NopCloser(strings.NewReader(redactedString))

	redact, origBody, _ := redactBody(req.Body, regexp.MustCompile(".*"))

	if redact {
		origLength := req.ContentLength

		req.Body = redactedReader
		req.ContentLength = int64(len(redactedString))

		restore = func() {
			req.Body = origBody
			req.ContentLength = origLength
		}
	} else {
		req.Body = origBody
	}

	return
}

func redactResponseBody(res *http.Response) (restore func()) {
	restore = func() {}

	redactedReader := io.NopCloser(strings.NewReader(redactedString))

	redact, origBody, _ := redactBody(res.Body, regexp.MustCompile("{\"protected\":\".*\",\"payload\":\".*\",\"signature\":\".*\"}"))

	if redact {
		origLength := res.ContentLength

		res.Body = redactedReader
		res.ContentLength = int64(len(redactedString))

		restore = func() {
			res.Body = origBody
			res.ContentLength = origLength
		}
	} else {
		res.Body = origBody
	}

	return
}

// dumpRequest logs the contents of a given HTTP request, but first:
// 1. sanitizes the Authorization header
// 2. sanitizes the request body if the request is for authentication
func (d *dumpTransport) dumpRequest(req *http.Request) []byte {
	restoreAuthz := redactAuthz(req)
	defer restoreAuthz()

	if strings.Contains(req.URL.Path, "/authn") {
		restoreBody := redactRequestBody(req)
		defer restoreBody()
	}

	dump, _ := httputil.DumpRequestOut(req, true)
	return dump
}

// dumpResponse logs the contents of a given HTTP response, but first
// sanitizes the response body if it includes a Conjur token.
func (d *dumpTransport) dumpResponse(res *http.Response) []byte {
	restoreBody := redactResponseBody(res)
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
