package xerrors_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"go.innotegrity.dev/mod/xerrors"
)

func TestCaller_DefaultWhenNotCaptured(t *testing.T) {
	t.Parallel()

	errNoCaller := xerrors.New(1, "msg")

	caller := errNoCaller.Caller()

	badDefault := caller.File != xerrors.UnknownCallerFile ||
		caller.Line != -1 ||
		caller.Func != xerrors.UnknownCallerFunc
	if badDefault {
		t.Fatalf("expected default caller info, got %+v", caller)
	}
}

func TestCaller_FromWithCaller(t *testing.T) {
	t.Parallel()

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	errWithCaller := xerrors.New(1, "msg").
		WithStripFilePrefixes(workDir + string(os.PathSeparator)).
		WithCaller()

	caller := errWithCaller.Caller()

	unknown := caller.File == xerrors.UnknownCallerFile ||
		caller.Line <= 0 ||
		caller.Func == xerrors.UnknownCallerFunc
	if unknown {
		t.Fatalf("expected captured caller, got %+v", caller)
	}

	if strings.HasPrefix(caller.File, workDir) {
		t.Fatalf("expected stripped file prefix, got %q", caller.File)
	}
}

func TestCaller_FromContextWithCapture(t *testing.T) {
	t.Parallel()

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	prefix := workDir + string(os.PathSeparator)

	ctx := xerrors.ContextWithErrorOptions(context.Background(),
		nil,
		xerrors.WithCaptureCaller(),
		xerrors.WithSkipBias(0),
		xerrors.WithStripFilePrefixes(prefix),
	)

	errFromCtx := xerrors.New(2, "ctx").WithOptionsFromContext(ctx)

	caller := errFromCtx.Caller()

	unknown := caller.File == xerrors.UnknownCallerFile ||
		caller.Line <= 0 ||
		caller.Func == xerrors.UnknownCallerFunc
	if unknown {
		t.Fatalf("expected captured caller from context, got %+v", caller)
	}

	if strings.HasPrefix(caller.File, prefix) {
		t.Fatalf("expected stripped file prefix, got %q", caller.File)
	}
}

// newErrAtWithCallerCallSite is only for TestCaller_CaptureMatchesCallSite.
// Keep the marker on the WithCaller line.
//
//go:noinline
func newErrAtWithCallerCallSite(stripPrefix string) xerrors.Error {
	errObj := xerrors.New(1, "callsite")
	errObj = errObj.WithStripFilePrefixes(stripPrefix)

	return errObj.WithCaller() // test:callsite-WithCaller
}

// newErrAtWithOptionsFromContextCallSite is used only from TestCaller_CaptureMatchesCallSite.
//
//go:noinline
func newErrAtWithOptionsFromContextCallSite(ctx context.Context, stripPrefix string) xerrors.Error {
	errObj := xerrors.New(2, "ctx-site")
	errObj = errObj.WithStripFilePrefixes(stripPrefix)

	return errObj.WithOptionsFromContext(ctx) // test:callsite-WithOptionsFromContext
}

func lineNumberOfMarker(t *testing.T, filePath, marker string) int {
	t.Helper()

	data, err := os.ReadFile(filePath) //nolint:gosec // test reads this file via path from runtime.Caller
	if err != nil {
		t.Fatalf("read %s: %v", filePath, err)
	}

	lines := bytes.Split(data, []byte("\n"))
	for lineIdx, line := range lines {
		if bytes.Contains(line, []byte(marker)) {
			return lineIdx + 1
		}
	}

	t.Fatalf("marker %q not found in %s", marker, filePath)

	return 0
}

func TestCaller_CaptureMatchesCallSite(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	prefix := workDir + string(os.PathSeparator)

	t.Run("WithCaller", func(t *testing.T) {
		t.Parallel()

		wantLine := lineNumberOfMarker(t, thisFile, "test:callsite-WithCaller")
		errObj := newErrAtWithCallerCallSite(prefix)
		got := errObj.Caller()

		if got.Line != wantLine {
			t.Fatalf("Line: got %d, want %d (source line of WithCaller call)", got.Line, wantLine)
		}

		if filepath.Base(got.File) != "caller_test.go" {
			t.Fatalf("File: got base name %q, want caller_test.go", filepath.Base(got.File))
		}
	})

	t.Run("WithOptionsFromContext", func(t *testing.T) {
		t.Parallel()

		wantLine := lineNumberOfMarker(t, thisFile, "test:callsite-WithOptionsFromContext")

		ctx := xerrors.ContextWithErrorOptions(context.Background(),
			xerrors.WithCaptureCaller(),
			xerrors.WithSkipBias(0),
			xerrors.WithStripFilePrefixes(prefix),
		)

		errObj := newErrAtWithOptionsFromContextCallSite(ctx, prefix)
		got := errObj.Caller()

		if got.Line != wantLine {
			t.Fatalf("Line: got %d, want %d (source line of WithOptionsFromContext call)", got.Line, wantLine)
		}

		if filepath.Base(got.File) != "caller_test.go" {
			t.Fatalf("File: got base name %q, want caller_test.go", filepath.Base(got.File))
		}
	})
}
