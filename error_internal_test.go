package xerrors

import "testing"

func TestJSONStdErrorError(t *testing.T) {
	e := &jsonStdError{Message: "std message"}
	if e.Error() != "std message" {
		t.Fatalf("expected message %q, got %q", "std message", e.Error())
	}
}
