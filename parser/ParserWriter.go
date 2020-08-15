package parser

import (
	"io"
)

type parserWriter struct{}

// Max []byte length = 512 including CRLF
func (parserWriter) Write(p []byte) (int, error) {
	for _, b := range p {
		HandleByte(b)
	}

	return len(p), nil
}

// TODO: is there a better way to handle this?
func MakeWriter() io.Writer {
	return parserWriter{}
}
