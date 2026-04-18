package xerrors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
)

const (
	// errFmtMarshalJSON is the format string for errors returned from [Error.MarshalJSON].
	errFmtMarshalJSON = "marshal extended error to JSON: %w"

	// errFmtUnmarshalJSON is the format string for errors returned from [Error.UnmarshalJSON].
	errFmtUnmarshalJSON = "unmarshal extended error from JSON: %w"

	// wrappedKindXError is the wrappedEnvelope.kind value for a nested [*Error] (recursive JSON shape).
	wrappedKindXError = "xerror"

	// wrappedKindStd is the wrappedEnvelope.kind value for a non-[Error] wrapped error ([jsonStdError]).
	wrappedKindStd = "std"
)

var (
	// errWrappedXErrorNeedsNested is returned when kind is [wrappedKindXError] but nested JSON is missing.
	errWrappedXErrorNeedsNested = errors.New(`xerrors: wrappedError kind "xerror" requires non-empty "nested"`)

	// errWrappedMissingKind is returned when [wrappedEnvelope.kind] is missing from wrappedError JSON.
	errWrappedMissingKind = errors.New(`xerrors: wrappedError must include "kind" ("xerror" or "std")`)
)

// jsonMarshalWrappedEnvelope serializes [wrappedEnvelope] for [Error.MarshalJSON]. Tests may replace it to cover the
// marshal error path (nested [Error.MarshalJSON] success always yields valid [wrappedEnvelope.Nested] JSON).
//
//nolint:gochecknoglobals // intentional indirection for tests; default is [json.Marshal].
var jsonMarshalWrappedEnvelope = json.Marshal

// Error is an extended error that holds an error code, message, caller information, attributes, and a wrapped error.
type Error struct {
	ErrorOptions

	// unexported variables
	attrs      map[string]any // error attributes
	caller     *CallerInfo    // information on where the error was generated
	code       int            // the error code
	message    string         // the error message
	wrappedErr error          // the wrapped error, if any
}

// jsonError is the JSON shape for [Error] used by [Error.MarshalJSON] and [Error.UnmarshalJSON].
type jsonError struct {
	// Attrs is a map of attributes associated with the error.
	Attrs map[string]any `json:"attrs,omitempty"`

	// Caller contains the information on where the error was generated.
	Caller *CallerInfo `json:"caller,omitempty"`

	// Code is the error code.
	Code int `json:"code"`

	// Message is the error message.
	Message string `json:"message"`

	// WrappedError holds JSON for [wrappedEnvelope] (tagged union).
	WrappedError json.RawMessage `json:"wrappedError,omitempty"`
}

// wrappedEnvelope is the tagged union written inside [jsonError.wrappedError].
type wrappedEnvelope struct {
	Kind string `json:"kind"`

	// Nested holds the recursive [*Error] JSON object when Kind is [wrappedKindXError].
	Nested json.RawMessage `json:"nested,omitempty"`

	// Message holds the text when Kind is [wrappedKindStd].
	Message string `json:"message,omitempty"`
}

// jsonStdError is a version of a standard Go error that is used to marshal the object to JSON.
type jsonStdError struct {
	// Message is the error message.
	Message string `json:"message"`
}

// Error returns the error message.
func (e *jsonStdError) Error() string {
	return e.Message
}

// New creates a new [XError] with the given code and message.
func New(code int, message string) XError {
	return &Error{
		code:    code,
		message: message,
	}
}

// Newf creates a new [XError] with the given code and formatted message.
//
// If args is empty, format is used as the literal message (no [fmt.Sprintf] processing), so strings that contain "%"
// are safe. If args is non-empty, format is passed to [fmt.Sprintf] with args.
func Newf(code int, format string, args ...any) XError {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	return &Error{
		code:    code,
		message: msg,
	}
}

// Wrap wraps the given error with the given code and message.
func Wrap(err error, code int, message string) XError {
	return &Error{
		code:       code,
		message:    message,
		wrappedErr: err,
	}
}

// Wrapf wraps the given error with the given code and formatted message.
//
// If args is empty, format is used as the literal message (no [fmt.Sprintf] processing), so strings that contain "%"
// are safe. If args is non-empty, format is passed to [fmt.Sprintf] with args.
func Wrapf(err error, code int, message string, args ...any) XError {
	msg := message
	if len(args) > 0 {
		msg = fmt.Sprintf(message, args...)
	}

	return &Error{
		code:       code,
		message:    msg,
		wrappedErr: err,
	}
}

// Attrs returns a shallow copy of attributes associated with the error, or nil if there are none.
//
// The copy does not alias the internal map, so mutating the returned map does not change the error; values that are
// reference types are still shared.
func (e *Error) Attrs() map[string]any {
	if e.attrs == nil {
		return nil
	}

	return maps.Clone(e.attrs)
}

// Caller returns the information on where the error was generated.
func (e *Error) Caller() CallerInfo {
	if e.caller == nil {
		caller := DefaultCallerInfo()

		return *caller
	}

	return *e.caller
}

// Code returns the error code.
func (e *Error) Code() int {
	return e.code
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.message
}

// Is returns true if the error matches the wrapped error in this object (if there is one) or false otherwise.
func (e *Error) Is(err error) bool {
	if e.wrappedErr == nil {
		return false
	}

	return errors.Is(err, e.wrappedErr)
}

