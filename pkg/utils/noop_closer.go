package utils

import "io"

// NoopReadCloser returns a ReadCloser with a no-op Close method.
func NoopReadCloser(r io.Reader) io.ReadCloser {
	return noopCloser{Reader: r}
}

// NoopWriteCloser returns a WriteCloser with a no-op Close method.
func NoopWriteCloser(w io.Writer) io.WriteCloser {
	return noopCloser{Writer: w}
}

type noopCloser struct {
	io.Reader
	io.Writer
}

func (noopCloser) Close() error { return nil }
