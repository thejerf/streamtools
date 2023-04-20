package streamtools

import (
	"io"

	_ "github.com/davecgh/go-spew/spew"
	"github.com/thejerf/streamtools/advstreamtools"
)

// NewBoundaryAtomicString returns a reader that will do its best to return
// the search string atomically in a Read call.
//
// An example use for this is filtering an HTTP POST body or response for
// passwords. Create a NewBoundaryString with a search string of
// "password", and it becomes safe to .Read from the reader and simply
// check to see if the .Read value is == "password".
//
// If the buffer is not large enough to contain the search string, it will
// still be chunked, but of course the search string will not be in one
// .Read call.
func NewBoundaryString(src io.Reader, search string) io.Reader {
	return advstreamtools.NewBoundary[byte](src, []byte(search))
}

// NewBoundaryStringCloser is the same as NewBoundaryString, except it
// returns something that can also be closed.
func NewBoundaryStringCloser(src io.ReadCloser, search string) io.ReadCloser {
	return advstreamtools.NewBoundary[byte](src, []byte(search))
}