// MarshalJSON marshals the error to JSON.
func (e *Error) MarshalJSON() ([]byte, error) {
	enc := jsonError{
		Caller:  e.caller,
		Code:    e.code,
		Message: e.message,
	}
	if e.wrappedErr != nil {
		var wrapped *Error

		var env wrappedEnvelope

		var err error

		if errors.As(e.wrappedErr, &wrapped) {
			env.Kind = wrappedKindXError
			env.Nested, err = json.Marshal(wrapped)
		} else {
			env.Kind = wrappedKindStd
			env.Message = e.wrappedErr.Error()
		}

		if err != nil {
			return nil, fmt.Errorf(errFmtMarshalJSON, err)
		}

		enc.WrappedError, err = jsonMarshalWrappedEnvelope(env)
		if err != nil {
			return nil, fmt.Errorf(errFmtMarshalJSON, err)
		}
	}

	if e.attrs != nil {
		enc.Attrs = make(map[string]any)
		maps.Copy(enc.Attrs, e.attrs)
	}

	data, err := json.Marshal(enc)
	if err != nil {
		return nil, fmt.Errorf(errFmtMarshalJSON, err)
	}

	return data, nil
}

// UnmarshalJSON unmarshals JSON produced by [Error.MarshalJSON] (or equivalent) into the receiver.
//
// The receiver is reset before fields are populated; [ErrorOptions] are not read from JSON and remain zero.
func (e *Error) UnmarshalJSON(data []byte) error {
	var payload jsonError

	err := json.Unmarshal(data, &payload)
	if err != nil {
		return fmt.Errorf(errFmtUnmarshalJSON, err)
	}

	*e = Error{}

	e.code = payload.Code
	e.message = payload.Message
	e.caller = payload.Caller

	if payload.Attrs != nil {
		e.attrs = make(map[string]any)
		maps.Copy(e.attrs, payload.Attrs)
	}

	if len(payload.WrappedError) == 0 {
		return nil
	}

	return e.decodeWrappedErrorJSON(payload.WrappedError)
}

// String returns the error (including the code, attributes, caller and wrapped error) represented as a JSON string.
func (e *Error) String() string {
	str, err := e.MarshalJSON()
	if err != nil {
		return "failed to marshal error to JSON: " + err.Error()
	}

	return string(str)
}

// Unwrap returns the wrapped error, if there is one.
func (e *Error) Unwrap() error {
	if e.wrappedErr == nil {
		return nil
	}

	return e.wrappedErr
}

// WithAttr adds an attribute to the error and returns itself.
func (e *Error) WithAttr(key string, value any) XError {
	if e.attrs == nil {
		e.attrs = make(map[string]any)
	}

	e.attrs[key] = value

	return e
}

// WithAttrs adds attributes to the error and returns itself.
func (e *Error) WithAttrs(attrs map[string]any) XError {
	if e.attrs == nil {
		e.attrs = make(map[string]any)
	}

	maps.Copy(e.attrs, attrs)

	return e
}

// WithCaller captures the caller information for the error and returns itself.
//
// As soon as this message is called, the caller information is captured and stored in the error object. Any further
// calls to this method will not capture new caller information.
//
// Be sure to call this method **after** the other With* methods that configure options for the [Error] but before
// any calls that return error details (such as [Error.Caller], etc.).
func (e *Error) WithCaller() XError {
	if e.caller == nil {
		e.caller = getCallerInfo(e.skipBias+1, e.stripFilePrefixes)
	}

	return e
}

// WithOptionsFromContext applies the options from the given context to the error and returns itself.
//
// If the context enables caller capture, the caller information is captured and stored in the error object
// immediately. Any further calls to this method or [Error.WithCaller] will not capture new caller information.
//
// If the context contains options that configure the error, they are applied first before any caller information
// is captured.
//
// Be sure to call this method **after** the other With* methods that configure options for the [Error] (in case they
// are not set in the context) but before any calls that return error details (such as [Error.Caller], etc.).
func (e *Error) WithOptionsFromContext(ctx context.Context) XError {
	opts := ErrorOptionsFromContext(ctx)
	for _, optsFn := range opts {
		if optsFn != nil {
			optsFn(&e.ErrorOptions)
		}
	}

	if e.captureCaller && e.caller == nil {
		e.caller = getCallerInfo(e.skipBias+1, e.stripFilePrefixes)
	}

	return e
}

// WithSkipBias sets the stack skip bias for caller capture in an [Error] when caller capture is enabled.
//
// The default skip bias is usually sufficient for most use cases.
func (e *Error) WithSkipBias(bias int) XError {
	e.skipBias = bias

	return e
}

// WithStripFilePrefixes sets the list of file path prefixes to strip from the caller file path in an [Error] when
// caller capture is enabled.
func (e *Error) WithStripFilePrefixes(prefixes ...string) XError {
	e.stripFilePrefixes = prefixes

	return e
}

// decodeWrappedErrorJSON decodes the wrapped error JSON into the error object.
func (e *Error) decodeWrappedErrorJSON(raw json.RawMessage) error {
	var env wrappedEnvelope

	err := json.Unmarshal(raw, &env)
	if err != nil {
		return fmt.Errorf(errFmtUnmarshalJSON, err)
	}

	if env.Kind == "" {
		return fmt.Errorf("extended error JSON: %w", errWrappedMissingKind)
	}

	switch env.Kind {
	case wrappedKindXError:
		if len(env.Nested) == 0 {
			return fmt.Errorf("extended error JSON: %w", errWrappedXErrorNeedsNested)
		}

		inner := new(Error)

		err = inner.UnmarshalJSON(env.Nested)
		if err != nil {
			return err
		}

		e.wrappedErr = inner

		return nil
	case wrappedKindStd:
		e.wrappedErr = &jsonStdError{Message: env.Message}

		return nil
	default:
		//nolint:err113 // kind is validated at JSON parse time and included for operator visibility
		return fmt.Errorf("extended error JSON: unknown wrappedError kind %q", env.Kind)
	}
}
