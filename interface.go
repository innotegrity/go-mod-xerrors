package xerrors

import (
	"context"
	"encoding/json"
)

// XError is the interface that all error objects must implement.
//
//nolint:interfacebloat // interface includes  methods from json and error packages
type XError interface {
	error
	json.Marshaler
	json.Unmarshaler

	// Attrs returns a shallow copy of the error's attributes, or nil if there are none.
	Attrs() map[string]any

	// Caller returns the caller information for the error, if available.
	Caller() CallerInfo

	// Code returns the code of the error.
	Code() int

	// Error returns the actual error message.
	Error() string

	// Is returns true if the error is equal to the target error.
	Is(target error) bool

	// String returns the object's JSON representation as a string.
	String() string

	// Unwrap returns the wrapped error, if there is one.
	Unwrap() error

	// WithAttr adds an attribute to the error and returns itself.
	WithAttr(key string, value any) XError

	// WithAttrs adds attributes to the error and returns itself.
	WithAttrs(attrs map[string]any) XError

	// WithCaller captures caller information for the error and returns itself.
	WithCaller() XError

	// WithOptionsFromContext applies options from the context to the error and returns itself.
	WithOptionsFromContext(ctx context.Context) XError

	// WithSkipBias sets the stack skip bias for caller capture and returns itself.
	WithSkipBias(bias int) XError

	// WithStripFilePrefixes sets file path prefixes to strip from recorded caller paths and returns itself.
	WithStripFilePrefixes(prefixes ...string) XError
}
