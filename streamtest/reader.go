package streamtest

import (
	"fmt"
	"io"
)


// ChunkReader will return a series of []bytes as the result of a .Read
// call, allowing precise specification of how the Reads will operate.
//
// ChunkReader assumes that all .Read calls will provide a buffer large
// enough for the largest value in Bytes, and will panic otherwise. That
// is, this makes no attempt at providing partial reads of the Bytes.
//
// If TerminalErr is set, it will be returned once all Bytes have been
// returned. Otherwise io.EOF will be returned.
type ChunkReader struct {
	Bytes [][]byte
	TerminalErr error
}

// NewChunkReader returns a ChunkReader populated with the given strings
// used as the []byte chunks. This is provided for ease-of-use in test
// code.
func NewChunkReader(strings ...string) *ChunkReader {
	b := make([][]byte, len(strings))
	for idx, s := range strings {
		b[idx] = []byte(s)
	}
	return &ChunkReader{Bytes: b}
}

// Read will hand out the given chunks until none are left, then return the
// TerminalError, or io.EOF if no such error is set.
func (cr *ChunkReader) Read(b []byte) (int, error) {
	if len(cr.Bytes) > 0 {
		n := copy(b, cr.Bytes[0])
		if n < len(cr.Bytes[0]) {
			panic(fmt.Sprintf("insufficient buffer for element of length %d",
				len(cr.Bytes[0])))
		}
		cr.Bytes = cr.Bytes[1:]
		return n, nil
	}

	if cr.TerminalErr != nil {
		return 0, cr.TerminalErr
	}

	return 0, io.EOF
}
