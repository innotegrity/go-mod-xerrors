package xerrors

import (
	"errors"
	"strings"
	"testing"
)

var errTestForcedEnvelopeMarshal = errors.New("forced envelope marshal failure")

//nolint:paralleltest // mutates assertUnderlyingXError; must not run alongside other tests using the hook.
func TestNewAs_PanicsWhenAssertUnderlyingXErrorFails(t *testing.T) {
	old := assertUnderlyingXError
	assertUnderlyingXError = func(Error) (*XError, bool) { return nil, false }

	t.Cleanup(func() { assertUnderlyingXError = old })

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic")
		}

		s, ok := recovered.(string)
		if !ok || !strings.Contains(s, "New must return *XError") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	_ = NewAs(func(*XError) *XError { panic("ctor must not run") }, 1, "x")
}

//nolint:paralleltest // mutates assertUnderlyingXError; must not run alongside other tests using the hook.
func TestNewfAs_PanicsWhenAssertUnderlyingXErrorFails(t *testing.T) {
	old := assertUnderlyingXError
	assertUnderlyingXError = func(Error) (*XError, bool) { return nil, false }

	t.Cleanup(func() { assertUnderlyingXError = old })

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic")
		}

		s, ok := recovered.(string)
		if !ok || !strings.Contains(s, "Newf must return *XError") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	_ = NewfAs(func(*XError) *XError { panic("ctor must not run") }, 1, "fmt %d", 1)
}

//nolint:paralleltest // mutates assertUnderlyingXError; must not run alongside other tests using the hook.
func TestWrapAs_PanicsWhenAssertUnderlyingXErrorFails(t *testing.T) {
	old := assertUnderlyingXError
	assertUnderlyingXError = func(Error) (*XError, bool) { return nil, false }

	t.Cleanup(func() { assertUnderlyingXError = old })

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic")
		}

		s, ok := recovered.(string)
		if !ok || !strings.Contains(s, "Wrap must return *XError") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	_ = WrapAs(func(*XError) *XError { panic("ctor must not run") }, errTestForcedEnvelopeMarshal, 1, "x")
}

//nolint:paralleltest // mutates assertUnderlyingXError; must not run alongside other tests using the hook.
func TestWrapfAs_PanicsWhenAssertUnderlyingXErrorFails(t *testing.T) {
	old := assertUnderlyingXError
	assertUnderlyingXError = func(Error) (*XError, bool) { return nil, false }

	t.Cleanup(func() { assertUnderlyingXError = old })

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic")
		}

		s, ok := recovered.(string)
		if !ok || !strings.Contains(s, "Wrapf must return *XError") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	_ = WrapfAs(func(*XError) *XError { panic("ctor must not run") }, errTestForcedEnvelopeMarshal, 1, "fmt %d", 1)
}

const stdMessage = "std message"

func TestJSONStdError_Error(t *testing.T) {
	t.Parallel()

	e := &jsonStdError{Message: stdMessage}
	if e.Error() != stdMessage {
		t.Fatalf("expected message %q, got %q", stdMessage, e.Error())
	}
}

func TestNewf_NoArgsUsesLiteralMessage(t *testing.T) {
	t.Parallel()

	const want = "literal plaintext without fmt verbs"

	e := Newf(201, want)
	if e.Error() != want {
		t.Fatalf("got %q, want %q", e.Error(), want)
	}
}

//nolint:paralleltest // mutates jsonMarshalWrappedEnvelope; must not run alongside other tests using the hook.
func TestMarshalJSON_WrappedEnvelopeMarshalFailure(t *testing.T) {
	old := jsonMarshalWrappedEnvelope
	jsonMarshalWrappedEnvelope = func(any) ([]byte, error) {
		return nil, errTestForcedEnvelopeMarshal
	}

	t.Cleanup(func() { jsonMarshalWrappedEnvelope = old })

	inner := New(1, "inner")
	outer := Wrap(inner, 2, "outer")

	_, err := outer.MarshalJSON()
	if err == nil || !strings.Contains(err.Error(), "marshal extended error to JSON") {
		t.Fatalf("expected marshal error, got %v", err)
	}
}
