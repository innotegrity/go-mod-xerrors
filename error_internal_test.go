package xerrors

import (
	"errors"
	"strings"
	"testing"
)

const stdMessage = "std message"

var errTestForcedEnvelopeMarshal = errors.New("forced envelope marshal failure")

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
