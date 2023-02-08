package utils

import "io"

type byteBufferedReader struct {
	io.Reader
}

func (b *byteBufferedReader) Read(dst []byte) (int, error) {
	buf := make([]byte, 1)
	n, err := b.Reader.Read(buf)
	if err != nil {
		return n, err
	}

	copy(dst, buf)
	return n, nil
}

// NewByteBufferedReader constructs a byteBufferedReader
//
// byteBufferedReader addresses the issue that promptui internally uses a 4096 bytes buffer to read the contents of
// the stdin reader. That internal buffer is unfortunately not shared across prompts using the same stdin reader.
// This results in an approach that is greedy and for example throws away the contents from stdin between
// sequential prompts when the newline is found anywhere other than at the end of the buffer.
//
// byteBufferedReader wraps the stdin reader so that it forces a 1 byte buffer by only ever returning a maximum
// of 1 byte per call to Read.
//
// Note that a much more cumbersome solution than the one presented here might be
// to pad out the 4096 bytes buffer after the newline.
func NewByteBufferedReader(r io.Reader) io.Reader {
	return &byteBufferedReader{
		Reader: r,
	}
}
