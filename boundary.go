package streamtools

import (
	"fmt"
	"io"
)

// The "boundary" prefix refers to stream manipulators that manipulate the
// otherwise-random boundary of the read calls, but do not manipulate the
// contents.

// and if found, is guaranteed to produce that chunk as its own .Read result.
//
// Example: Suppose your stream says
//
//	<p class="welcome">Hello!</p>
//
// and you set the search to []byte("class"). The resulting reader will
// produce the following three chunks:
//
//	<p
//	class
//	="welcome">Hello!</p>
//
// Regardless of whether or not the "class" was on a chunk boundary, as
// long as the reader byte buffer is large enough to contain it. If it is
// not, it will still put out as much of the matching string as the start
// as possible, but of course it will have to break it up.
//
// Memory usage: This has to accumulate a minimum of the string you passed
// in. For small strings dominated by the read buffer this is
// neglibible. If you search for a multi-megabyte string, this will result
// in a multi-megabyte buffer and a high likelihood of having a too-small
// buffer.
type BoundaryAtomicString struct {
	src          io.Reader
	searchString []byte

	// current buff
	currBuf []byte

	nonMatchAccum []byte
	matchAccum    []byte
	// the last error we got from a Read call. Per the io.Reader
	// documentation, we need to process the bytes in our buffer before
	// we return this error.
	lastReaderErr error
}

// NewBoundaryAtomicString constructs a new BoundaryAtomicString value correctly.
func NewBoundaryAtomicString(src io.Reader, search string) *BoundaryAtomicString {
	return &BoundaryAtomicString{
		src:          src,
		searchString: []byte(search),
	}
}

// FIXME: This should be abstractable into a generalized function that
// takes the next value and returns either:
//  1. no, I'm not interested in this value
//  2. yes, I'm interested but not done yet
//  3. this completes my partition, please return it

func (bas *BoundaryAtomicString) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	if len(buf) < len(bas.searchString) {
		return 0, errorf(
			ErrBufferTooSmall,
			"BoundaryAtomicString: buffer sized %d insufficient to hold %d-sized search string",
			len(buf),
			len(bas.searchString),
		)
	}

	for {
		// If we're sitting on a match, return it; next call will continue on.
		if len(bas.matchAccum) == len(bas.searchString) {
			copy(buf, bas.matchAccum)
			bas.matchAccum = bas.matchAccum[:0]
			return len(bas.searchString), nil
		}

		// if we're sitting on an error, return it now
		if len(bas.matchAccum) == 0 &&
			len(bas.nonMatchAccum) == 0 &&
			len(bas.currBuf) == 0 &&
			bas.lastReaderErr != nil {
			return 0, bas.lastReaderErr
		}

		// if we're sitting on a non-match and it is already equal to or
		// larger than the buffer, return it
		if len(bas.nonMatchAccum) >= len(buf) {
			n := copy(buf, bas.nonMatchAccum)
			bas.nonMatchAccum = bas.nonMatchAccum[n:]
			return n, nil
		}

		// If we had nothing in our buffer, we need to populate it.
		// FIXME: Handles "0, nil" properly?
		if len(bas.currBuf) == 0 {
			n, err := bas.src.Read(buf)
			bas.currBuf = append(bas.currBuf, buf[:n]...)
			bas.lastReaderErr = err
			fmt.Println("Did read:", string(bas.currBuf), err)
		}

		buffer := bas.currBuf
		for idx, b := range buffer {
			fmt.Println(idx, b)
			switch {
			case b == bas.searchString[len(bas.matchAccum)]:
				bas.matchAccum = append(bas.matchAccum, b)

				// if we now have our match, return it
				if len(bas.matchAccum) == len(bas.searchString) {
					copy(buf, bas.matchAccum)
					bas.matchAccum = bas.matchAccum[:0]
					bas.currBuf = bas.currBuf[idx:]
					return len(bas.searchString), nil
				}

			default:
				// If we had prior matches but this breaks the
				// match, we must put the prior match onto the
				// non-matching accumulator
				if len(bas.matchAccum) != 0 {
					bas.nonMatchAccum =
						append(bas.nonMatchAccum,
							bas.matchAccum...)
					bas.matchAccum = bas.matchAccum[:0]
				}

				// now accumulate this non matcher
				bas.nonMatchAccum = append(bas.nonMatchAccum, b)
			}

			bas.currBuf = bas.currBuf[1:]

			// if we've accumulated a full buffer now, return
			// it
			if len(bas.nonMatchAccum) == len(buf) {
				copy(buf, bas.nonMatchAccum)
				bas.nonMatchAccum = bas.nonMatchAccum[0:]
				return len(buf), nil
			}
		}

		// if we have something to return now, we can't have it be
		// too long
		if len(bas.nonMatchAccum) > len(buf) {
			panic("oops, bad math")
		}

		if len(bas.nonMatchAccum) > 0 {
			n := copy(buf, bas.nonMatchAccum)
			if n == len(bas.nonMatchAccum) {
				bas.nonMatchAccum = bas.nonMatchAccum[:0]
			} else {
				bas.nonMatchAccum = bas.nonMatchAccum[n:]
			}
			return n, nil
		}
	}
}
