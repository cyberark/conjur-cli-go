package utils

import "io"

// NopReadCloser returns a ReadCloser with a no-op Close method.
func NopReadCloser(r io.Reader) io.ReadCloser {
	return nopCloser{Reader: r}
}

// NopWriteCloser returns a WriteCloser with a no-op Close method.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopCloser{Writer: w}
}

type nopCloser struct {
	io.Reader
	io.Writer
}

func (nopCloser) Close() error { return nil }
