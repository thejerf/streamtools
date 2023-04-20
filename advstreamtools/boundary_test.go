package advstreamtools

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
	// Readers can return bytes AND error at the same time, this lets
	// us configure that behavior.
	LastChunkEOFs bool
}

func (cr *ChunkReader) Read(b []byte) (int, error) {
	if len(cr.Chunks) == 0 {
		return 0, io.EOF
	}

	chunk := cr.Chunks[0]
	cr.Chunks = cr.Chunks[1:]

	if len(chunk) > len(b) {
		// this is a panic, rather than filling things in, because
		// the whole point here is to specify EXACTLY what comes
		// out of the reading process.
		panic(fmt.Sprintln("insufficiently-sized buffer:", len(chunk),
			"expected to fit into", len(b)))
	}

	n := copy(b, chunk)
	if cr.LastChunkEOFs && len(cr.Chunks) == 0 {
		return n, io.EOF
	} else {
		return n, nil
	}
}

func (cr *ChunkReader) Close() error {
	cr.Chunks = nil
	return nil
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
		// 	[]string{"01234567890123456789012345678901"},
		// 	[]string{"01234567890123456789012345678901"},
		// },
		{
			"ABC",
			[]string{"AB", "A", "", "", "B", "C"},
			[]string{"AB", "ABC"},
		},
		{
			"ABC",
			[]string{"AB", "A", "", "", "B", "CABCD"},
			[]string{"AB", "ABC", "ABC", "D"},
		},
		{
			"ABC",
			[]string{"AB", "A", "", "", "B", "CABC"},
			[]string{"AB", "ABC", "ABC"},
		},
		{
			"ABC",
			[]string{"AB", "A", "", "", "B", "CABC", ""},
			[]string{"AB", "ABC", "ABC"},
		},
		{
			"ABC",
			[]string{"AB", "AB", "C"},
			[]string{"AB", "ABC"},
		},
		{
			"password",
			[]string{"password=mumble&username=moo"},
			[]string{"password", "=mumble&username=moo"},
		},
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
			[]string{"bbbbbb"},
		},
	} {
		for _, lastChunkEOFs := range []bool{true, false} {
			cr := &ChunkReader{test.In, lastChunkEOFs}
			bas := NewBoundary[byte](cr, []byte(test.Search))

			outChunks := []string{}

			var err error
			var n int
			for err != io.EOF {
				buf := make([]byte, 32)
				n, err = bas.Read(buf)
				if err != nil && err != io.EOF {
					t.Fatalf("unexpected error: %v", err)
				}
				if err != io.EOF {
					outChunks = append(outChunks,
						string(buf[:n]))
				}
			}

			if !reflect.DeepEqual(outChunks, test.Out) {
				spew.Dump(test.Out, outChunks)
				t.Fatalf("TestSimpleBoundaryTest case %d/%v failed",
					idx, lastChunkEOFs)
			}
		}
	}
}

func TestSimpleBoundaryErrorCases(t *testing.T) {
	bas := NewBoundary[byte](nil, []byte("abcd"))
	n, err := bas.Read(nil)
	if n != 0 || err != nil {
		t.Fatalf("wrong error returns")
	}
	bas.Close() // coverage

	cr := &ChunkReader{}
	bas = NewBoundary[byte](cr, []byte("abcd"))
	bas.Close() // coverage
}
