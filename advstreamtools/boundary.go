package advstreamtools

import (
	"io"
)

// A generalReader is like a reader but it operates on slices of anything,
// instead of bytes in particular. When specialized to byte this is
// compatible with io.Reader.
type GeneralReader[In comparable] interface {
	Read([]In) (int, error)
}

type GeneralReadCloser[In comparable] interface {
	GeneralReader[In]
	io.Closer
}

type boundaryState byte

const (
	baNonMatching = boundaryState(iota)
	baMatching
	baYieldingMatch
	baDrainDueToError
	baErroring
)

func NewBoundary[In comparable](src GeneralReader[In], search []In) GeneralReadCloser[In] {
	return &boundaryAtomic[In]{
		r:      src,
		search: search,
	}
}

// boundaryAtomic tries to return a given search string as an atomic Read value.
type boundaryAtomic[In comparable] struct {
	r GeneralReader[In]

	// the sequence of values we are looking for.
	search []In
	// the buffer of things we are looking for.
	badMatchBuf []In
	currBuf     []In

	nonMatchingAccum []In
	matchingAccum    []In

	err error

	state boundaryState
}

// Close will close the underlying reader if it is an io.Closer.
func (ba *boundaryAtomic[In]) Close() error {
	closer, isCloser := ba.r.(io.Closer)
	if isCloser {
		return closer.Close()
	}
	return nil
}

// Read will read from the wrapped reader, trying its best to yield the
// search term as a single read result.
func (ba *boundaryAtomic[In]) Read(buf []In) (int, error) {
	// common actions the states perform

	// extend the buffer. If the return value is true, we should
	// immediately continue, otherwise the state can consider the
	// buffer extension "done".
	nextVal := func() (In, bool) {
		if len(ba.badMatchBuf) > 0 {
			return pop(&ba.badMatchBuf), false
		}
		if len(ba.currBuf) > 0 {
			return pop(&ba.currBuf), false
		}
		// if we're out of stuff to return, and there was an error
		// on the last read call, transition to the error state
		if ba.err != nil {
			var zero In
			ba.state = baDrainDueToError
			return zero, true
		}

		// There is no next value, so we need to try to extend the
		// buffer first.

		n, err := ba.r.Read(buf)
		if n == 0 {
			// readers are supposed to read through this case.
			ba.err = err
			var zero In
			return zero, true
		}
		// successfully read a buffer.
		ba.currBuf = append(ba.currBuf, buf[:n]...)
		if err != nil {
			// errors are supposed to still return what they
			// can, not cut off the values returned so far.
			// fortunately this is the same behavior for all
			// states using this.
			ba.err = err
		}
		return pop(&ba.currBuf), false
	}

StateLoop:
	for {
		switch ba.state {
		// The initial state because we start with a non-matching
		// state. In this state, the matchingAccum should always be
		// empty.
		case baNonMatching:
			// This accounts for the possibility of variable
			// sized buffers; just because the buffer used to
			// be larger than this doesn't mean that the next
			// call's buffer will be.
			if len(buf) <= len(ba.matchingAccum) {
				n := move(buf, &ba.matchingAccum)
				return n, nil
			}

			if len(ba.matchingAccum) > 0 {
				panic("in baNonMatching state with non-empty matching accumulation buffer")
			}

			nextVal, immediatelyContinue := nextVal()
			if immediatelyContinue {
				continue StateLoop
			}

			if nextVal == ba.search[0] {
				ba.matchingAccum = append(ba.matchingAccum, nextVal)
				ba.state = baMatching
				continue StateLoop
			}

			ba.nonMatchingAccum = append(ba.nonMatchingAccum, nextVal)
			continue StateLoop

		case baMatching:
			// FIXME: Push criterion func

			// We have a match. Push out the non-matching stuff
			// first if any, then push the match.
			if len(ba.matchingAccum) == len(ba.search) {
				ba.state = baYieldingMatch
				continue StateLoop
			}

			// see if the next thing matches the search
			// criterion
			nextValue, immediatelyContinue := nextVal()
			if immediatelyContinue {
				continue StateLoop
			}

			if nextValue == ba.search[len(ba.matchingAccum)] {
				ba.matchingAccum = append(ba.matchingAccum,
					nextValue)
				continue StateLoop
			}

			// otherwise, this DOESN'T match. We need to put
			// the first thing we thought might be a match into
			// the nonMatchingAccum, and roll the remainder
			// into the badMatchBuf for a retry on the
			// matching.
			//
			// For types like byte with a very limited number
			// of values in it, there are more efficient
			// algorithms that this. Those algorithms depend on
			// the number of values in the type being small and
			// break down for almost any other type, and this
			// is, honestly, still not bad, especially versus
			// the competition, which is trying to hold the
			// entire stream in RAM at once.
			ba.nonMatchingAccum = append(ba.nonMatchingAccum,
				ba.matchingAccum[0])
			ba.badMatchBuf = append(ba.badMatchBuf,
				ba.matchingAccum[1:]...)
			ba.badMatchBuf = append(ba.badMatchBuf, nextValue)
			ba.matchingAccum = ba.matchingAccum[:0]
			ba.state = baNonMatching
			continue StateLoop

		case baYieldingMatch:
			if len(ba.badMatchBuf) > 0 {
				panic("bad match buffer has contents in yield")
			}

			// If there was any non-matching stuff before this,
			// yield it.
			if len(ba.nonMatchingAccum) > 0 {
				n := move(buf, &ba.nonMatchingAccum)
				return n, nil
			}

			// then we clear the match buffer. If this buffer
			// is too small to handle the match, we return it
			// as best as we can.
			if len(ba.matchingAccum) > 0 {
				n := move(buf, &ba.matchingAccum)
				return n, nil
			}

			// now we've apparently cleared everything, resume
			// the matching process
			ba.state = baNonMatching
			continue StateLoop

		// we have received a read error, but we had more stuff to
		// yield first. return the stuff from before the error,
		// prior to returning the error.
		case baDrainDueToError:
			if len(ba.matchingAccum) > 0 {
				panic("in error drain, have stuff in the matching accumulation buffer")
			}
			if len(ba.currBuf) > 0 {
				panic("in error drain, still have stuff in buffer")
			}
			// is there anything in the non-matching accumulator?
			if len(ba.nonMatchingAccum) > 0 {
				n := move[In](buf, &ba.nonMatchingAccum)
				return n, nil
			}

			// if we reach here, we have finished writing out
			// the buffers prior to an error.
			ba.state = baErroring
			continue StateLoop

		// This state is terminal; once we start erroring, we never stop.
		case baErroring:
			return 0, ba.err
		}
	}
}
