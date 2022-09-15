package utils

import "io"

func NopReadCloser(r io.Reader) io.ReadCloser {
	return nopCloser{Reader: r}
}

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopCloser{Writer: w}
}

type nopCloser struct {
	io.Reader
	io.Writer
}

func (nopCloser) Close() error { return nil }
