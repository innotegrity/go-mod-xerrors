package xerrors

import "testing"

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
