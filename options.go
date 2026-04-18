package xerrors

// ErrorOptionFn is used to configure error options for an [Error] object.
type ErrorOptionFn func(*ErrorOptions)

// ErrorOptions holds configurable options for [Error] objects.
type ErrorOptions struct {
	// unexported variables
	captureCaller     bool     // when true, records the call site when the error is constructed.
	skipBias          int      // the stack skip bias for caller capture.
	stripFilePrefixes []string // lists path prefixes to strip from the recorded file path.
}

// WithCaptureCaller enables caller capture for an [Error].
func WithCaptureCaller() ErrorOptionFn {
	return func(opts *ErrorOptions) {
		opts.captureCaller = true
	}
}

// WithSkipBias sets the stack skip bias for caller capture in an [Error] when caller capture is enabled.
func WithSkipBias(bias int) ErrorOptionFn {
	return func(opts *ErrorOptions) {
		opts.skipBias = bias
	}
}

// WithStripFilePrefixes sets the list of file path prefixes to strip from the caller file path in an [Error] when
// caller capture is enabled.
func WithStripFilePrefixes(prefixes ...string) ErrorOptionFn {
	return func(opts *ErrorOptions) {
		opts.stripFilePrefixes = prefixes
	}
}
