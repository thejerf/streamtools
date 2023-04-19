package streamtools

import (
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type SimpleBoundaryTest struct {
	Search string
	In     []string
	Out    []string
}

type ChunkReader struct {
	Chunks []string
}

func (cr *ChunkReader) Read(b []byte) (int, error) {
	if len(cr.Chunks) == 0 {
		return 0, io.EOF
	}

	chunk := cr.Chunks[0]
	cr.Chunks = cr.Chunks[1:]

	if len(chunk) > len(b) {
		panic("insufficiently-sized buffer")
	}

	n := copy(b, chunk)
	return n, nil
}

func TestSimpleBoundaryTest(t *testing.T) {
	for idx, test := range []SimpleBoundaryTest{
		// FIXME: Handle a null search string correctly
		{
			"a",
			[]string{"b", "bb", "bbb"},
			[]string{"b", "bb", "bbb"},
		},
	} {
		cr := &ChunkReader{test.In}
		bas := NewBoundaryAtomicString(cr, test.Search)

		outChunks := []string{}

		var err error
		var n int
		for err != io.EOF {
			buf := make([]byte, 1024)
			n, err = bas.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatalf("unexpected error: %v", err)
			}
			if err != io.EOF {
				outChunks = append(outChunks, string(buf[:n]))
				fmt.Println("Got out chunk:",
					string(buf[:n]))
			}
		}

		if !reflect.DeepEqual(outChunks, test.Out) {
			spew.Dump(test.Out, outChunks)
			t.Fatalf("TestSimpleBoundaryTest case %d failed",
				idx)
		}
	}
}
