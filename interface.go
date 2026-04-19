package xerrors

import (
	"context"
	"encoding/json"
)

// Error is the interface that all error objects must implement.
type Error interface {
	error
	json.Marshaler
	json.Unmarshaler

	// Attrs returns a shallow copy of the error's attributes, or nil if there are none.
	Attrs() map[string]any

	// Caller returns the caller information for the error, if available.
	Caller() CallerInfo

	// Code returns the code of the error.
	Code() int

	// Is returns true if the error is equal to the target error.
	Is(target error) bool

	// String returns the object's JSON representation as a string.
	String() string

	// Unwrap returns the wrapped error, if there is one.
	Unwrap() error

	// WithAttr adds an attribute to the error and returns itself.
	WithAttr(key string, value any) Error

	// WithAttrs adds attributes to the error and returns itself.
	WithAttrs(attrs map[string]any) Error

	// WithCaller captures caller information for the error and returns itself.
	WithCaller() Error

	// WithOptionsFromContext applies options from the context to the error and returns itself.
	WithOptionsFromContext(ctx context.Context) Error

	// WithSkipBias sets the stack skip bias for caller capture and returns itself.
	WithSkipBias(bias int) Error

	// WithStripFilePrefixes sets file path prefixes to strip from recorded caller paths and returns itself.
	WithStripFilePrefixes(prefixes ...string) Error
}
