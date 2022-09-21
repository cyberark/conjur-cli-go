package utils

import (
	"net/http"
	"net/http/httputil"
)

type dumpTransport struct {
	roundTripper http.RoundTripper
	logRequest func([]byte)
    logResponse func([]byte)
}

func (d *dumpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(req, true)
	if d.logRequest != nil {
		d.logRequest(dump)
	}
	res, err := d.roundTripper.RoundTrip(req)
	dump, _ = httputil.DumpResponse(res, true)
	if d.logResponse != nil {
		d.logResponse(dump)
	}
	return res, err
}

// TODO?: maybe add newlines to the dumps here to seperate logs

// NewDumpTransport creates a RoundTripper that can log the dumps of requests and responses
func NewDumpTransport(roundTripper http.RoundTripper, logFunc func([]byte)) *dumpTransport {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}

	return &dumpTransport{
		roundTripper: roundTripper,
		logRequest: logFunc,
		logResponse: logFunc,
	}
}
