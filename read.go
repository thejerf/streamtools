package streamtools

import "io"

// This contains reader-centric operations.

// ReadUntil will read the given reader until the given byte is
// encountered, and put the values into the buffer. The int will return the
// number of bytes written into the buffer. The boolean will be true if the
// read completed, and false if there is more to read yet. error will be
// returned from the underlying reader, if any.
//
// The checked byte will be consumed, but nothing else will be.
func ReadUntil(r io.Reader, b byte, buf []byte) (int, bool, error) {
	mybuf := make([]byte, 1)
	var n int
	var err error
	for idx := range buf {
		n, err = r.Read(mybuf)
		if n == 1 {
			if mybuf[0] == b {
				return idx, true, nil
			}
			buf[idx] = mybuf[0]
		}
		if err != nil {
			return idx, true, err
		}
	}

	return len(buf), false, nil
}

// ReadUntilAny will read the given reader until one of the given bytes is
// encountered, and put the values into the buffer. The int will return the
// number of bytes written into the buffer. The boolean will be true if the
// read completed, and false if there is more to read yet. error will be
// returned from the underlying reader, if any.
//
// The checked byte will be consumed, but nothing else will be.
//
// If no ends bytes are passed, this degenerates into a call to io.ReadFull,
// except ErrUnexpectedEOF will be ignored because the caller is not trying
// to guarantee that there are enough bytes to fill the buffer.
func ReadUntilAny(r io.Reader, ends []byte, buf []byte) (int, bool, error) {
	if len(ends) == 0 {
		n, err := io.ReadFull(r, buf)
		if err == io.ErrUnexpectedEOF {
			return n, true, nil
		}
		if err != nil {
			return n, true, err
		}
		return n, false, nil
	}

	mybuf := make([]byte, 1)
	var n int
	var err error
	for idx := range buf {
		n, err = r.Read(mybuf)
		if n == 1 {
			// generally a for loop is faster than a map lookup
			// belowe about 10 elements, which is probably the
			// common case.
			for _, b := range ends {
				if mybuf[0] == b {
					return idx, true, nil
				}
			}
			buf[idx] = mybuf[0]
		}
		if err != nil {
			return idx, true, err
		}
	}

	return len(buf), false, nil
}
