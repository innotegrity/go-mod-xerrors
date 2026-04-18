package xerrors

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultCallerInfo(t *testing.T) {
	t.Parallel()

	caller := DefaultCallerInfo()
	if caller.File != UnknownCallerFile || caller.Line != -1 || caller.Func != UnknownCallerFunc {
		t.Fatalf("unexpected default caller: %+v", caller)
	}
}

func TestGetCallerInfo_RuntimeCallerFails(t *testing.T) {
	t.Parallel()

	// Skip past the top of the stack so runtime.Caller returns !ok.
	caller := getCallerInfo(1<<20, nil)
	if caller.File != UnknownCallerFile || caller.Line != -1 || caller.Func != UnknownCallerFunc {
		t.Fatalf("expected default caller when runtime.Caller fails, got %+v", caller)
	}
}

func TestGetCallerInfo_NoStripPrefixes(t *testing.T) {
	t.Parallel()

	caller := getCallerInfo(0, nil)
	if caller.File == UnknownCallerFile {
		t.Fatal("expected real file path")
	}

	if !strings.HasSuffix(caller.File, "caller_internal_test.go") {
		t.Fatalf("unexpected file %q", caller.File)
	}

	if caller.Line <= 0 {
		t.Fatalf("expected positive line, got %d", caller.Line)
	}

	if caller.Func == "" || caller.Func == UnknownCallerFunc {
		t.Fatalf("expected real func name, got %q", caller.Func)
	}
}

func TestGetCallerInfo_StripsMatchingPrefix(t *testing.T) {
	t.Parallel()

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	prefix := workDir + string(os.PathSeparator)

	caller := getCallerInfo(0, []string{prefix})
	if strings.HasPrefix(caller.File, prefix) {
		t.Fatalf("expected prefix stripped, got %q", caller.File)
	}

	if !strings.HasSuffix(caller.File, "caller_internal_test.go") {
		t.Fatalf("unexpected relative file %q", caller.File)
	}
}

func TestGetCallerInfo_SkipsNonMatchingThenStrips(t *testing.T) {
	t.Parallel()

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	prefix := workDir + string(os.PathSeparator)

	// First prefix does not match the real file path; second does.
	caller := getCallerInfo(0, []string{"/this/prefix/should/not/match/", prefix})
	if strings.HasPrefix(caller.File, prefix) {
		t.Fatalf("expected prefix stripped, got %q", caller.File)
	}
}
