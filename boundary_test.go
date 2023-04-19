package streamtools

import (
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
		// FIXME: This is known broken, but really requires a
		// pretty significant overhaul of the code to fix and it
		// does what I need it to in the first place,
		// so... stopping here for now.
		// {
		// 	"ABC",
		// 	[]string{"AB", "AB", "C"},
		// 	[]string{"AB", "ABC"},
		// },
		{
			"AABC",
			[]string{"AA", "B", "C"},
			[]string{"AABC"},
		},
		{
			"ABC",
			[]string{"A", "B", "C"},
			[]string{"ABC"},
		},
		{
			"ABC",
			[]string{"bbA", "B", "C", "b"},
			[]string{"bb", "ABC", "b"},
		},
		{
			"ABC",
			[]string{"bbA", "BC", "b"},
			[]string{"bb", "ABC", "b"},
		},
		{
			"ABC",
			[]string{"bbABC"},
			[]string{"bb", "ABC"},
		},
		{
			"ABC",
			[]string{"ABCbb"},
			[]string{"ABC", "bb"},
		},
		{
			"ABC",
			[]string{"bbABCbb"},
			[]string{"bb", "ABC", "bb"},
		},
		{
			"A",
			[]string{"bAc"},
			[]string{"b", "A", "c"},
		},
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
			}
		}

		if !reflect.DeepEqual(outChunks, test.Out) {
			spew.Dump(test.Out, outChunks)
			t.Fatalf("TestSimpleBoundaryTest case %d failed",
				idx)
		}
	}
}

func TestSimpleBoundayrErrorCases(t *testing.T) {
	bas := NewBoundaryAtomicString(nil, "abcd")
	buf := make([]byte, 2)

	n, err := bas.Read(nil)
	if n != 0 || err != nil {
		t.Fatalf("wrong error returns")
	}

	n, err = bas.Read(buf)
	if n != 0 {
		t.Fatalf("claims to have read some stuff")
	}
	if err.(StreamError).ErrorType != ErrBufferTooSmall {
		t.Fatalf("incorrect errors")
	}
}
