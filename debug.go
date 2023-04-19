package streamtools

import (
	"fmt"
	"runtime/debug"
)

const (
	// ErrBufferTooSmall indicates that a Read call was made with a
	// buffer too small for some guarantee to be held.
	ErrBufferTooSmall = ErrorType(iota + 1)
)

// ErrorType is a constant that indicates the type of error that has
// occurred.
type ErrorType int

// DebugStackTrace indicates whether or not errors from this library will
// yield a stack trace. This is not thread-safe. It is intended to be set
// at start time, generally by test code. It should not be modified once
// streamtools are being used.
var DebugStackTrace bool

// StreamError is the type of error returned by everything in this
// package. It supports having an ErrorType that can be examined without
// having to pull apart the inner error, and an optional stack trace.
type StreamError struct {
	ErrorType    ErrorType
	WrappedError error
	StackTrace   string
}

// Error implements the error interface.
func (se StreamError) Error() string {
	errStr := se.WrappedError.Error()
	if se.StackTrace != "" {
		errStr += "\nstack trace:\n" + se.StackTrace
	}
	return errStr
}

// Unwrap implements the error unwrapping protocol.
func (se StreamError) Unwrap() []error {
	return []error{se.WrappedError}
}

func errorf(ty ErrorType, format string, args ...any) StreamError {
	st := ""
	if DebugStackTrace {
		st = string(debug.Stack())
	}

	return StreamError{
		ty,
		fmt.Errorf(format, args...),
		st,
	}
}
