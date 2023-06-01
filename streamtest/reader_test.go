package streamtest

import (
	"errors"
	"fmt"
	"testing"
	"io"
	"reflect"
)

func TestChunkReader(t *testing.T) {
	// it turns out the test has the exact same fields as the structure
	// itself
	for idx, test := range []ChunkReader{
		{
			Bytes: [][]byte{
				[]byte("a"),
				[]byte("abcd"),
				[]byte("e"),
			},
		},
		{
			Bytes: [][]byte{},
		},
		{
			Bytes: [][]byte{},
			TerminalErr: errors.New("hello"),
		},
		{
			Bytes: [][]byte{
				[]byte("a"),
			},
			TerminalErr: errors.New("oh no"),
		},
	} {
		cr := &ChunkReader{
			Bytes: append([][]byte{}, test.Bytes...),
			TerminalErr: test.TerminalErr,
		}

		obtained := [][]byte{}

		for {
			buf := make([]byte, 1024)
			n, err := cr.Read(buf)

			if err != nil {
				if test.TerminalErr == nil {
					if err != io.EOF {
						t.Fatalf("%d: unexpected final error: %v",
							idx, err)
					}
				} else {
					if test.TerminalErr != err {
						t.Fatalf("%d: unexpected final error: %v",
							idx, err)
					}
				}

				break
			} else {
				obtained = append(obtained, buf[:n])
			}
		}

		if !reflect.DeepEqual(obtained, test.Bytes) {
			fmt.Printf("%#v\n%#v\n", obtained, test.Bytes)
			t.Fatalf("%d: incorrect bytes obtained", idx)
		}
	}
}
