package xerrors

import "testing"

const stdMessage = "std message"

func TestJSONStdError_Error(t *testing.T) {
	e := &jsonStdError{Message: stdMessage}
	if e.Error() != stdMessage {
		t.Fatalf("expected message %q, got %q", stdMessage, e.Error())
	}
}
